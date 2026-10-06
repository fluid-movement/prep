package tui

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/ui"
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
	st := lipgloss.NewStyle()
	state := textarea.StyleState{Base: st, CursorLine: st, Text: m.th.S.Body, Placeholder: m.th.S.Subtle, EndOfBuffer: m.th.S.Subtle, Prompt: m.th.S.Subtle}
	s := textarea.Styles{Focused: state, Blurred: state, Cursor: textarea.CursorStyle{Color: m.th.C.Accent}}
	s.Focused.Prompt = st.Foreground(m.th.C.Accent)
	a.SetStyles(s) // a static cursor: no blink timer redrawing the screen
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

// openEditMenu offers what can be edited on the selected issue: e r
// requirement, e c context, e t title.
func (m *Model) openEditMenu() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	if m.tree.State(id).Terminal() {
		return m.flashErr(fmt.Errorf("resolved issues are history; only their priority and tags change"))
	}
	m.modal = &modal{kind: modalMenu, id: id, heading: "Edit", items: []action{
		{key: "r", label: "Requirement", run: func() tea.Cmd { return m.openText(id, "requirement") }},
		{key: "c", label: "Context", run: func() tea.Cmd { return m.openText(id, "context") }},
		{key: "t", label: "Title", run: m.openRename},
	}}
	return nil
}

// linkKeys number the entries of link menus: digits, then letters that
// do not move (j, k, h and l navigate every menu).
const linkKeys = "123456789bcdfgmnprstuvwxyz"

// openLinks lists the selected issue's links as a numbered menu, in the
// order the detail shows them: o 3 goes to the third.
func (m *Model) openLinks() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	rels := m.relations(id)
	if len(rels) == 0 {
		return m.flash("no linked issues: no parent, children or dependencies")
	}
	lead := map[string]string{"path": "↑ ", "child": "├ ", "depends on": "← needs ", "blocks": "→ unblocks ", "knowledge": "≡ "}
	// Children first after the path, as the detail draws them.
	order := func(l string) int {
		return map[string]int{"path": 0, "child": 1, "depends on": 2, "blocks": 3, "knowledge": 4}[l]
	}
	sorted := append([]relation(nil), rels...)
	slices.SortStableFunc(sorted, func(a, b relation) int { return order(a.label) - order(b.label) })
	keys := linkKeys
	d := &modal{kind: modalMenu, id: id, heading: "Go to"}
	for n, r := range sorted {
		if n >= len(keys) {
			break
		}
		target := r.id
		if r.label == "knowledge" {
			e := m.tree.Knowledge[target]
			d.items = append(d.items, action{key: keys[n : n+1], label: lead[r.label] + e.Title + "  " + e.Path, run: func() tea.Cmd {
				m.back = append(m.back, id)
				return m.openKnowledge(target)
			}})
			continue
		}
		d.items = append(d.items, action{key: keys[n : n+1], label: lead[r.label] + shortID(target) + " " + m.tree.Issues[target].Title,
			run: func() tea.Cmd { return m.jump(target, true) }})
	}
	m.modal = d
	return nil
}

// --- keymap ---

// openHelp shows the full keymap of the current screen, grouped, and on the
// issue screens the action menu's sequences for the selected issue.
func (m *Model) openHelp() tea.Cmd {
	m.modal = &modal{kind: modalHelp}
	return nil
}

func (m *Model) helpLines(inner int) []string {
	bs := m.bindings()
	if m.screen == screenIssues && m.tree != nil && m.selected() != "" {
		for _, a := range m.actions() {
			desc := a.label
			if a.reason != "" {
				desc += " (not now: " + a.reason + ")"
			}
			bs = append(bs, bind("Action menu", "a "+a.key, desc, false))
		}
	}
	var out []string
	group := ""
	for _, x := range bs {
		if x.group != group {
			if group != "" {
				out = append(out, "")
			}
			group = x.group
			out = append(out, m.th.S.Heading.Render(group))
		}
		key := m.th.S.Key.Render(fmt.Sprintf("%-15s", x.key.Keys))
		out = append(out, ui.Fit("  "+key+m.th.S.KeyDesc.Render(x.key.Desc), inner))
	}
	return out
}

// helpColumn re-fits keymap lines to a column width.
func (m *Model) helpColumn(lines []string, width int) string {
	out := make([]string, len(lines))
	for k, l := range lines {
		out[k] = ui.Fit(l, width)
	}
	return strings.Join(out, "\n")
}

// --- settings ---

// settingsRows is the number of selectable rows: the commit mode, then the views.
// settingsRows counts the settings rows: the commit mode, the views, and
// the user's mouse choice last.
func (m *Model) settingsRows() int { return 2 + len(domain.ViewNames(m.tree.Project.Config)) }

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
// commit mode or the mouse, enter edits a view, n adds one, d d deletes it,
// and m starts moving it.
func (m *Model) settingsKey(s string) tea.Cmd {
	if m.moving {
		return m.moveKey(s)
	}
	rows := m.settingsRows()
	m.setIdx = clamp(m.setIdx, 0, rows-1)
	cfg := m.settingsConfig()
	view := m.setIdx - 1 // index into cfg.ViewOrder; -1 is the commit mode
	onMouse := m.setIdx == rows-1
	if onMouse {
		view = -2
	}
	if s != "d" {
		m.confirm = ""
	}
	switch s {
	case "up", "k":
		m.setIdx = clamp(m.setIdx-1, 0, rows-1)
	case "down", "j":
		m.setIdx = clamp(m.setIdx+1, 0, rows-1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.setIdx = clamp(int(s[0]-'0'), 0, rows-2)
	case "space", "enter":
		if onMouse {
			return m.toggleMouse()
		}
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
	case "m":
		if view >= 0 {
			m.moving = true
			m.notice, m.noticeTone = "moving "+cfg.ViewOrder[view]+": j/k or arrows, enter when done", 0
		}
	}
	return nil
}

// moveKey moves the selected view while in move mode: j/k or arrows move
// it one place (one write each), enter or esc ends the mode.
func (m *Model) moveKey(s string) tea.Cmd {
	cfg := m.settingsConfig()
	view := m.setIdx - 1
	to := view
	switch s {
	case "up", "k":
		to--
	case "down", "j":
		to++
	case "enter", "esc", "m":
		m.moving, m.notice = false, ""
		return nil
	default:
		return nil
	}
	if view < 0 || to < 0 || to >= len(cfg.ViewOrder) {
		return nil
	}
	cfg.ViewOrder[view], cfg.ViewOrder[to] = cfg.ViewOrder[to], cfg.ViewOrder[view]
	m.setIdx = to + 1
	return m.saveSettings("moved view "+cfg.ViewOrder[to], cfg, false)
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
