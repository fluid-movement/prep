package tui

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/fluid-movement/prep/internal/domain"
)

// --- multi-line text ---

// textHeight is the editor's height in dialogs, in lines.
const textHeight = 10

func (m *Model) newArea(placeholder, value string, width int) textarea.Model {
	a := textarea.New()
	a.Placeholder = placeholder
	a.ShowLineNumbers = false
	a.Prompt = "│ "
	a.CharLimit = 0
	a.MaxHeight = 0
	a.SetWidth(width)
	a.SetHeight(textHeight)
	st := m.th.R.NewStyle()
	for _, s := range []*textarea.Style{&a.FocusedStyle, &a.BlurredStyle} {
		s.Base = st
		s.CursorLine = st
		s.Text = m.th.S.Body
		s.Placeholder = m.th.S.Subtle
		s.EndOfBuffer = m.th.S.Subtle
		s.Prompt = m.th.S.Subtle
	}
	a.FocusedStyle.Prompt = st.Foreground(m.th.C.Accent)
	a.Cursor.Style = st.Foreground(m.th.C.Accent)
	a.Cursor.SetMode(cursor.CursorStatic)
	a.SetValue(value)
	return a
}

// openText edits the requirement or the context inline. Text a rejected
// write left behind is offered again instead of the stored one.
func (m *Model) openText(id, field string) tea.Cmd {
	if id == "" || m.tree == nil || m.tree.Issues[id] == nil {
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
	d := &modal{kind: modalText, id: id, field: field, before: before}
	d.area = m.newArea(field, text, m.dialogInner())
	m.modal = d
	return d.area.Focus()
}

// dialogInner is the text width inside a dialog.
func (m *Model) dialogInner() int { return min(m.w-4, 90) - 4 }

// saveText writes the inline editor's text; a rejected write keeps the
// dialog open with the error and remembers the text.
func (m *Model) saveText() tea.Cmd {
	d := m.modal
	text := strings.TrimSpace(d.area.Value())
	if text == strings.TrimSpace(d.before) {
		m.modal = nil
		delete(m.pending, d.id+"|"+d.field)
		return m.flash("no changes")
	}
	cmd := m.planText(d.id, d.field, text, true)
	key := d.id + "|" + d.field
	return func() tea.Msg {
		w := cmd().(wroteMsg)
		w.pendingKey, w.text = key, text
		return w
	}
}

// planText writes requirement or context text for an issue.
func (m *Model) planText(id, field, text string, fromModal bool) tea.Cmd {
	actor, now := m.opts.Actor, m.now()
	if field == "context" {
		return m.write("context saved", func(t *domain.Tree) (*domain.Change, error) {
			return t.PlanRecord(id, domain.OpContext, domain.RecordInput{Actor: actor, Now: now, Text: text})
		}, fromModal)
	}
	return m.write("requirement saved", func(t *domain.Tree) (*domain.Change, error) {
		return t.PlanEdit(id, domain.EditInput{Actor: actor, Now: now, Body: &text})
	}, fromModal)
}

// areaEditedMsg returns the text of a dialog's editor from $EDITOR.
type areaEditedMsg struct {
	path string
	err  error
}

// handOff writes the dialog's text to a file and opens it in $EDITOR; the
// result comes back into the same field.
func (m *Model) handOff() tea.Cmd {
	d := m.modal
	f, err := os.CreateTemp("", "prep-"+shortID(d.id)+"-"+d.field+"-*.md")
	if err != nil {
		return m.flashErr(err)
	}
	_, err = f.WriteString(d.area.Value() + "\n")
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return m.flashErr(err)
	}
	path := f.Name()
	return m.runEditor(path, func(err error) tea.Msg { return areaEditedMsg{path: path, err: err} })
}

func (m *Model) handleAreaEdited(msg areaEditedMsg) tea.Cmd {
	b, rerr := os.ReadFile(msg.path)
	os.Remove(msg.path)
	switch {
	case msg.err != nil:
		return m.flashErr(fmt.Errorf("editor: %v", msg.err))
	case rerr != nil:
		return m.flashErr(rerr)
	case m.modal == nil || (m.modal.kind != modalText && m.modal.kind != modalCreate):
		return nil
	}
	m.modal.area.SetValue(strings.TrimRight(string(b), "\n"))
	return nil
}

// directCriteria opens the checklist dialog, or says why it cannot.
func (m *Model) directCriteria() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	switch {
	case m.tree.State(id).Terminal():
		return m.flashErr(fmt.Errorf("resolved issues are history"))
	case len(m.tree.Issues[id].Criteria) == 0:
		return m.flashErr(fmt.Errorf("%s has no criteria yet", shortID(id)))
	}
	return m.openCriteria()
}

// --- next transition ---

// nextOp is the transition > runs for an issue: acknowledging a stale
// issue, else the first forward transition in lifecycle order. Release and
// drop are ways out, not forward.
func (m *Model) nextOp(id string) (domain.Op, bool) {
	t := m.tree
	if t.Stale(id) && t.Applicable(id, domain.OpAck) {
		return domain.OpAck, true
	}
	for _, op := range domain.Ops {
		if op == domain.OpAck || op == domain.OpRelease || op == domain.OpDrop {
			continue
		}
		if t.Applicable(id, op) {
			return op, true
		}
	}
	return "", false
}

// nextTransition runs on >: the first press names the transition and asks
// for a second; complete opens its dialog. A transition that does not apply
// says why.
func (m *Model) nextTransition() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	op, ok := m.nextOp(id)
	if !ok {
		return m.flashErr(fmt.Errorf("%s has no next step: it is %s", shortID(id), m.tree.State(id)))
	}
	if op == domain.OpComplete {
		if i := m.tree.Issues[id]; i.Kind == domain.KindCode && !m.tree.IsParent(id) {
			return m.flashErr(fmt.Errorf("code issues are completed by an agent with commit evidence"))
		}
		m.confirm = ""
		return m.openComplete()
	}
	if reason := m.gate(id, op); reason != "" {
		m.confirm = ""
		return m.flashErr(fmt.Errorf("%s: %s", op, reason))
	}
	if m.confirm != id+"|"+string(op) {
		m.confirm = id + "|" + string(op)
		m.notice, m.noticeTone = fmt.Sprintf("%s %s? press > again", op, shortID(id)), 0
		return nil
	}
	m.confirm = ""
	m.notice = ""
	return m.transition(id, op, pastTense(op))()
}

func pastTense(op domain.Op) string {
	switch op {
	case domain.OpReady:
		return "marked ready"
	case domain.OpAck:
		return "acknowledged"
	case domain.OpClaim:
		return "claimed"
	}
	return string(op) + "d"
}

// --- settings ---

// settingsRows is the number of selectable rows: the commit mode, then the views.
func (m *Model) settingsRows() int { return 1 + len(domain.ViewNames(m.tree.Project.Config)) }

// settingsConfig copies the project configuration with its view order spelled out.
func (m *Model) settingsConfig() domain.Config {
	c := m.tree.Project.Config
	cp := domain.Config{CommitMode: c.CommitMode, Views: map[string]string{}, ViewOrder: domain.ViewNames(c)}
	for _, n := range cp.ViewOrder {
		cp.Views[n] = c.Views[n]
	}
	return cp
}

func (m *Model) saveSettings(what string, cfg domain.Config, fromModal bool) tea.Cmd {
	return m.write(what, func(t *domain.Tree) (*domain.Change, error) { return t.PlanConfig(cfg) }, fromModal)
}

// settingsKey handles the settings screen: arrows or digits select a row
// (digit n is view n, the tab it shows as), space or enter toggles the
// commit mode, enter edits a view, n adds one, d d deletes it, and
// shift+arrows or K and J move it.
func (m *Model) settingsKey(s string) tea.Cmd {
	rows := m.settingsRows()
	m.setIdx = clamp(m.setIdx, 0, rows-1)
	cfg := m.settingsConfig()
	view := m.setIdx - 1 // index into cfg.ViewOrder; -1 is the commit mode
	if s != "d" {
		m.confirm = ""
	}
	switch s {
	case "up", "k":
		m.setIdx = clamp(m.setIdx-1, 0, rows-1)
	case "down", "j":
		m.setIdx = clamp(m.setIdx+1, 0, rows-1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.setIdx = clamp(int(s[0]-'0'), 0, rows-1)
	case " ", "enter":
		if view < 0 {
			cfg.CommitMode = map[string]string{domain.CommitOff: domain.CommitAll, domain.CommitAll: domain.CommitOff}[cfg.CommitMode]
			return m.saveSettings("commit mode "+cfg.CommitMode, cfg, false)
		}
		if s == "enter" {
			name := cfg.ViewOrder[view]
			return m.openViewEdit(name, cfg.Views[name])
		}
	case "n":
		return m.openViewEdit("", "")
	case "d":
		if view < 0 {
			return nil
		}
		name := cfg.ViewOrder[view]
		if m.confirm != "delete|"+name {
			m.confirm = "delete|" + name
			m.notice, m.noticeTone = fmt.Sprintf("delete view %q? press d again", name), 0
			return nil
		}
		m.confirm, m.notice = "", ""
		cfg.ViewOrder = slices.Delete(cfg.ViewOrder, view, view+1)
		delete(cfg.Views, name)
		return m.saveSettings("deleted view "+name, cfg, false)
	case "shift+up", "K", "shift+down", "J":
		to := view - 1
		if s == "shift+down" || s == "J" {
			to = view + 1
		}
		if view < 0 || to < 0 || to >= len(cfg.ViewOrder) {
			return nil
		}
		cfg.ViewOrder[view], cfg.ViewOrder[to] = cfg.ViewOrder[to], cfg.ViewOrder[view]
		m.setIdx = to + 1
		return m.saveSettings("moved view "+cfg.ViewOrder[to], cfg, false)
	}
	return nil
}

// openViewEdit edits a saved view's name and query, or adds one when name is empty.
func (m *Model) openViewEdit(name, query string) tea.Cmd {
	m.modal = &modal{kind: modalViewEdit, field: name, inputs: []textinput.Model{m.newInput("name", name), m.newInput("query flags, e.g. --state ready --priority high", query)}}
	return m.modal.inputs[0].Focus()
}

func (m *Model) saveViewEdit() tea.Cmd {
	d := m.modal
	old := d.field
	name, query := strings.TrimSpace(d.inputs[0].Value()), strings.TrimSpace(d.inputs[1].Value())
	cfg := m.settingsConfig()
	if old == "" {
		if _, taken := cfg.Views[name]; taken {
			d.err = fmt.Sprintf("a view named %q exists", name)
			return nil
		}
		cfg.ViewOrder = append(cfg.ViewOrder, name)
	} else {
		if name != old {
			if _, taken := cfg.Views[name]; taken {
				d.err = fmt.Sprintf("a view named %q exists", name)
				return nil
			}
			delete(cfg.Views, old)
		}
		cfg.ViewOrder[slices.Index(cfg.ViewOrder, old)] = name
	}
	cfg.Views[name] = query
	what := "saved view " + name
	if old == "" {
		m.setIdx = len(cfg.ViewOrder)
		what = "added view " + name
	}
	return m.saveSettings(what, cfg, true)
}
