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
	"github.com/fluid-movement/prep/internal/palette"
	"github.com/fluid-movement/prep/internal/tui/theme"
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
// requirement, e c context, e t title, e g tags.
func (m *Model) openEditMenu() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	tags := action{key: "g", label: "Tags", run: m.openTags}
	if m.tree.State(id).Terminal() {
		// Resolved issues are history; only their tags (and priority, p)
		// still change.
		m.modal = &modal{kind: modalMenu, id: id, heading: "Edit", items: []action{tags}}
		return nil
	}
	m.modal = &modal{kind: modalMenu, id: id, heading: "Edit", items: []action{
		{key: "r", label: "Requirement", run: func() tea.Cmd { return m.openText(id, "requirement") }},
		{key: "c", label: "Context", run: func() tea.Cmd { return m.openText(id, "context") }},
		{key: "t", label: "Title", run: m.openRename},
		tags,
	}}
	return nil
}

// linkKeys number the entries of link menus: digits, then letters that
// do not move (j, k, h and l navigate every menu).
const linkKeys = "123456789bcdfgmnprstuvwxyz"

// openLinks starts link mode in the detail pane: the link lines show
// their keys and the first one is highlighted where it is drawn.
func (m *Model) openLinks() tea.Cmd {
	id := m.selected()
	if id == "" || m.tree == nil {
		return nil
	}
	if len(linkTargets(m.relations(id))) == 0 {
		return m.flash("no linked issues: no parent, children or dependencies")
	}
	m.linkMode, m.linkSel, m.focus = true, 0, focusDetail
	return nil
}

// linkKey handles keys in link mode: j/k or arrows move, enter follows,
// a link's key follows it at once, esc or o leave. Other keys leave link
// mode and do what they do in the detail.
func (m *Model) linkKey(s string) (tea.Cmd, bool) {
	id := m.selected()
	rels := m.relations(id)
	targets := linkTargets(rels)
	switch s {
	case "down", "j": // wraps around, like the menus
		m.linkSel = (m.linkSel + 1) % max(1, len(targets))
		return nil, true
	case "up", "k":
		m.linkSel = (m.linkSel + len(targets) - 1) % max(1, len(targets))
		return nil, true
	case "esc", "o":
		m.linkMode = false
		return nil, true
	case "enter":
		return m.followLink(id, rels, targets, m.linkSel), true
	}
	if n := strings.Index(linkKeys, s); len(s) == 1 && n >= 0 && n < len(targets) {
		return m.followLink(id, rels, targets, n), true
	}
	m.linkMode = false
	return nil, false
}

// followLink goes to the n-th link target and leaves link mode.
func (m *Model) followLink(id string, rels []relation, targets []int, n int) tea.Cmd {
	m.linkMode = false
	if n < 0 || n >= len(targets) {
		return nil
	}
	return m.followRel(id, rels[targets[n]])
}

// followRel opens a relation: a linked issue, or a knowledge entry.
func (m *Model) followRel(id string, r relation) tea.Cmd {
	if r.label == "knowledge" {
		m.pushBack()
		return m.openKnowledge(r.id)
	}
	return m.jump(r.id, true)
}

// linkTargets lists the relations link mode moves over, as the detail
// draws them: the parent (the breadcrumb's last step), children, needs,
// unblocks, knowledge. Ancestors above the parent are reached from it.
func linkTargets(rels []relation) []int {
	var out []int
	parent := -1
	for k, r := range rels {
		if r.label == "path" {
			parent = k
		}
	}
	if parent >= 0 {
		out = append(out, parent)
	}
	for _, label := range []string{"child", "depends on", "blocks", "knowledge"} {
		for k, r := range rels {
			if r.label == label {
				out = append(out, k)
			}
		}
	}
	return out
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

// settingsRows counts the settings rows: the theme, the views, and the
// user's mouse choice last.
func (m *Model) settingsRows() int { return 2 + len(domain.ViewNames(m.tree.Project.Config)) }

// settingsConfig copies the project configuration with its view order spelled out.
func (m *Model) settingsConfig() domain.Config {
	c := m.tree.Project.Config
	cp := domain.Config{Views: map[string]string{}, ViewOrder: domain.ViewNames(c), Theme: c.Theme, Themes: c.Themes}
	for _, n := range cp.ViewOrder {
		cp.Views[n] = c.Views[n]
	}
	return cp
}

func (m *Model) saveSettings(what string, cfg domain.Config, fromModal bool) tea.Cmd {
	return m.write(what, func(t *domain.Tree) (*domain.Change, error) { return t.PlanConfig(cfg) }, fromModal)
}

// settingsKey handles the settings screen: arrows or digits select a row
// (digit n is view n, the tab it shows as), t and T switch the theme
// (space or enter on its row too), space or enter toggles the mouse, enter
// edits a view, n adds one, d d deletes it, and m starts moving it.
func (m *Model) settingsKey(s string) tea.Cmd {
	if m.moving {
		return m.moveKey(s)
	}
	rows := m.settingsRows()
	m.setIdx = clamp(m.setIdx, 0, rows-1)
	cfg := m.settingsConfig()
	view := m.setIdx - 1 // index into cfg.ViewOrder; -1 on the theme row
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
		m.setIdx = clamp(int(s[0]-'0'), 1, rows-2)
	case "t", "T":
		m.startPreview()
		m.stepPreview(s == "T")
	case "space", "enter":
		if onMouse {
			return m.toggleMouse()
		}
		if view == -1 {
			m.startPreview()
			return nil
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

// themePreview shows themes on the issue screen before one is kept: the
// palette from before the preview (esc restores it), the theme names and
// the one being shown.
type themePreview struct {
	from  palette.Palette
	names []string
	idx   int
}

// startPreview shows the issue screen with a theme bar in the footer;
// nothing is saved until enter.
func (m *Model) startPreview() {
	names := palette.Names(m.tree.Project.Config.Themes)
	m.preview = &themePreview{from: m.th.Palette, names: names, idx: max(0, slices.Index(names, m.th.Palette.Name))}
	m.screen = screenIssues
}

// stepPreview shows the next (or previous) theme.
func (m *Model) stepPreview(back bool) {
	pv := m.preview
	step := 1
	if back {
		step = len(pv.names) - 1
	}
	pv.idx = (pv.idx + step) % len(pv.names)
	if p, err := palette.Resolve(pv.names[pv.idx], m.tree.Project.Config.Themes); err == nil {
		m.setTheme(theme.From(p, m.th.Profile))
	}
}

// previewKey handles keys during a theme preview: arrows, j/k and t/T
// move through the themes, enter keeps the one shown (saved to the user's
// own config), esc restores the one from before; both return to settings.
func (m *Model) previewKey(s string) tea.Cmd {
	switch s {
	case "right", "l", "down", "j", "t":
		m.stepPreview(false)
	case "left", "h", "up", "k", "T":
		m.stepPreview(true)
	case "enter":
		name := m.preview.names[m.preview.idx]
		m.preview, m.screen = nil, screenSettings
		cfg := m.settingsConfig()
		cfg.Theme = name
		return m.saveSettings("theme "+name, cfg, false)
	case "esc":
		m.setTheme(theme.From(m.preview.from, m.th.Profile))
		m.preview, m.screen = nil, screenSettings
	case "ctrl+c", "q":
		return tea.Quit
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
