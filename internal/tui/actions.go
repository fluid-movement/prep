package tui

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// errorFor is how long an error notice stays in the footer.
const errorFor = 6 * time.Second

type modalKind int

const (
	modalMenu modalKind = iota + 1
	modalCreate
	modalRename
	modalTags
	modalReparent
	modalCriteria
	modalDrop
	modalComplete
	modalText     // requirement or context, inline
	modalViewEdit // a saved view's name and query
	modalHelp     // the keymap of the current screen
)

// action is one entry of the action menu.
type action struct {
	key, label string
	reason     string // why it is unavailable now; empty when available
	run        func() tea.Cmd
}

// modal is the open dialog. Forms keep their inputs here; a rejected write
// leaves the dialog open with the error.
type modal struct {
	kind    modalKind
	id      string
	inputs  []textinput.Model
	focus   int // index into inputs; len(inputs) is the kind selector
	kindIdx int
	cursor  int
	step    int // modalCreate: 0 title, 1 kind, 2 requirement
	items   []action
	checks  []bool
	picks   []string // reparent candidates; "" is top-level
	heading string   // a menu's title in place of "Actions"
	field   string   // modalText: requirement or context; modalViewEdit: the view's old name
	before  string   // modalText: the stored text, to tell whether anything changed
	area    textarea.Model
	err     string
	marks   []entryMark // clickable entries of the last render
	// scroll is the first pick shown; scrolled means the wheel set it, so
	// it no longer follows the cursor until a key.
	scroll   int
	scrolled bool
}

// entryMark is a clickable entry in a dialog's body: its ID, the body line
// it starts on, its column and its size in cells.
type entryMark struct {
	id      string
	line, x int
	w, h    int
}

var kinds = []domain.Kind{domain.KindCode, domain.KindManual, domain.KindResearch, domain.KindDecision}

type wroteMsg struct {
	id, what   string
	err        error
	fromModal  bool
	pendingKey string // editor text to keep on failure
	text       string
}

type editedMsg struct {
	id, field, path, before string
	err                     error
}

func (m *Model) newInput(placeholder, value string) textinput.Model {
	in := textinput.New()
	in.Prompt = "› "
	in.Placeholder = placeholder
	in.SetStyles(inputStyles(m.th))
	in.SetValue(value)
	in.CursorEnd()
	return in
}

// inputStyles styles a text input from the theme, with a static cursor so
// no blink timer redraws the screen.
func inputStyles(th *theme.Theme) textinput.Styles {
	state := textinput.StyleState{Prompt: th.S.Title, Text: th.S.Body, Placeholder: th.S.Subtle}
	return textinput.Styles{Focused: state, Blurred: state, Cursor: textinput.CursorStyle{Color: th.C.Accent}}
}

// write runs a planned change through the injected writer in the
// background and reports the result as a wroteMsg.
func (m *Model) write(what string, plan func(*domain.Tree) (*domain.Change, error), fromModal bool) tea.Cmd {
	if m.opts.Write == nil {
		return m.flashErr(fmt.Errorf("this TUI is read-only"))
	}
	w := m.opts.Write
	return func() tea.Msg {
		id, err := w(plan)
		return wroteMsg{id: id, what: what, err: err, fromModal: fromModal}
	}
}

func (m *Model) now() time.Time { return time.Now() }

func (m *Model) handleWrote(msg wroteMsg) tea.Cmd {
	if msg.err != nil {
		if msg.pendingKey != "" {
			m.pending[msg.pendingKey] = msg.text
		}
		if msg.fromModal && m.modal != nil {
			m.modal.err = msg.err.Error()
			return nil
		}
		return m.flashErr(msg.err)
	}
	if msg.pendingKey != "" {
		delete(m.pending, msg.pendingKey)
	}
	if msg.fromModal {
		m.modal = nil
	}
	m.pendingSelect = msg.id
	return tea.Batch(m.flash(msg.what), m.reload())
}

// --- action menu ---

// openMenu lists the actions that apply to the selected issue now; the
// direct keys explain why an action does not apply.
func (m *Model) openMenu() tea.Cmd {
	var items []action
	for _, a := range m.actions() {
		if a.reason == "" {
			items = append(items, a)
		}
	}
	m.modal = &modal{kind: modalMenu, id: m.selected(), items: items}
	return nil
}

// actions lists what the user can do with the selected issue. Transitions
// are checked with the domain's gates for the human actor; placeholder
// arguments keep argument gates (reason, documentation) out of the way.
func (m *Model) actions() []action {
	id := m.selected()
	items := []action{{key: "n", label: "New issue", run: m.openCreate}}
	if id == "" || m.tree == nil {
		return items
	}
	i := m.tree.Issues[id]
	edit := ""
	if m.tree.State(id).Terminal() {
		edit = "resolved issues are history"
	}
	crit := edit
	if crit == "" && len(i.Criteria) == 0 {
		crit = "no criteria yet"
	}
	complete := m.gate(id, domain.OpComplete)
	if i.Kind == domain.KindCode && !m.tree.IsParent(id) {
		complete = "code issues are completed by an agent with commit evidence"
	}
	return append(items,
		action{key: "t", label: "Edit title", reason: edit, run: m.openRename},
		action{key: "e", label: "Edit requirement", reason: edit, run: func() tea.Cmd { return m.openText(id, "requirement") }},
		action{key: "c", label: "Edit context", reason: edit, run: func() tea.Cmd { return m.openText(id, "context") }},
		action{key: "v", label: "Tick off criteria (verify)", reason: crit, run: m.openCriteria},
		action{key: "m", label: "Move to another parent", reason: edit, run: m.openReparent},
		action{key: "i", label: "Set priority", run: m.openPriority},
		action{key: "g", label: "Edit tags", run: m.openTags},
		action{key: "d", label: "Define", reason: m.gate(id, domain.OpDefine), run: m.transition(id, domain.OpDefine, "defined")},
		action{key: "r", label: "Mark ready", reason: m.gate(id, domain.OpReady), run: m.transition(id, domain.OpReady, "marked ready")},
		action{key: "a", label: "Acknowledge change", reason: m.gate(id, domain.OpAck), run: m.transition(id, domain.OpAck, "acknowledged")},
		action{key: "x", label: "Drop", reason: m.gate(id, domain.OpDrop), run: m.openDrop},
		action{key: "f", label: "Complete", reason: complete, run: m.openComplete},
	)
}

func (m *Model) gate(id string, op domain.Op) string {
	in := domain.Input{Actor: m.opts.Actor, Reason: "-", NoImpact: "-"}
	if unmet, _ := m.tree.Gates(id, op, &in); len(unmet) > 0 {
		return unmet[0].Message
	}
	return ""
}

func (m *Model) transition(id string, op domain.Op, done string) func() tea.Cmd {
	return func() tea.Cmd {
		m.modal = nil
		actor := m.opts.Actor
		now := m.now()
		return m.write(done+" "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
			return t.Plan(id, op, domain.Input{Actor: actor, Now: now})
		}, false)
	}
}

// --- dialogs ---

// openCreate starts the new-issue wizard: title, kind, requirement.
func (m *Model) openCreate() tea.Cmd {
	m.modal = &modal{kind: modalCreate, field: "requirement", inputs: []textinput.Model{m.newInput("what should change", "")}}
	m.modal.area = m.newArea("what and why (optional; ctrl+s creates)", "", m.dialogInner())
	m.modal.area.SetHeight(textHeight - 2)
	return m.modal.inputs[0].Focus()
}

// createStep moves the wizard to a step and focuses its field.
func (m *Model) createStep(step int) tea.Cmd {
	d := m.modal
	d.step, d.err = step, ""
	d.inputs[0].Blur()
	d.area.Blur()
	switch step {
	case 0:
		return d.inputs[0].Focus()
	case 2:
		return d.area.Focus()
	}
	return nil
}

// createKey runs the wizard: enter continues from the title; a kind's
// letter (c m r d) picks it and continues, as do arrows and enter; the
// requirement takes text until ctrl+s creates the issue.
func (m *Model) createKey(s string, k tea.KeyPressMsg) tea.Cmd {
	d := m.modal
	switch d.step {
	case 0:
		if s == "enter" || s == "tab" {
			if strings.TrimSpace(d.inputs[0].Value()) == "" {
				d.err = "a title is needed"
				return nil
			}
			return m.createStep(1)
		}
		var cmd tea.Cmd
		d.inputs[0], cmd = d.inputs[0].Update(k)
		return cmd
	case 1:
		for n, kd := range kinds {
			if s == string(kd)[:1] {
				d.kindIdx = n
				return m.createStep(2)
			}
		}
		switch s {
		case "up", "left", "k", "h":
			d.kindIdx = (d.kindIdx - 1 + len(kinds)) % len(kinds)
		case "down", "right", "j", "l":
			d.kindIdx = (d.kindIdx + 1) % len(kinds)
		case "enter", "tab":
			return m.createStep(2)
		}
		return nil
	}
	switch s {
	case "ctrl+s":
		return m.submit()
	case "ctrl+e":
		return m.handOff()
	}
	var cmd tea.Cmd
	d.area, cmd = d.area.Update(k)
	return cmd
}

// openPriority offers the four levels as a menu, the current one selected.
// Priority is metadata, so it can change on resolved issues too.
func (m *Model) openPriority() tea.Cmd {
	id := m.selected()
	cur := m.tree.Issues[id].Priority.Effective()
	d := &modal{kind: modalMenu, id: id, heading: "Priority"}
	for k, p := range domain.Priorities {
		label := strings.ToUpper(string(p[:1])) + string(p[1:])
		if p == cur {
			label += " (current)"
			d.cursor = k
		}
		d.items = append(d.items, action{key: string(p[:1]), label: label, run: func() tea.Cmd {
			level := string(p)
			actor, now := m.opts.Actor, m.now()
			return m.write("priority "+level+" "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
				return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Priority: &level})
			}, false)
		}})
	}
	m.modal = d
	return nil
}

func (m *Model) openRename() tea.Cmd {
	id := m.selected()
	m.modal = &modal{kind: modalRename, id: id, inputs: []textinput.Model{m.newInput("title", m.tree.Issues[id].Title)}}
	return m.modal.inputs[0].Focus()
}

// openTags edits the selected issue's tags. Tags are metadata, so they
// change on resolved issues too.
func (m *Model) openTags() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	m.modal = &modal{kind: modalTags, id: id, inputs: []textinput.Model{m.newInput("tags, space separated", strings.Join(m.tree.Issues[id].Tags, " "))}}
	m.modal.inputs[0].CursorEnd()
	m.tagSuggestions()
	return m.modal.inputs[0].Focus()
}

// tagSuggestions offers the tags in use for the word being typed in the
// tags dialog, as the filter bar does for values: prefix matches first,
// none for an empty word, a tag typed in full or tags already listed.
func (m *Model) tagSuggestions() {
	d := m.modal
	d.picks, d.cursor = nil, 0
	before := []rune(d.inputs[0].Value())
	word := string(before[:min(d.inputs[0].Position(), len(before))])
	word = word[strings.LastIndexAny(word, " ,")+1:]
	if word == "" {
		return
	}
	listed := strings.FieldsFunc(d.inputs[0].Value(), func(r rune) bool { return r == ' ' || r == ',' })
	var tags []string
	for _, f := range m.tree.QueryFlags() {
		if f.Name != "tag" {
			continue
		}
		for _, v := range f.Values {
			if v.Value == word {
				return
			}
			if !slices.Contains(listed, v.Value) {
				tags = append(tags, v.Value)
			}
		}
	}
	d.picks = rank(tags, func(s string) string { return s }, word)
	if len(d.picks) > maxSuggestions {
		d.picks = d.picks[:maxSuggestions]
	}
}

// insertTag replaces the word being typed with the highlighted tag.
func (m *Model) insertTag() {
	d := m.modal
	if d.cursor >= len(d.picks) {
		return
	}
	runes := []rune(d.inputs[0].Value())
	pos := min(d.inputs[0].Position(), len(runes))
	start := strings.LastIndexAny(string(runes[:pos]), " ,") + 1
	start = len([]rune(string(runes[:pos])[:start]))
	tag := []rune(d.picks[d.cursor] + " ")
	out := append(append(append([]rune{}, runes[:start]...), tag...), runes[pos:]...)
	d.inputs[0].SetValue(string(out))
	d.inputs[0].SetCursor(start + len(tag))
	m.tagSuggestions()
}

func (m *Model) openDrop() tea.Cmd {
	m.modal = &modal{kind: modalDrop, id: m.selected(), inputs: []textinput.Model{m.newInput("why it is not needed", "")}}
	return m.modal.inputs[0].Focus()
}

func (m *Model) openComplete() tea.Cmd {
	m.modal = &modal{kind: modalComplete, id: m.selected(), inputs: []textinput.Model{
		m.newInput("/components/cli.md /overview.md", ""),
		m.newInput("or: why the knowledge base is unaffected", ""),
	}}
	return m.modal.inputs[0].Focus()
}

func (m *Model) openCriteria() tea.Cmd {
	id := m.selected()
	var checks []bool
	for _, c := range m.tree.Issues[id].Criteria {
		checks = append(checks, c.Checked)
	}
	m.modal = &modal{kind: modalCriteria, id: id, checks: checks}
	return nil
}

func (m *Model) openReparent() tea.Cmd {
	id := m.selected()
	m.modal = &modal{kind: modalReparent, id: id, inputs: []textinput.Model{m.newInput("type to filter", "")}}
	m.modal.picks = m.parentCandidates(id, "")
	return m.modal.inputs[0].Focus()
}

// parentCandidates lists possible parents: top-level first, then every
// unresolved issue that is not the issue itself or below it.
func (m *Model) parentCandidates(id, filter string) []string {
	below := map[string]bool{id: true}
	for _, d := range m.tree.Descendants(id) {
		below[d] = true
	}
	filter = strings.ToLower(strings.TrimSpace(filter))
	out := []string{""}
	for _, c := range m.tree.TreeOrder(m.tree.IDs()) {
		if below[c] || m.tree.State(c).Terminal() {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(m.tree.Issues[c].Title+" "+c), filter) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// modalKey handles keys while a dialog is open.
func (m *Model) modalKey(k tea.KeyPressMsg) tea.Cmd {
	d := m.modal
	s := k.String()
	switch s {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		if d.kind == modalCreate && d.step > 0 {
			return m.createStep(d.step - 1)
		}
		m.modal = nil
		return nil
	}
	switch d.kind {
	case modalHelp:
		m.modal = nil // any key closes it
		return nil
	case modalMenu:
		// Letters run actions; arrows and j/k move (no menu uses j or k).
		switch s {
		case "up", "k":
			d.cursor = (d.cursor - 1 + len(d.items)) % len(d.items)
		case "down", "j":
			d.cursor = (d.cursor + 1) % len(d.items)
		case "enter":
			return m.runAction(d.items[d.cursor])
		case "q":
			m.modal = nil
		default:
			for _, it := range d.items {
				if it.key == s {
					return m.runAction(it)
				}
			}
		}
		return nil
	case modalCriteria:
		switch s {
		case "up", "k":
			d.cursor = clamp(d.cursor-1, 0, len(d.checks)-1)
		case "down", "j":
			d.cursor = clamp(d.cursor+1, 0, len(d.checks)-1)
		case "space", "x":
			if len(d.checks) > 0 {
				d.checks[d.cursor] = !d.checks[d.cursor]
			}
		case "enter":
			return m.saveCriteria()
		}
		return nil
	case modalReparent:
		switch s {
		case "up", "ctrl+p":
			d.cursor = clamp(d.cursor-1, 0, len(d.picks)-1)
			return nil
		case "down", "ctrl+n":
			d.cursor = clamp(d.cursor+1, 0, len(d.picks)-1)
			return nil
		case "enter":
			if len(d.picks) == 0 {
				return nil
			}
			id, parent := d.id, d.picks[d.cursor]
			actor, now := m.opts.Actor, m.now()
			what := "moved to top level"
			if parent != "" {
				what = "moved under " + shortID(parent)
			}
			return m.write(what, func(t *domain.Tree) (*domain.Change, error) {
				return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Parent: &parent})
			}, true)
		}
		var cmd tea.Cmd
		d.inputs[0], cmd = d.inputs[0].Update(k)
		d.picks = m.parentCandidates(d.id, d.inputs[0].Value())
		d.cursor = clamp(d.cursor, 0, len(d.picks)-1)
		return cmd
	case modalTags:
		switch s {
		case "up":
			d.cursor = clamp(d.cursor-1, 0, max(0, len(d.picks)-1))
			return nil
		case "down":
			d.cursor = clamp(d.cursor+1, 0, max(0, len(d.picks)-1))
			return nil
		case "tab":
			m.insertTag()
			return nil
		case "enter":
			if len(d.picks) > 0 {
				m.insertTag()
				return nil
			}
			return m.submit()
		}
		var cmd tea.Cmd
		d.inputs[0], cmd = d.inputs[0].Update(k)
		d.err = ""
		m.tagSuggestions()
		return cmd
	case modalCreate:
		return m.createKey(s, k)
	case modalText:
		switch s {
		case "ctrl+s":
			return m.saveText()
		case "ctrl+e":
			return m.handOff()
		}
		var cmd tea.Cmd
		d.area, cmd = d.area.Update(k)
		return cmd
	}

	// Forms: tab moves between fields, enter submits.
	fields := len(d.inputs)
	switch s {
	case "tab", "shift+tab":
		step := 1
		if s == "shift+tab" {
			step = -1
		}
		d.focus = (d.focus + step + fields) % fields
		for k := range d.inputs {
			if k == d.focus {
				d.inputs[k].Focus()
			} else {
				d.inputs[k].Blur()
			}
		}
		return nil
	case "enter":
		return m.submit()
	}
	var cmd tea.Cmd
	d.inputs[d.focus], cmd = d.inputs[d.focus].Update(k)
	return cmd
}

// paste sends pasted text to the field that takes text now: the filter bar
// or the open dialog's input or editor. Bubble Tea reports a paste as its
// own message, not as keys.
func (m *Model) paste(p tea.PasteMsg) tea.Cmd {
	var cmd tea.Cmd
	d := m.modal
	switch {
	case d == nil:
		if m.filtering {
			m.input, cmd = m.input.Update(p)
			m.filterChanged()
		}
	case d.kind == modalText || d.kind == modalCreate && d.step == 2:
		d.area, cmd = d.area.Update(p)
	case d.kind == modalCreate && d.step == 0:
		d.inputs[0], cmd = d.inputs[0].Update(p)
	case d.kind == modalReparent:
		d.inputs[0], cmd = d.inputs[0].Update(p)
		d.picks = m.parentCandidates(d.id, d.inputs[0].Value())
		d.cursor = clamp(d.cursor, 0, len(d.picks)-1)
	case d.kind == modalTags:
		d.inputs[0], cmd = d.inputs[0].Update(p)
		m.tagSuggestions()
	case d.kind == modalRename || d.kind == modalDrop || d.kind == modalComplete || d.kind == modalViewEdit:
		d.inputs[d.focus], cmd = d.inputs[d.focus].Update(p)
	}
	return cmd
}

func (m *Model) runAction(a action) tea.Cmd {
	if a.reason != "" {
		return m.flashErr(fmt.Errorf("%s: %s", strings.ToLower(a.label), a.reason))
	}
	m.modal = nil
	return a.run()
}

func (m *Model) submit() tea.Cmd {
	d := m.modal
	actor, now := m.opts.Actor, m.now()
	switch d.kind {
	case modalCreate:
		title, kind, parent, body := d.inputs[0].Value(), kinds[d.kindIdx], m.scopeTop(), strings.TrimSpace(d.area.Value())
		return m.write("created "+strings.TrimSpace(title), func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanNew(domain.NewIssueInput{Title: title, Kind: kind, Parent: parent, Body: body}, now)
		}, true)
	case modalViewEdit:
		return m.saveViewEdit()
	case modalTags:
		id := d.id
		tags := strings.FieldsFunc(d.inputs[0].Value(), func(r rune) bool { return r == ' ' || r == ',' })
		return m.write("tags of "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Tags: &tags})
		}, true)
	case modalRename:
		id, title := d.id, d.inputs[0].Value()
		return m.write("renamed "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Title: &title})
		}, true)
	case modalDrop:
		id, reason := d.id, d.inputs[0].Value()
		return m.write("dropped "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
			return t.Plan(id, domain.OpDrop, domain.Input{Actor: actor, Now: now, Reason: reason})
		}, true)
	case modalComplete:
		id := d.id
		docs := strings.FieldsFunc(d.inputs[0].Value(), func(r rune) bool { return r == ',' || r == ' ' })
		noImpact := d.inputs[1].Value()
		return m.write("completed "+shortID(id), func(t *domain.Tree) (*domain.Change, error) {
			return t.Plan(id, domain.OpComplete, domain.Input{Actor: actor, Now: now, Docs: docs, NoImpact: noImpact})
		}, true)
	}
	return nil
}

func (m *Model) saveCriteria() tea.Cmd {
	d := m.modal
	var ops []domain.AcceptanceOp
	for k, c := range m.tree.Issues[d.id].Criteria {
		if k < len(d.checks) && d.checks[k] != c.Checked {
			op := "uncheck"
			if d.checks[k] {
				op = "check"
			}
			ops = append(ops, domain.AcceptanceOp{Op: op, Index: k + 1})
		}
	}
	if len(ops) == 0 {
		m.modal = nil
		return nil
	}
	id, actor, now := d.id, m.opts.Actor, m.now()
	return m.write("criteria saved", func(t *domain.Tree) (*domain.Change, error) {
		return t.PlanRecord(id, domain.OpCriterion, domain.RecordInput{Actor: actor, Now: now, Acceptance: ops})
	}, true)
}

// --- editor ---

// execEditor suspends the TUI and runs $VISUAL, $EDITOR or vi on path.
func execEditor(path string, done func(error) tea.Msg) tea.Cmd {
	ed := os.Getenv("VISUAL")
	if ed == "" {
		ed = os.Getenv("EDITOR")
	}
	if ed == "" {
		ed = "vi"
	}
	parts := strings.Fields(ed)
	return tea.ExecProcess(exec.Command(parts[0], append(parts[1:], path)...), done)
}

func (m *Model) handleEdited(msg editedMsg) tea.Cmd {
	b, rerr := os.ReadFile(msg.path)
	os.Remove(msg.path)
	if msg.err != nil {
		return m.flashErr(fmt.Errorf("editor: %v", msg.err))
	}
	if rerr != nil {
		return m.flashErr(rerr)
	}
	text := strings.TrimSpace(string(b))
	key := msg.id + "|" + msg.field
	if text == strings.TrimSpace(msg.before) {
		delete(m.pending, key)
		return m.flash("no changes")
	}
	id, actor, now := msg.id, m.opts.Actor, m.now()
	var cmd tea.Cmd
	if msg.field == "context" {
		cmd = m.write("context saved", func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanRecord(id, domain.OpContext, domain.RecordInput{Actor: actor, Now: now, Text: text})
		}, false)
	} else {
		cmd = m.write("requirement saved", func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Body: &text})
		}, false)
	}
	return func() tea.Msg {
		w := cmd().(wroteMsg)
		w.pendingKey, w.text = key, text
		return w
	}
}

// --- rendering ---

var (
	menuKeys     = []ui.Key{{Keys: "↑↓ j k", Desc: "move"}, {Keys: "enter or key", Desc: "run"}, {Keys: "esc", Desc: "close"}}
	formKeys     = []ui.Key{{Keys: "enter", Desc: "save"}, {Keys: "tab", Desc: "next field"}, {Keys: "esc", Desc: "cancel"}}
	pickKeys     = []ui.Key{{Keys: "type", Desc: "filter"}, {Keys: "↑↓", Desc: "move"}, {Keys: "enter", Desc: "move here"}, {Keys: "esc", Desc: "cancel"}}
	criteriaKeys = []ui.Key{{Keys: "↑↓ j k", Desc: "move"}, {Keys: "space", Desc: "toggle"}, {Keys: "enter", Desc: "save"}, {Keys: "esc", Desc: "cancel"}}
	textKeys     = []ui.Key{{Keys: "ctrl+s", Desc: "save"}, {Keys: "ctrl+e", Desc: "$EDITOR"}, {Keys: "esc", Desc: "cancel"}}
	createKeys   = []ui.Key{{Keys: "enter", Desc: "continue"}, {Keys: "esc", Desc: "back"}}
	createLast   = []ui.Key{{Keys: "ctrl+s", Desc: "create"}, {Keys: "ctrl+e", Desc: "$EDITOR"}, {Keys: "esc", Desc: "back"}}
)

func (m *Model) modalKeys() []ui.Key {
	switch m.modal.kind {
	case modalMenu:
		return menuKeys
	case modalReparent:
		return pickKeys
	case modalCriteria:
		return criteriaKeys
	case modalText:
		return textKeys
	case modalCreate:
		if m.modal.step == 2 {
			return createLast
		}
		return createKeys
	case modalHelp:
		return []ui.Key{{Keys: "any key", Desc: "close"}}
	case modalTags:
		return []ui.Key{{Keys: "↑/↓", Desc: "choose"}, {Keys: "enter tab", Desc: "pick, or save"}, {Keys: "esc", Desc: "cancel"}}
	}
	return formKeys
}

func (m *Model) modalView(width, height int) string {
	d := m.modal
	w := min(width-4, 90)
	inner := w - 4
	title := ""
	if d.id != "" && m.tree.Issues[d.id] != nil {
		title = shortID(d.id) + " " + m.tree.Issues[d.id].Title
	}
	var body []string
	d.marks = nil
	// next is the body line the next element starts on; mark makes h lines
	// from there clickable, w cells from column x.
	next := func() int {
		line := 0
		for _, b := range body {
			line += strings.Count(b, "\n") + 1
		}
		return line
	}
	mark := func(id string, line, x, w, h int) {
		d.marks = append(d.marks, entryMark{id: id, line: line, x: x, w: w, h: h})
	}
	switch d.kind {
	case modalMenu:
		heading := "Actions"
		if d.heading != "" {
			heading = d.heading
		}
		title = heading + " · " + title
		if d.id == "" {
			title = heading
		}
		for k, it := range d.items {
			mark(fmt.Sprintf("menu:%d", k), next(), 0, inner, 1)
			body = append(body, ui.MenuRow(m.th, it.key, it.label, it.reason, it.reason == "", k == d.cursor, inner))
		}
	case modalCreate:
		title = fmt.Sprintf("New issue · %d/3", d.step+1)
		if p := m.scopeTop(); p != "" {
			title += " · under " + shortID(p) + " " + m.tree.Issues[p].Title
		}
		done := func(label, value string) string {
			return m.th.S.Subtle.Render(label+": ") + m.th.S.Muted.Render(value)
		}
		switch d.step {
		case 0:
			mark("field:0", next(), 0, inner, 2)
			body = append(body, ui.Field(m.th, "Title", d.inputs[0].View(), true), "", m.th.S.Subtle.Render("enter continues"))
		case 1:
			var opts []string
			body = append(body, done("Title", d.inputs[0].Value()), "")
			x := 0
			for k, kd := range kinds {
				w := len(kd) + 2                                   // " code ", one space between options
				mark(fmt.Sprintf("kind:%d", k), next()+1, x, w, 1) // under the field label
				x += w + 1
				letter, rest := string(kd)[:1], string(kd)[1:]
				label := lipgloss.NewStyle().Foreground(m.th.C.Accent).Bold(true).Render(letter) + m.th.S.Body.Render(rest)
				if k == d.kindIdx {
					label = lipgloss.NewStyle().Background(m.th.C.Selection).Render(" " + label + " ")
				} else {
					label = " " + label + " "
				}
				opts = append(opts, label)
			}
			body = append(body, ui.Field(m.th, "Kind", strings.Join(opts, " "), true), "",
				m.th.S.Subtle.Render("its letter picks it · arrows and enter work too · esc goes back"))
		default:
			body = append(body, done("Title", d.inputs[0].Value()), done("Kind", string(kinds[d.kindIdx])), "",
				ui.Field(m.th, "Requirement", d.area.View(), true))
		}
	case modalText:
		title = map[string]string{"requirement": "Requirement", "context": "Context"}[d.field] + " · " + title
		body = append(body, d.area.View())
	case modalHelp:
		title = "Keys"
		body = m.helpLines(inner)
		if room := height - 4; len(body) > room {
			// Two columns, split at the group boundary nearest the middle.
			cut := len(body) / 2
			for k := cut; k < len(body); k++ {
				if body[k] == "" {
					cut = k
					break
				}
			}
			colW := (inner - 2) / 2
			left, right := m.helpColumn(body[:cut], colW), m.helpColumn(body[min(cut+1, len(body)):], colW)
			body = strings.Split(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right), "\n")
		}
	case modalViewEdit:
		title = "Add view"
		if d.field != "" {
			title = "View · " + d.field
		}
		mark("field:0", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Name", d.inputs[0].View(), d.focus == 0), "")
		mark("field:1", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Query", d.inputs[1].View(), d.focus == 1), "",
			m.th.S.Subtle.Render("prep list flags; empty lists all issues."))
	case modalTags:
		title = "Tags · " + title
		mark("field:0", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Tags", d.inputs[0].View(), true))
		for k, t := range d.picks {
			n := 0
			for _, id := range m.tree.IDs() {
				if slices.Contains(m.tree.Issues[id].Tags, t) {
					n++
				}
			}
			body = append(body, ui.Suggestion(m.th, t, n, "", k == d.cursor, inner))
		}
		body = append(body, "", m.th.S.Subtle.Render("Saving replaces the tags; empty removes them all."))
	case modalRename:
		title = "Rename · " + title
		mark("field:0", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Title", d.inputs[0].View(), true))
	case modalDrop:
		title = "Drop · " + title
		mark("field:0", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Reason", d.inputs[0].View(), true))
	case modalComplete:
		title = "Complete · " + title
		body = append(body, m.th.S.Muted.Render("Documentation decision: the knowledge entries this changed, or why none."), "")
		mark("field:0", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "Knowledge entries", d.inputs[0].View(), d.focus == 0), "")
		mark("field:1", next(), 0, inner, 2)
		body = append(body, ui.Field(m.th, "No impact because", d.inputs[1].View(), d.focus == 1))
	case modalCriteria:
		title = "Criteria · " + title
		for k, c := range m.tree.Issues[d.id].Criteria {
			box := "[ ]"
			if k < len(d.checks) && d.checks[k] {
				box = "[x]"
			}
			mark(fmt.Sprintf("crit:%d", k), next(), 0, inner, 1)
			line := ui.Fit(fmt.Sprintf("%s %d. %s", box, k+1, c.Text), inner-2)
			if k == d.cursor {
				line = lipgloss.NewStyle().Background(m.th.C.Selection).Width(inner).Render(lipgloss.NewStyle().Foreground(m.th.C.Accent).Render("▌ ") + line)
			} else {
				line = "  " + m.th.S.Body.Render(line)
			}
			body = append(body, line)
		}
	case modalReparent:
		title = "Move · " + title
		body = append(body, d.inputs[0].View(), "")
		rows := max(3, height-10)
		start := clamp(d.cursor-rows/2, 0, max(0, len(d.picks)-rows))
		if d.scrolled {
			start = clamp(d.scroll, 0, max(0, len(d.picks)-rows))
		}
		d.scroll = start
		for k := start; k < len(d.picks) && k < start+rows; k++ {
			id := d.picks[k]
			sel := k == d.cursor
			mark(fmt.Sprintf("pick:%d", k), next(), 0, inner, 1)
			if id == "" {
				body = append(body, ui.MenuRow(m.th, "", "(top level)", "no parent", true, sel, inner))
				continue
			}
			body = append(body, ui.LinkRow(m.th, "", m.tree.State(id), shortID(id), m.tree.Issues[id].Title, sel, inner))
		}
	}
	if d.err != "" {
		body = append(body, "", ui.Error(m.th, strings.ReplaceAll(d.err, "\n", " ")))
	}
	return ui.Modal(m.th, title, strings.Join(body, "\n"), w)
}
