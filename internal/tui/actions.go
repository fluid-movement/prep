package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// errorFor is how long an error notice stays in the footer.
const errorFor = 6 * time.Second

type modalKind int

const (
	modalMenu modalKind = iota + 1
	modalCreate
	modalRename
	modalReparent
	modalCriteria
	modalDrop
	modalComplete
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
	items   []action
	checks  []bool
	picks   []string // reparent candidates; "" is top-level
	err     string
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
	in.PromptStyle = m.th.S.Title
	in.TextStyle = m.th.S.Body
	in.PlaceholderStyle = m.th.S.Subtle
	in.Cursor.Style = m.th.R.NewStyle().Foreground(m.th.C.Accent)
	in.Cursor.SetMode(cursor.CursorStatic)
	in.SetValue(value)
	in.CursorEnd()
	return in
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

func (m *Model) openMenu() tea.Cmd {
	m.modal = &modal{kind: modalMenu, id: m.selected(), items: m.actions()}
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
		action{key: "e", label: "Edit requirement", reason: edit, run: func() tea.Cmd { return m.editText(id, "requirement") }},
		action{key: "c", label: "Edit context", reason: edit, run: func() tea.Cmd { return m.editText(id, "context") }},
		action{key: "k", label: "Check criteria", reason: crit, run: m.openCriteria},
		action{key: "m", label: "Move to another parent", reason: edit, run: m.openReparent},
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

func (m *Model) openCreate() tea.Cmd {
	m.modal = &modal{kind: modalCreate, inputs: []textinput.Model{m.newInput("what should change", "")}}
	return m.modal.inputs[0].Focus()
}

func (m *Model) openRename() tea.Cmd {
	id := m.selected()
	m.modal = &modal{kind: modalRename, id: id, inputs: []textinput.Model{m.newInput("title", m.tree.Issues[id].Title)}}
	return m.modal.inputs[0].Focus()
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
func (m *Model) modalKey(k tea.KeyMsg) tea.Cmd {
	d := m.modal
	s := k.String()
	switch s {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		m.modal = nil
		return nil
	}
	switch d.kind {
	case modalMenu:
		// Letters run actions, so only arrows move.
		switch s {
		case "up":
			d.cursor = (d.cursor - 1 + len(d.items)) % len(d.items)
		case "down":
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
		case "up":
			d.cursor = clamp(d.cursor-1, 0, len(d.checks)-1)
		case "down":
			d.cursor = clamp(d.cursor+1, 0, len(d.checks)-1)
		case " ", "x":
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
	}

	// Forms: tab moves between fields, enter submits.
	fields := len(d.inputs)
	if d.kind == modalCreate {
		fields++ // the kind selector
	}
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
	if d.kind == modalCreate && d.focus == len(d.inputs) {
		switch s {
		case "left", "h":
			d.kindIdx = (d.kindIdx - 1 + len(kinds)) % len(kinds)
		case "right", "l", " ":
			d.kindIdx = (d.kindIdx + 1) % len(kinds)
		}
		return nil
	}
	var cmd tea.Cmd
	d.inputs[d.focus], cmd = d.inputs[d.focus].Update(k)
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
		title, kind, parent := d.inputs[0].Value(), kinds[d.kindIdx], m.scopeTop()
		return m.write("created "+strings.TrimSpace(title), func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanNew(domain.NewIssueInput{Title: title, Kind: kind, Parent: parent}, now)
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

// editText opens the requirement or the context in the user's editor. Text
// a rejected write left behind is offered again instead of the stored one.
func (m *Model) editText(id, field string) tea.Cmd {
	if id == "" || m.tree.Issues[id] == nil {
		return nil
	}
	if m.tree.State(id).Terminal() {
		return m.flashErr(fmt.Errorf("resolved issues are history"))
	}
	i := m.tree.Issues[id]
	before := i.Body
	if field == "context" {
		before = i.Context
	}
	text := before
	if p, ok := m.pending[id+"|"+field]; ok {
		text = p
	}
	f, err := os.CreateTemp("", "prep-"+shortID(id)+"-"+field+"-*.md")
	if err != nil {
		return m.flashErr(err)
	}
	_, err = f.WriteString(text + "\n")
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return m.flashErr(err)
	}
	path := f.Name()
	return m.runEditor(path, func(err error) tea.Msg { return editedMsg{id: id, field: field, path: path, before: before, err: err} })
}

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
	menuKeys     = []ui.Key{{Keys: "↑↓", Desc: "move"}, {Keys: "enter or key", Desc: "run"}, {Keys: "esc", Desc: "close"}}
	formKeys     = []ui.Key{{Keys: "enter", Desc: "save"}, {Keys: "tab", Desc: "next field"}, {Keys: "esc", Desc: "cancel"}}
	pickKeys     = []ui.Key{{Keys: "type", Desc: "filter"}, {Keys: "↑↓", Desc: "move"}, {Keys: "enter", Desc: "move here"}, {Keys: "esc", Desc: "cancel"}}
	criteriaKeys = []ui.Key{{Keys: "↑↓", Desc: "move"}, {Keys: "space", Desc: "toggle"}, {Keys: "enter", Desc: "save"}, {Keys: "esc", Desc: "cancel"}}
)

func (m *Model) modalKeys() []ui.Key {
	switch m.modal.kind {
	case modalMenu:
		return menuKeys
	case modalReparent:
		return pickKeys
	case modalCriteria:
		return criteriaKeys
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
	switch d.kind {
	case modalMenu:
		title = "Actions · " + title
		if d.id == "" {
			title = "Actions"
		}
		for k, it := range d.items {
			body = append(body, ui.MenuRow(m.th, it.key, it.label, it.reason, it.reason == "", k == d.cursor, inner))
		}
	case modalCreate:
		title = "New issue"
		if p := m.scopeTop(); p != "" {
			title += " under " + shortID(p) + " " + m.tree.Issues[p].Title
		}
		kindLine := ""
		for k, kd := range kinds {
			label := " " + string(kd) + " "
			if k == d.kindIdx {
				kindLine += m.th.R.NewStyle().Foreground(m.th.C.Accent).Background(m.th.C.Selection).Bold(true).Render(label)
			} else {
				kindLine += m.th.S.Muted.Render(label)
			}
		}
		body = append(body, ui.Field(m.th, "Title", d.inputs[0].View(), d.focus == 0), "", ui.Field(m.th, "Kind  ←/→", kindLine, d.focus == 1), "",
			m.th.S.Subtle.Render("Write the requirement next with e."))
	case modalRename:
		title = "Rename · " + title
		body = append(body, ui.Field(m.th, "Title", d.inputs[0].View(), true))
	case modalDrop:
		title = "Drop · " + title
		body = append(body, ui.Field(m.th, "Reason", d.inputs[0].View(), true))
	case modalComplete:
		title = "Complete · " + title
		body = append(body, m.th.S.Muted.Render("Documentation decision: the knowledge entries this changed, or why none."), "",
			ui.Field(m.th, "Knowledge entries", d.inputs[0].View(), d.focus == 0), "",
			ui.Field(m.th, "No impact because", d.inputs[1].View(), d.focus == 1))
	case modalCriteria:
		title = "Criteria · " + title
		for k, c := range m.tree.Issues[d.id].Criteria {
			mark := "[ ]"
			if k < len(d.checks) && d.checks[k] {
				mark = "[x]"
			}
			line := ui.Fit(fmt.Sprintf("%s %d. %s", mark, k+1, c.Text), inner-2)
			if k == d.cursor {
				line = m.th.R.NewStyle().Background(m.th.C.Selection).Width(inner).Render(m.th.R.NewStyle().Foreground(m.th.C.Accent).Render("▌ ") + line)
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
		for k := start; k < len(d.picks) && k < start+rows; k++ {
			id := d.picks[k]
			sel := k == d.cursor
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
	return ui.Center(m.th, ui.Modal(m.th, title, strings.Join(body, "\n"), w), width, height)
}
