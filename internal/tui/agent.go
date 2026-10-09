package tui

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// The Agent screen shows what an agent does with prep, live: the issue it
// works on, a feed of its actions grouped by issue, and its usage when its
// harness reports tokens. It reads the activity stream (Options.Activity);
// the person runs it in a split next to the agent.

const (
	freshFor  = 2 * time.Second  // a new line stays marked this long
	agentTick = 30 * time.Second // relative times refresh while the screen shows
	idleAfter = 2 * time.Minute  // a quiet agent is called idle after this
)

// agentState is the Agent screen: the stream, the agent shown (empty
// follows the most recent), the feed's scroll and what just arrived.
type agentState struct {
	events     []activity.Event
	err        error
	loaded     bool
	agent      string // pinned agent key; "" follows the newest
	view       int    // the view shown: agentIssue, agentActivity, agentUsage
	offsets    [3]int // lines scrolled past in each view (the feed's from the newest)
	shownIssue string // the issue the Issue view shows, to scroll it to the top when it changes
	doc        string // the Issue view's rendered document, for docKey
	docKey     string
	newest     time.Time // the newest event seen, to mark what arrives after it
	freshAfter time.Time // events after this are fresh until freshUntil
	freshUntil time.Time
	ticking    bool
	// guide caches the Now card's guide for guideID in guideTree: building
	// one validates the whole tree, too much for every frame.
	guide     domain.Guide
	guideID   string
	guideTree *domain.Tree
}

type activityMsg struct {
	events []activity.Event
	err    error
}
type agentTickMsg struct{}
type agentFadeMsg struct{}

// loadActivity reads the stream; nil without a loader.
func (m *Model) loadActivity() tea.Cmd {
	if m.opts.Activity == nil {
		return nil
	}
	return func() tea.Msg {
		ev, err := m.opts.Activity()
		return activityMsg{ev, err}
	}
}

// applyActivity installs a read of the stream and marks what is new.
func (m *Model) applyActivity(msg activityMsg) tea.Cmd {
	a := &m.agentS
	a.err = msg.err
	if msg.err != nil {
		return nil
	}
	a.events, a.loaded = msg.events, true
	var newest time.Time
	if n := len(a.events); n > 0 {
		newest = a.events[n-1].At
	}
	first := a.newest.IsZero()
	if newest.After(a.newest) {
		if !first {
			a.freshAfter, a.freshUntil = a.newest, m.clock().Add(freshFor)
		}
		a.newest = newest
		if !first && m.screen == screenAgent {
			return m.after(freshFor, func(time.Time) tea.Msg { return agentFadeMsg{} })
		}
	}
	return nil
}

// openAgent shows the Agent screen and starts the slow tick for relative
// times.
func (m *Model) openAgent() tea.Cmd {
	m.screen = screenAgent
	m.agentS.offsets = [3]int{}
	return tea.Batch(m.loadActivity(), m.startAgentTick())
}

func (m *Model) startAgentTick() tea.Cmd {
	if m.agentS.ticking {
		return nil
	}
	m.agentS.ticking = true
	return m.after(agentTick, func(time.Time) tea.Msg { return agentTickMsg{} })
}

// agentTicked redraws relative times while the screen shows and stops the
// tick otherwise, so nothing runs while nothing is watched.
func (m *Model) agentTicked() tea.Cmd {
	m.agentS.ticking = false
	if m.screen != screenAgent {
		return nil
	}
	return m.startAgentTick()
}

// agents lists the agents in the stream, the most recently active first.
func (m *Model) agents() []string { return activity.Agents(m.agentS.events) }

// shownAgent is the pinned agent while it is in the stream, else the most
// recent one.
func (m *Model) shownAgent() string {
	as := m.agents()
	if m.agentS.agent != "" && slices.Contains(as, m.agentS.agent) {
		return m.agentS.agent
	}
	if len(as) > 0 {
		return as[0]
	}
	return ""
}

// agentKey handles the Agent screen: j/k and pages scroll the view shown,
// n cycles the agents, enter opens the current issue. tab and the digits
// switch the views and esc leaves, as on every screen.
func (m *Model) agentKey(s string) tea.Cmd {
	a := &m.agentS
	off := &a.offsets[a.view]
	switch s {
	case "down", "j":
		*off++
	case "up", "k":
		*off = max(0, *off-1)
	case "pgdown":
		*off += m.bodyHeight() / 2
	case "pgup":
		*off = max(0, *off-m.bodyHeight()/2)
	case "g", "home":
		*off = 0
	case "G", "end":
		*off = 1 << 30 // rendering clamps it
	case "n":
		as := m.agents()
		if len(as) < 2 {
			return m.flash("only one agent so far")
		}
		k := (slices.Index(as, m.shownAgent()) + 1) % len(as)
		a.agent, a.offsets = as[k], [3]int{}
		if k == 0 {
			a.agent = "" // the newest: follow whoever is active
			return m.flash("following the most recent agent")
		}
		return m.flash("showing " + agentLabel(as[k], m.agentS.events))
	case "enter":
		v := m.agentView(m.shownAgent())
		if v.focus == "" || m.tree.Issues[v.focus] == nil {
			return m.flash("no current issue")
		}
		m.pushBack()
		m.screen = screenIssues
		cmd := m.jump(v.focus, false)
		m.focus = focusDetail
		return cmd
	}
	return nil
}

// agentLabel names an agent for people: the actor of its latest prep
// command (the name the agent gives itself; its harness's bridge uses its
// own), else of any event, and its session.
func agentLabel(key string, events []activity.Event) string {
	for k := len(events) - 1; k >= 0; k-- {
		if e := events[k]; e.Agent() == key && (e.Kind == activity.KindPrep || e.Kind == activity.KindFocus) {
			if e.Session != "" {
				return e.Actor + " · " + shortSession(e.Session)
			}
			return e.Actor
		}
	}
	for k := len(events) - 1; k >= 0; k-- {
		if e := events[k]; e.Agent() == key {
			if e.Session != "" {
				return e.Actor + " · " + shortSession(e.Session)
			}
			return e.Actor
		}
	}
	return strings.TrimPrefix(strings.TrimPrefix(key, "actor:"), "session:")
}

func shortSession(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// agentData is what the screen shows of one agent, computed from the
// stream.
type agentData struct {
	key, label string
	events     []activity.Event // the agent's, oldest first
	chapter    []string         // the issue each event belongs to
	focus      string
	last       time.Time

	contextUsed, contextWindow int
	tokens                     activity.Tokens
	requests                   int
	cost                       float64
	areas                      map[string]int // characters read back
	reads                      []read         // knowledge reads, largest first
}

type read struct {
	target string
	chars  int
}

func (m *Model) agentView(key string) agentData {
	d := agentData{key: key, label: agentLabel(key, m.agentS.events), areas: map[string]int{}}
	current := ""
	byTarget := map[string]int{}
	for _, e := range m.agentS.events {
		if e.Agent() != key {
			continue
		}
		if e.MovesFocus() {
			current = e.Issue
		}
		d.events = append(d.events, e)
		d.chapter = append(d.chapter, current)
		d.last = e.At
		switch e.Kind {
		case activity.KindRequest:
			// A turn's end reports context and cost without tokens; only
			// events with tokens are model requests.
			if e.Tokens != nil {
				d.requests++
				d.tokens.Input += e.Tokens.Input
				d.tokens.Output += e.Tokens.Output
				d.tokens.CacheRead += e.Tokens.CacheRead
				d.tokens.CacheWrite += e.Tokens.CacheWrite
			}
			if e.Context != nil {
				d.contextUsed, d.contextWindow = e.Context.Tokens, e.Context.Window
			}
			d.cost += e.CostUSD
		default:
			if e.Chars > 0 {
				area := e.Area
				if area == "" {
					area = activity.AreaOther
				}
				d.areas[area] += e.Chars
				if area == activity.AreaKnowledge && e.Target != "" {
					byTarget[e.Target] += e.Chars
				}
			}
		}
	}
	d.focus = current
	for t, c := range byTarget {
		d.reads = append(d.reads, read{t, c})
	}
	sort.Slice(d.reads, func(i, j int) bool {
		if d.reads[i].chars != d.reads[j].chars {
			return d.reads[i].chars > d.reads[j].chars
		}
		return d.reads[i].target < d.reads[j].target
	})
	return d
}

// ago renders how long ago t was: now, 12s, 4m, 3h, 2d.
func ago(now, t time.Time) string {
	d := now.Sub(t)
	switch {
	case d < 5*time.Second:
		return "now"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// agentPane shows the agent in the view chosen in the second tier.
func (m *Model) agentPane(w, h int) string {
	d := m.agentView(m.shownAgent())
	if m.agentS.err == nil && len(m.agentS.events) == 0 {
		inner := ui.Inner(w)
		m.pane("feed", w, h)
		return ui.Pane{Title: "Agent", Body: ui.Empty(m.th, "No agent activity yet", "Each prep command an agent runs in this project shows up here, live", inner, h-2), Width: w, Height: h}.View(m.th)
	}
	switch m.agentS.view {
	case agentActivity:
		return m.feedPane(d, w, h)
	case agentUsage:
		return m.usagePane(d, w, h)
	}
	return m.issuePane(d, w, h)
}

// agentStatus names the agent shown and whether it is active, for the
// second tier's right end: with detail 2 also its position among the
// agents, with 1 its label and liveliness, with 0 only the liveliness.
func (m *Model) agentStatus(detail int) string {
	key := m.shownAgent()
	if key == "" {
		return ""
	}
	d := m.agentView(key)
	now := m.clock()
	status := ui.Note(m.th, "● active "+ago(now, d.last), ui.ToneAccent)
	if now.Sub(d.last) >= idleAfter {
		status = ui.Note(m.th, "idle for "+ago(now, d.last), ui.ToneMuted)
	}
	if detail >= 1 {
		status = m.th.S.Muted.Render(d.label) + space(theme.SpaceM) + status
	}
	if as := m.agents(); detail >= 2 && len(as) > 1 {
		pos := fmt.Sprintf("%sagent %d/%d", space(theme.SpaceM), slices.Index(as, d.key)+1, len(as))
		if m.agentS.agent != "" {
			pos += " pinned"
		}
		status += m.th.S.Muted.Render(pos)
	}
	return status
}

// issuePane is the agent's current issue in full: where it stands (state,
// step, acceptance, the next transition, the knowledge its guide points
// to) over the issue's document, scrolled with j/k.
func (m *Model) issuePane(d agentData, w, h int) string {
	inner := ui.Inner(w)
	rows := max(1, h-2)
	m.pane("issue", w, h)
	switch {
	case m.agentS.err != nil:
		return ui.Pane{Title: "Issue", Body: ui.Error(m.th, m.agentS.err.Error()), Width: w, Height: h}.View(m.th)
	case d.focus == "" || m.tree.Issues[d.focus] == nil:
		return ui.Pane{Title: "Issue", Body: ui.Empty(m.th, "No current issue", "its next prep guide or write sets one; prep focus <id> too", inner, rows), Width: w, Height: h}.View(m.th)
	}
	a := &m.agentS
	if a.shownIssue != d.focus {
		a.shownIssue, a.offsets[agentIssue] = d.focus, 0
	}
	lines := m.nowCard(d.focus, inner)[1:] // the pane's title names the issue
	if ks := a.guide.Knowledge; len(ks) > 0 {
		lines = append(lines, "", m.th.S.Heading.Render("Knowledge"))
		for _, k := range ks[:min(3, len(ks))] {
			lines = append(lines, ui.Fit(m.th.S.Body.Render(k.Title)+"  "+m.th.S.Subtle.Render(k.Path), inner))
		}
	}
	if key := fmt.Sprintf("%s|%d|%d", d.focus, inner, m.gen); key != a.docKey {
		doc, err := ui.Markdown(m.th, detailSections(m.tree, d.focus), inner)
		if err != nil {
			doc = ui.Error(m.th, err.Error())
		}
		a.doc, a.docKey = doc, key
	}
	lines = append(lines, "")
	lines = append(lines, strings.Split(a.doc, "\n")...)
	a.offsets[agentIssue] = clamp(a.offsets[agentIssue], 0, max(0, len(lines)-rows))
	off := a.offsets[agentIssue]
	title := shortID(d.focus) + " " + m.tree.Issues[d.focus].Title
	if len(lines) > rows {
		title += fmt.Sprintf("  %d%%", 100*(off+rows)/len(lines))
	}
	return ui.Pane{Title: title, Body: strings.Join(lines[off:min(len(lines), off+rows)], "\n"), Focused: true, Width: w, Height: h}.View(m.th)
}

// nowCard is the current issue: title, state, kind and step, acceptance,
// and the next transition with what still blocks it.
func (m *Model) nowCard(id string, width int) []string {
	t := m.tree
	i := t.Issues[id]
	s := t.State(id)
	if a := &m.agentS; a.guideTree != t || a.guideID != id {
		a.guide, a.guideID, a.guideTree = t.BuildGuide(id, func(string) string { return "" }, nil), id, t
	}
	g := m.agentS.guide
	lines := []string{
		ui.Fit(m.th.S.Subtle.Render(shortID(id))+" "+m.th.S.Title.Render(i.Title), width),
		ui.Fit(ui.StateBadge(m.th, s)+ui.KindTag(m.th, i.Kind)+m.th.S.Muted.Render("step "+g.Step), width),
	}
	if n := len(i.Criteria); n > 0 {
		done := 0
		for _, c := range i.Criteria {
			if c.Checked {
				done++
			}
		}
		lines = append(lines, ui.Fit(ui.Progress(m.th, done, n)+m.th.S.Muted.Render(" acceptance"), width))
	}
	if tr, ok := nextTransition(g, t.Stale(id)); ok {
		if tr.Allowed {
			lines = append(lines, ui.Fit(m.th.S.Muted.Render("next ")+ui.Note(m.th, tr.Command, ui.ToneSuccess), width))
		} else {
			why := ""
			if len(tr.Unmet) > 0 {
				why = " — " + tr.Unmet[0].Message
			}
			lines = append(lines, ui.Fit(m.th.S.Muted.Render("next ")+ui.Note(m.th, string(tr.Op)+why, ui.ToneWarning), width))
		}
	}
	return lines
}

// nextTransition is the step that moves the issue on: ack when it is
// stale, else the first besides ack, release and drop.
func nextTransition(g domain.Guide, stale bool) (domain.Transition, bool) {
	for _, tr := range g.Transitions {
		if stale && tr.Op == domain.OpAck {
			return tr, true
		}
	}
	for _, tr := range g.Transitions {
		if tr.Op != domain.OpAck && tr.Op != domain.OpRelease && tr.Op != domain.OpDrop {
			return tr, true
		}
	}
	return domain.Transition{}, false
}

// usagePane shows what the agent read back, and its tokens when its
// harness reports them.
func (m *Model) usagePane(d agentData, w, h int) string {
	inner := ui.Inner(w)
	m.pane("usage", w, h)
	var lines []string
	if d.requests > 0 {
		lines = append(lines, ui.Meter(m.th, "context", d.contextUsed, d.contextWindow, inner))
		tk := d.tokens
		lines = append(lines, ui.Fit(m.th.S.Muted.Render(fmt.Sprintf("%-10s", "tokens"))+m.th.S.Body.Render(fmt.Sprintf("in %s  out %s", ui.Count(tk.Input), ui.Count(tk.Output))), inner))
		lines = append(lines, ui.Fit(m.th.S.Muted.Render(fmt.Sprintf("%-10s", "cache"))+m.th.S.Body.Render(fmt.Sprintf("%s read  %s written", ui.Count(tk.CacheRead), ui.Count(tk.CacheWrite))), inner))
		figures := fmt.Sprintf("%d", d.requests)
		if d.cost > 0 {
			figures += fmt.Sprintf(" · $%.2f", d.cost)
		}
		lines = append(lines, ui.Fit(m.th.S.Muted.Render(fmt.Sprintf("%-10s", "requests"))+m.th.S.Body.Render(figures), inner), "")
	}
	if d.focus != "" {
		cmds, checked, know := 0, 0, 0
		for k, e := range d.events {
			if d.chapter[k] != d.focus {
				continue
			}
			if e.Kind == activity.KindPrep {
				cmds++
			}
			if e.Op == "criterion" {
				checked += strings.Count(e.Verb, "checked criterion") - strings.Count(e.Verb, "unchecked criterion")
			}
			if e.Area == activity.AreaKnowledge {
				know += e.Chars
			}
		}
		lines = append(lines, m.th.S.Heading.Render("This issue"),
			ui.Fit(m.th.S.Body.Render(fmt.Sprintf("%d commands · %d checked · %s knowledge", cmds, checked, ui.Count(know/4))), inner), "")
	}
	largest := 0
	for _, c := range d.areas {
		largest = max(largest, c)
	}
	if largest > 0 {
		lines = append(lines, m.th.S.Heading.Render("Read back")+m.th.S.Subtle.Render("  ≈ tokens"))
		for _, area := range []string{activity.AreaIssue, activity.AreaKnowledge, activity.AreaCode, activity.AreaPrep, activity.AreaOther} {
			if c := d.areas[area]; c > 0 {
				lines = append(lines, ui.BarRow(m.th, area, c, largest, ui.Count(c/4), inner))
			}
		}
	}
	if len(d.reads) > 0 {
		lines = append(lines, "", m.th.S.Heading.Render("Largest knowledge reads"))
		for _, r := range d.reads[:min(3, len(d.reads))] {
			lines = append(lines, ui.Fit(m.th.S.Body.Render(fmt.Sprintf("%5s ", ui.Count(r.chars/4)))+m.th.S.Muted.Render(r.target), inner))
		}
	}
	if d.requests == 0 {
		lines = append(lines, "", m.th.S.Subtle.Render("Token figures appear when the harness reports them."))
	}
	rows := max(0, h-2)
	off := clamp(m.agentS.offsets[agentUsage], 0, max(0, len(lines)-rows))
	m.agentS.offsets[agentUsage] = off
	lines = lines[off:min(len(lines), off+rows)]
	return ui.Pane{Title: "Usage", Body: strings.Join(lines, "\n"), Width: w, Height: h}.View(m.th)
}

// feedPane is the agent's activity, newest first, under a heading per
// issue; completing an issue gets its own line.
func (m *Model) feedPane(d agentData, w, h int) string {
	inner := ui.Inner(w)
	rows := max(1, h-2)
	m.pane("feed", w, h)
	now := m.clock()
	var lines []string
	chapter := "\x00"
	for k := len(d.events) - 1; k >= 0; k-- {
		e := d.events[k]
		if e.Kind == activity.KindRequest {
			continue
		}
		if c := d.chapter[k]; c != chapter {
			chapter = c
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			if i := m.tree.Issues[c]; i != nil {
				lines = append(lines, ui.Chapter(m.th, shortID(c), i.Title, m.tree.State(c), inner))
			} else {
				lines = append(lines, ui.Fit(m.th.S.Subtle.Render("no current issue"), inner))
			}
		}
		if e.Op == string(domain.OpComplete) && e.Kind == activity.KindPrep {
			n := 0
			for j, f := range d.events {
				if d.chapter[j] == e.Issue && f.Kind == activity.KindPrep {
					n++
				}
			}
			lines = append(lines, ui.Celebrate(m.th, fmt.Sprintf("done in %d prep commands", n), inner))
		}
		verb, target := e.Verb, e.Target
		if verb == "" {
			verb = e.Op
		}
		if e.Kind == activity.KindPrep && e.Area == activity.AreaIssue && e.Issue == chapter {
			target = "" // the chapter heading names it
		}
		fresh := e.At.After(m.agentS.freshAfter) && !m.agentS.freshAfter.IsZero() && now.Before(m.agentS.freshUntil)
		lines = append(lines, ui.ActivityLine(m.th, ui.Activity{Area: e.Area, Ago: ago(now, e.At), Verb: verb, Target: target, Failed: e.Failed, Fresh: fresh}, inner))
	}
	var body string
	if len(lines) == 0 {
		body = ui.Empty(m.th, "Nothing yet", "Each prep command the agent runs lands here", inner, rows)
	} else {
		m.agentS.offsets[agentActivity] = clamp(m.agentS.offsets[agentActivity], 0, max(0, len(lines)-rows))
		end := min(len(lines), m.agentS.offsets[agentActivity]+rows)
		body = strings.Join(lines[m.agentS.offsets[agentActivity]:end], "\n")
	}
	title := "Activity"
	if off := m.agentS.offsets[agentActivity]; off > 0 {
		title += fmt.Sprintf("  ↑ %d newer", off)
	}
	return ui.Pane{Title: title, Body: body, Width: w, Height: h}.View(m.th)
}

var agentBindings = []binding{
	bind("Agent", "tab 1-3", "issue, activity or usage", true),
	bind("Agent", "↑↓ j k", "scroll the view", false),
	bind("Agent", "g G pgup pgdn", "top, bottom, page", false),
	bind("Agent", "n", "next agent (back to the first follows the most recent)", true),
	bind("Agent", "enter", "open the current issue", true),
	bind("Agent", "y", "copy the current issue's ID", false),
	bind("Screens", "i b", "issues, knowledge", false),
	bind("Screens", "⌫", "back to where you were", false),
	bind("Screens", "esc w", "back to the issues", true),
	bind("Screens", "c", "check", false),
	bind("Screens", "s", "settings", false),
	bind("Screens", "?", "all keys", true),
	bind("Screens", "q", "quit", true),
}
