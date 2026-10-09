package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// The navigation is two tiers: the screens bar on top (Issues, Agent,
// Knowledge, with Check and Settings at its right end), and under it the
// views of the current screen (the issue lists, the Agent views). tab and
// the digits always move in the second tier; one back stack covers both.

// navScreens are the first tier, in bar order.
var navScreens = []screen{screenIssues, screenAgent, screenKnowledge}

var navLabels = map[screen]string{screenIssues: "Issues", screenAgent: "Agent", screenKnowledge: "Knowledge"}

// The Agent screen's views.
const (
	agentIssue = iota
	agentActivity
	agentUsage
)

var agentViews = []string{"Issue", "Activity", "Usage"}

// shortHeight is the terminal height below which the two tiers share one
// line, so the TUI fits in a strip above or below the agent; from
// tallHeight on, a blank row sets the navigation off from the body.
const (
	shortHeight = 16
	tallHeight  = 30
)

// headerH is how many lines the navigation takes.
func (m *Model) headerH() int {
	switch {
	case m.h < shortHeight:
		return 1
	case m.h >= tallHeight:
		return 2 + theme.Row
	}
	return 2
}

// gutter is the space at the screen's left and right edges of the header
// and footer: as wide as a pane's border and padding, so their text lines
// up with the text inside the panes.
var gutter = strings.Repeat(" ", theme.SpaceM)

func space(n int) string { return strings.Repeat(" ", n) }

// place is where the person was: a screen and what it had selected.
type place struct {
	screen screen
	issue  string
	focus  focus
	entry  string
}

const maxBack = 100

func (m *Model) here() place {
	return place{screen: m.screen, issue: m.selected(), focus: m.focus, entry: m.know.path}
}

// pushBack remembers the current place before the person moves on.
func (m *Model) pushBack() {
	p := m.here()
	if n := len(m.back); n > 0 && m.back[n-1] == p {
		return
	}
	m.back = append(m.back, p)
	if len(m.back) > maxBack {
		m.back = m.back[len(m.back)-maxBack:]
	}
}

// goBack returns to the last remembered place: its screen and selection.
func (m *Model) goBack() tea.Cmd {
	n := len(m.back)
	if n == 0 || m.tree == nil {
		return m.flash("nowhere to go back to")
	}
	p := m.back[n-1]
	m.back = m.back[:n-1]
	m.moving, m.linkMode = false, false
	switch p.screen {
	case screenIssues:
		m.screen = screenIssues
		if p.issue == "" || m.tree.Issues[p.issue] == nil {
			return nil
		}
		cmd := m.jump(p.issue, false)
		m.focus = p.focus
		return cmd
	case screenKnowledge:
		if m.tree.Knowledge[p.entry] == nil {
			p.entry = ""
		}
		return m.openKnowledge(p.entry)
	case screenAgent:
		return m.openAgent()
	case screenCheck:
		m.screen = screenCheck
		return m.runCheck()
	}
	m.screen = p.screen
	return nil
}

// goScreen moves to a screen from the bar or its key, remembering where
// the person was.
func (m *Model) goScreen(sc screen) tea.Cmd {
	if m.tree == nil || sc == m.screen {
		return nil
	}
	m.pushBack()
	m.moving, m.linkMode = false, false
	switch sc {
	case screenAgent:
		return m.openAgent()
	case screenKnowledge:
		return m.openKnowledge("")
	case screenCheck:
		m.screen = sc
		m.page.GotoTop()
		return m.runCheck()
	case screenSettings:
		m.page.GotoTop()
	}
	m.screen = sc
	return nil
}

// toggleScreen is a screen's key: it opens the screen, or from the screen
// itself returns to the issues.
func (m *Model) toggleScreen(sc screen) tea.Cmd {
	if m.screen == sc {
		return m.goScreen(screenIssues)
	}
	return m.goScreen(sc)
}

// views is the second tier of the current screen; nil when it has none.
func (m *Model) views() []ui.Tab {
	switch m.screen {
	case screenIssues:
		tabs := make([]ui.Tab, 0, len(m.tabs))
		for _, tb := range m.tabs {
			tabs = append(tabs, ui.Tab{Label: tb.name, Count: tb.count})
		}
		return tabs
	case screenAgent:
		tabs := make([]ui.Tab, 0, len(agentViews))
		for _, v := range agentViews {
			tabs = append(tabs, ui.Tab{Label: v, Count: -1})
		}
		return tabs
	}
	return nil
}

func (m *Model) activeView() int {
	switch m.screen {
	case screenIssues:
		return m.active
	case screenAgent:
		return m.agentS.view
	}
	return -1
}

// setView selects view n of the current screen, wrapping around.
func (m *Model) setView(n int) tea.Cmd {
	vs := m.views()
	if len(vs) == 0 {
		return nil
	}
	n = (n%len(vs) + len(vs)) % len(vs)
	switch m.screen {
	case screenIssues:
		m.switchTab(n)
	case screenAgent:
		m.agentS.view = n
	}
	return nil
}

// header draws both tiers and marks what can be clicked: the screens, the
// check counts, the settings gear and the views.
func (m *Model) header() string {
	left := gutter + m.th.S.Title.Render("prep") + space(theme.SpaceL)
	var items []ui.NavItem
	active := -1
	for k, sc := range navScreens {
		it := ui.NavItem{Label: navLabels[sc]}
		if sc == m.screen {
			active = k
		}
		items = append(items, it)
	}
	xs, ws := ui.NavSpans(items)
	for k := range xs {
		m.markAt(fmt.Sprintf("screen:%d", k), lipgloss.Width(left)+xs[k], 0, ws[k], 1, zRow)
	}
	line := left + ui.Nav(m.th, items, active)

	right := m.utilities()
	end := m.w - theme.SpaceM // the right gutter
	tier2 := ""
	if m.headerH() == 1 {
		sep := space(theme.SpaceM) + m.th.S.Subtle.Render("›") + space(theme.SpaceM-theme.SpaceS)
		x := lipgloss.Width(line) + lipgloss.Width(sep)
		if tier2 = m.secondTier(end-x-lipgloss.Width(right)-theme.SpaceM, x, 0); tier2 != "" {
			line += sep + tier2
		}
	} else {
		// Tabs pad their labels by SpaceS, so the bar starts that much
		// left of the gutter and the labels line up with it.
		x := theme.SpaceM - theme.SpaceS
		tier2 = space(x) + m.secondTier(end-x, x, 1)
	}
	x := max(lipgloss.Width(line)+theme.SpaceM, end-lipgloss.Width(right))
	m.markUtilities(x)
	line += space(max(0, x-lipgloss.Width(line))) + right
	out := ui.Fit(line, m.w)
	if m.headerH() > 1 {
		out += "\n" + ui.Fit(tier2, m.w)
	}
	if m.headerH() > 2 {
		out += strings.Repeat("\n", theme.Row)
	}
	return out
}

// utilities is the bar's right end: the check's counts and the gear.
func (m *Model) utilities() string {
	out := ""
	if m.diagDone {
		errs, warns := m.diagCounts()
		out = ui.Note(m.th, fmt.Sprintf("✕ %d", errs), toneIf(errs > 0, ui.ToneError)) + space(theme.SpaceM) + ui.Note(m.th, fmt.Sprintf("▲ %d", warns), toneIf(warns > 0, ui.ToneWarning)) + space(theme.SpaceL)
	}
	return out + ui.Note(m.th, "⚙", toneIf(m.screen == screenSettings, ui.ToneAccent))
}

// markUtilities marks the counts (they open the check) and the gear
// (settings), drawn from x on.
func (m *Model) markUtilities(x int) {
	if m.diagDone {
		errs, warns := m.diagCounts()
		w := len(fmt.Sprintf("✕ %d  ▲ %d", errs, warns)) - 4 // ✕ and ▲ are three bytes, one cell
		m.markAt("check", x, 0, w, 1, zRow)
		x += w + theme.SpaceL
	}
	m.markAt("settings", x, 0, 1, 1, zRow)
}

func (m *Model) diagCounts() (errs, warns int) {
	for _, d := range m.diags {
		if d.Severity == domain.SevError {
			errs++
		} else {
			warns++
		}
	}
	return errs, warns
}

// secondTier draws the current screen's views as tabs at (x, y), marking
// each, or a caption for screens without views.
func (m *Model) secondTier(width, x, y int) string {
	if width <= 0 {
		return ""
	}
	vs := m.views()
	if len(vs) == 0 {
		// SpaceS in, where a tab's label starts.
		return ui.Fit(space(theme.SpaceS)+m.th.S.Muted.Render(m.caption()), width)
	}
	xs, ws := ui.TabSpans(vs, width)
	for k := range xs {
		m.markAt(fmt.Sprintf("tab:%d", k), x+xs[k], y, ws[k], 1, zRow)
	}
	line := ui.Tabs(m.th, vs, m.activeView(), width)
	if m.screen == screenAgent && y > 0 {
		// The fullest status that fits: label, liveliness, position.
		for detail := 2; detail >= 0; detail-- {
			if s := m.agentStatus(detail); s != "" && lipgloss.Width(line)+lipgloss.Width(s)+theme.SpaceM <= width {
				line += space(width-lipgloss.Width(line)-lipgloss.Width(s)) + s
				break
			}
		}
	}
	return line
}

// caption stands in the second tier of screens without views.
func (m *Model) caption() string {
	switch m.screen {
	case screenKnowledge:
		if m.tree == nil {
			return ""
		}
		paths, _ := m.knowledgePaths()
		if n := len(m.tree.Knowledge); len(paths) != n {
			return fmt.Sprintf("%d of %d entries", len(paths), n)
		}
		return fmt.Sprintf("%d entries", len(paths))
	case screenCheck:
		return "Check"
	case screenSettings:
		return "Settings"
	}
	return ""
}
