package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// knowState is the knowledge screen: the entries the filter and the
// attention toggle leave, the selection, and whether the entry has the keys.
type knowState struct {
	filter    string
	attention bool // only entries that need an agent's attention
	cursor    int
	offset    int
	path      string // selected entry, kept across reloads
	onEntry   bool   // the entry pane has the keys (scrolling)
	docKey    string
	vp        viewport.Model // the entry, apart from the issue detail's
}

// knowledgeCodes are the check results that mark an entry for an agent.
var knowledgeCodes = []string{domain.CodeKnowledgeLink, domain.CodeKnowledgeSize, domain.CodeKnowledgeDrift, domain.CodeKnowledgeCommit}

// openKnowledge shows the knowledge screen, with path selected when given,
// and runs the check that marks entries needing attention.
func (m *Model) openKnowledge(path string) tea.Cmd {
	m.screen = screenKnowledge
	m.know.onEntry = false
	if path != "" {
		m.know.filter, m.know.attention, m.know.path = "", false, path
	}
	m.syncKnowledge()
	if !m.diagDone {
		return m.runCheck()
	}
	return nil
}

// knowledgePaths are the entries listed now: the filter, then attention.
func (m *Model) knowledgePaths() ([]string, error) {
	f, err := domain.ParseKnowledgeFilter(strings.Fields(m.know.filter))
	if err != nil {
		return nil, err
	}
	paths := m.tree.QueryKnowledge(f)
	if m.know.attention {
		paths = slices.DeleteFunc(paths, func(p string) bool { return len(m.attentionOf(p)) == 0 })
	}
	return paths, nil
}

// syncKnowledge keeps the cursor on the selected entry when the list changes.
func (m *Model) syncKnowledge() {
	paths, _ := m.knowledgePaths()
	if k := slices.Index(paths, m.know.path); k >= 0 {
		m.know.cursor = k
	}
	m.know.cursor = clamp(m.know.cursor, 0, len(paths)-1)
	if len(paths) > 0 {
		m.know.path = paths[m.know.cursor]
	}
}

// attentionOf returns the check's findings for an entry; empty until the
// check has run.
func (m *Model) attentionOf(path string) []domain.Diagnostic {
	var out []domain.Diagnostic
	for _, d := range m.diags {
		if d.File == ".prep/knowledge"+path && slices.Contains(knowledgeCodes, d.Code) {
			out = append(out, d)
		}
	}
	return out
}

// knowledgeKey handles the knowledge screen: arrows or j/k select, enter
// gives the entry the keys for scrolling, f filters, a toggles attention,
// o goes to an issue that changed the entry, backspace returns to the
// issue the screen was opened from.
func (m *Model) knowledgeKey(s string, k tea.KeyPressMsg) tea.Cmd {
	paths, _ := m.knowledgePaths()
	if m.know.onEntry {
		switch s {
		case "esc", "left", "h":
			m.know.onEntry = false
			return nil
		case "o":
			return m.openEntryLinks()
		}
		var cmd tea.Cmd
		m.know.vp, cmd = m.know.vp.Update(k)
		return cmd
	}
	move := func(to int) {
		m.know.cursor = clamp(to, 0, len(paths)-1)
		if len(paths) > 0 {
			m.know.path = paths[m.know.cursor]
		}
	}
	switch s {
	case "up", "k":
		move(m.know.cursor - 1)
	case "down", "j":
		move(m.know.cursor + 1)
	case "g", "home":
		move(0)
	case "G", "end":
		move(len(paths) - 1)
	case "enter", "right", "l":
		if len(paths) > 0 {
			m.know.onEntry = true
		}
	case "f", "/":
		m.filtering, m.filterErr = true, ""
		m.input.SetValue(m.know.filter)
		m.input.CursorEnd()
		return m.input.Focus()
	case "a":
		m.know.attention = !m.know.attention
		m.syncKnowledge()
		if m.know.attention && !m.diagDone {
			return m.flash("the check is still running; marks appear when it is done")
		}
	case "o":
		return m.openEntryLinks()
	case "backspace":
		if n := len(m.back); n > 0 {
			id := m.back[n-1]
			m.back = m.back[:n-1]
			m.screen = screenIssues
			return m.jump(id, false)
		}
		return m.flash("no earlier issue")
	case "esc":
		if m.know.filter != "" {
			m.know.filter = ""
			m.syncKnowledge()
			return nil
		}
		m.screen = screenIssues
	}
	return nil
}

// knowledgeFilterKey applies the filter bar on the knowledge screen.
func (m *Model) knowledgeFilterKey(text string) {
	if text != "" {
		if _, err := domain.ParseKnowledgeFilter(strings.Fields(text)); err != nil {
			m.filterErr = err.Error()
			return
		}
	}
	m.filtering, m.filterErr = false, ""
	m.input.Blur()
	m.know.filter = text
	m.syncKnowledge()
}

// openEntryLinks lists the issues that changed the selected entry; going
// to one returns to the issue screen.
func (m *Model) openEntryLinks() tea.Cmd {
	issues := m.tree.EntryIssues(m.know.path)
	if len(issues) == 0 {
		return m.flash("no issue has changed this entry yet")
	}
	keys := linkKeys
	d := &modal{kind: modalMenu, heading: "Changed by"}
	for n, id := range issues {
		if n >= len(keys) {
			break
		}
		target := id
		d.items = append(d.items, action{key: keys[n : n+1], label: shortID(id) + " " + m.tree.Issues[id].Title, run: func() tea.Cmd {
			m.screen = screenIssues
			return m.jump(target, false)
		}})
	}
	m.modal = d
	return nil
}

// knowledgePane draws the list of entries beside the selected entry.
func (m *Model) knowledgePane(w, h int) string {
	listW, entryW := ui.Split(w, listRatio, minListW, minDetailW)
	paths, err := m.knowledgePaths()
	if entryW == 0 {
		if m.know.onEntry {
			return m.entryPane(w, h)
		}
		return m.knowledgeList(paths, err, w, h)
	}
	list := m.knowledgeList(paths, err, listW, h)
	m.at.x += listW
	return lipgloss.JoinHorizontal(lipgloss.Top, list, m.entryPane(entryW, h))
}

func (m *Model) knowledgeList(paths []string, err error, w, h int) string {
	inner := w - 2 - 2*theme.Pad
	rows := h - 2
	title := "Knowledge"
	if m.know.attention {
		title += "  needs attention"
	}
	if m.know.filter != "" {
		title += "  / " + m.know.filter
	}
	if !m.diagDone && m.opts.Check != nil {
		title += "  (checking …)"
	}
	var bar []string
	if m.filtering {
		m.input.SetWidth(max(1, inner-4))
		bar = append(bar, m.input.View())
		if m.filterErr != "" {
			bar = append(bar, ui.Error(m.th, m.filterErr))
		}
		bar = append(bar, "")
		rows = max(1, rows-len(bar))
	}
	var body string
	switch {
	case err != nil:
		body = ui.Error(m.th, err.Error())
	case len(paths) == 0 && m.know.attention:
		body = ui.Empty(m.th, "Nothing needs attention", "a shows all entries again", inner, rows)
	case len(paths) == 0:
		body = ui.Empty(m.th, "No entries", "Agents write knowledge as they complete issues", inner, rows)
	default:
		c := m.know.cursor
		off := clamp(m.know.offset, 0, max(0, len(paths)-rows))
		if c < off {
			off = c
		}
		if c >= off+rows {
			off = c - rows + 1
		}
		m.know.offset = off
		var lines []string
		for k := off; k < len(paths) && k < off+rows; k++ {
			m.mark(fmt.Sprintf("know:%d", k), paneInner.x, paneInner.y+len(bar)+k-off, inner, 1)
			lines = append(lines, m.knowledgeRow(paths[k], k == c, inner))
		}
		body = strings.Join(lines, "\n")
	}
	if len(bar) > 0 {
		body = strings.Join(bar, "\n") + "\n" + body
	}
	m.pane("knowledge", w, h)
	return ui.Pane{Title: title, Body: body, Focused: !m.know.onEntry, Width: w, Height: h}.View(m.th)
}

// knowledgeRow is one entry: type, status, attention mark, title.
func (m *Model) knowledgeRow(path string, selected bool, width int) string {
	e := m.tree.Knowledge[path]
	marker := "  "
	titleStyle := m.th.S.Body
	if selected {
		marker = lipgloss.NewStyle().Foreground(m.th.C.Accent).Render("▌ ")
		titleStyle = m.th.S.Heading
	}
	mark := "  "
	if len(m.attentionOf(path)) > 0 {
		mark = lipgloss.NewStyle().Foreground(m.th.C.Warning).Render("! ")
	}
	status := e.Status
	if status == "" {
		status = "—"
	}
	line := marker + m.th.S.Muted.Render(fmt.Sprintf("%-11s", e.Type)) + m.th.S.Subtle.Render(fmt.Sprintf("%-11s", status)) + mark + titleStyle.Render(e.Title)
	line = ui.Fit(line, width)
	if selected {
		return lipgloss.NewStyle().Background(m.th.C.Selection).Width(width).Render(line)
	}
	return line
}

// entryPane renders the selected entry: meta, findings, body, and the
// issues that changed it.
func (m *Model) entryPane(w, h int) string {
	inner := w - 2 - 2*theme.Pad
	m.pane("entry", w, h)
	e := m.tree.Knowledge[m.know.path]
	if e == nil {
		return ui.Pane{Title: "Entry", Body: ui.Empty(m.th, "Nothing selected", "", inner, h-2), Width: w, Height: h}.View(m.th)
	}
	key := fmt.Sprintf("%s|%d|%d|%d", e.Path, inner, m.gen, len(m.diags))
	if key != m.know.docKey {
		var b strings.Builder
		// Entries lead with their own title heading; the meta goes under it.
		body, heading := e.Body, "# "+e.Title
		if first, rest, _ := strings.Cut(body, "\n"); strings.HasPrefix(first, "# ") {
			heading, body = first, strings.TrimSpace(rest)
		}
		b.WriteString(heading + "\n\n")
		meta := []string{"**" + e.Type + "**"}
		if e.Status != "" {
			meta = append(meta, e.Status)
		}
		meta = append(meta, "`"+e.Path+"`")
		b.WriteString(strings.Join(meta, " · ") + "\n\n")
		if e.Description != "" {
			b.WriteString("_" + e.Description + "_\n\n")
		}
		if len(e.Scope) > 0 {
			fmt.Fprintf(&b, "Scope: `%s`", strings.Join(e.Scope, "`, `"))
			if e.ConfirmedCommit != "" {
				fmt.Fprintf(&b, " · confirmed at `%.7s`", e.ConfirmedCommit)
			}
			b.WriteString("\n\n")
		}
		if ds := m.attentionOf(e.Path); len(ds) > 0 {
			b.WriteString("**Needs an agent's attention**\n\n")
			for _, d := range ds {
				fmt.Fprintf(&b, "- %s %s\n", d.Code, d.Message)
			}
			b.WriteString("\n")
		}
		if issues := m.tree.EntryIssues(e.Path); len(issues) > 0 {
			b.WriteString("**Changed by** (o goes to one)\n\n")
			for _, id := range issues {
				fmt.Fprintf(&b, "- `%s` %s\n", shortID(id), m.tree.Issues[id].Title)
			}
			b.WriteString("\n")
		}
		b.WriteString("---\n\n" + body + "\n")
		doc, err := ui.Markdown(m.th, b.String(), inner)
		if err != nil {
			doc = ui.Error(m.th, err.Error())
		}
		if !strings.HasPrefix(m.know.docKey, e.Path+"|") {
			m.know.vp.GotoTop()
		}
		m.know.docKey = key
		m.know.vp.SetContent(doc)
	}
	m.know.vp.SetWidth(inner)
	m.know.vp.SetHeight(h - 2)
	return ui.Pane{Title: e.Path, Body: m.know.vp.View(), Focused: m.know.onEntry, Width: w, Height: h}.View(m.th)
}
