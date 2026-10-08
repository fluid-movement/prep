package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Gallery renders every token and component in every variant from fixed
// sample values. It is the place to check a visual change, and the snapshot
// tests render it.
func Gallery(t *theme.Theme, width int) string {
	var b strings.Builder
	section := func(title string) {
		if b.Len() > 0 {
			cur := strings.TrimRight(b.String(), "\n")
			b.Reset()
			b.WriteString(cur + strings.Repeat("\n", theme.Section+1))
		}
		b.WriteString(t.S.Title.Render(title) + "\n")
	}

	section("Colors")
	var swatches []string
	for _, tok := range t.Tokens() {
		swatches = append(swatches, lipgloss.NewStyle().Foreground(tok.Color).Render("██")+" "+t.S.Muted.Render(tok.Name))
	}
	b.WriteString(wrapItems(swatches, width, "  "))

	section("Text styles")
	for _, s := range []struct {
		name  string
		style lipgloss.Style
	}{{"title", t.S.Title}, {"heading", t.S.Heading}, {"body", t.S.Body}, {"muted", t.S.Muted}, {"subtle", t.S.Subtle}, {"code", t.S.Code}, {"key", t.S.Key}} {
		b.WriteString(Fit(s.style.Render("The quick brown fox")+"  "+t.S.Subtle.Render(s.name), width) + "\n")
	}

	section("State badges")
	var badges []string
	for _, s := range []domain.State{domain.StateOpen, domain.StateDefined, domain.StateReady, domain.StateInProgress, domain.StateDone, domain.StateDropped} {
		badges = append(badges, StateBadge(t, s))
	}
	b.WriteString(wrapItems(badges, width, ""))

	section("Kind tags")
	var kinds []string
	for _, k := range []domain.Kind{domain.KindCode, domain.KindManual, domain.KindResearch, domain.KindDecision} {
		kinds = append(kinds, KindTag(t, k))
	}
	b.WriteString(wrapItems(kinds, width, ""))

	section("Progress and notes")
	b.WriteString(Fit(Progress(t, 0, 3)+"   "+Progress(t, 2, 3)+"   "+Progress(t, 3, 3), width) + "\n")
	b.WriteString(Fit(Note(t, "blocked", ToneWarning)+"  "+Note(t, "stale", ToneError)+"  "+Note(t, "actionable", ToneSuccess)+"  "+Note(t, "claimed", ToneAccent)+"  "+Note(t, "note", ToneMuted), width))

	section("Navigation")
	// The screens bar: Agent active, then idle; the views of the screen below.
	live, idle := ToneAccent, ToneMuted
	b.WriteString(Fit(Nav(t, []NavItem{{Label: "Issues"}, {Label: "Agent", Dot: &live}, {Label: "Knowledge"}}, 1), width) + "\n")
	b.WriteString(Fit(Nav(t, []NavItem{{Label: "Issues"}, {Label: "Agent", Dot: &idle}, {Label: "Knowledge"}}, 0), width) + "\n")

	section("Tabs")
	// The views prep init writes, then two a project added.
	tabs := []Tab{{"Unresolved", 17}, {"All", 25}, {"In progress", 2}, {"Release 0.1.0", 6}}
	b.WriteString(Tabs(t, tabs, 0, width) + "\n")
	b.WriteString(Tabs(t, tabs, 2, width) + "\n")
	b.WriteString(Tabs(t, tabs, 3, min(width, 40)))

	section("List rows")
	rows := []Row{
		{ID: "T4QF2N", State: domain.StateOpen, Kind: domain.KindCode, Note: Progress(t, 1, 4), Title: "TUI: human client next to the harness"},
		{ID: "9HCXWA", State: domain.StateDefined, Kind: domain.KindCode, Note: Note(t, "blocked", ToneWarning), Title: "TUI issue views: prep tui, tabs, list, detail pane, live reload"},
		{ID: "MZ3K7D", State: domain.StateInProgress, Kind: domain.KindCode, Title: "TUI design system: tokens, styles, components, gallery"},
		{ID: "VB8R1E", State: domain.StateReady, Kind: domain.KindDecision, Note: Note(t, "stale", ToneError), Title: "Decide the install and update flow", Tags: []string{"release", "install"}, Priority: domain.PriorityHigh},
		{ID: "2YJ6PG", State: domain.StateDone, Kind: domain.KindResearch, Title: "Cloud sessions without the prep binary"},
		{ID: "QK0D5S", State: domain.StateDropped, Kind: domain.KindManual, Title: "Free-form tags"},
	}
	for k, r := range rows {
		b.WriteString(ListRow(t, r, k == 2, width) + "\n")
	}

	section("Tree rows")
	tree := []Row{
		{ID: "T4QF2N", State: domain.StateOpen, Kind: domain.KindCode, Title: "TUI: human client next to the harness", Dimmed: true},
		{ID: "9HCXWA", State: domain.StateDone, Kind: domain.KindCode, Title: "TUI issue views"},
		{ID: "MZ3K7D", State: domain.StateDone, Kind: domain.KindCode, Note: Progress(t, 1, 2), Title: "TUI design system"},
		{ID: "F7NW3H", State: domain.StateOpen, Kind: domain.KindCode, Title: "TUI theming"},
		{ID: "XA4M9C", State: domain.StateInProgress, Kind: domain.KindCode, Title: "Replace the purple accent"},
		{ID: "R2GT6V", State: domain.StateDefined, Kind: domain.KindCode, Note: Note(t, "blocked", ToneWarning), Title: "TUI hierarchy"},
	}
	prefixes := TreePrefixes([]int{0, 1, 1, 2, 2, 1})
	for k, r := range tree {
		r.Tree = prefixes[k]
		b.WriteString(ListRow(t, r, k == 4, width) + "\n")
	}

	section("Link rows")
	b.WriteString(LinkRow(t, "parent", domain.StateOpen, "T4QF2N", "TUI: human client next to the harness", false, width) + "\n")
	b.WriteString(LinkRow(t, "child", domain.StateInProgress, "R2GT6V", "TUI hierarchy: tree mode, parent focus, navigable relations", true, width) + "\n")
	b.WriteString(LinkRow(t, "depends on", domain.StateDone, "9HCXWA", "TUI issue views", false, width) + "\n")
	b.WriteString(LinkRow(t, "blocks", domain.StateDefined, "J8EPK0", "TUI filter bar, stale diff, check output, settings", false, width))

	section("Diff")
	oldReq := []string{"Export rows as CSV.", "", "Quote fields that contain separators."}
	newReq := []string{"Export rows as CSV and JSON.", "", "Quote fields that contain separators.", "", "Write a header line first; long lines wrap with a hanging indent so the sign column stays clear."}
	b.WriteString(Diff(t, DiffLines(oldReq, newReq), width))

	section("Diagnostics")
	b.WriteString(DiagnosticRow(t, Diagnostic{Error: true, Code: "I019", Where: "01K6W3Y8GZ5M0T7C2RNB4QHXDE · decisions.md", Message: "decision D1 has no date", Fix: "entries are '## <id>: <title>' followed by date: YYYY-MM-DD", Class: "manual"}, width) + "\n")
	b.WriteString(DiagnosticRow(t, Diagnostic{Code: "K005", Where: "components/cli.md", Message: "scoped paths changed since 0da33f5: internal/cli/cli.go", Fix: "re-check the entry against the code, update it and confirmed_commit", Class: "guided"}, width))

	section("Menu and modal")
	w := min(width, 64)
	menu := strings.Join([]string{
		MenuRow(t, "n", "New issue", "", true, false, w-4),
		MenuRow(t, "e", "Edit requirement", "", true, true, w-4),
		MenuRow(t, "d", "Define", "the Open questions section is not empty", false, false, w-4),
		MenuRow(t, "f", "Complete", "code issues are completed by an agent", false, false, w-4),
	}, "\n")
	b.WriteString(Modal(t, "Actions · 6DWQ1B TUI editing and transitions", menu, w) + "\n")
	b.WriteString(Modal(t, "Drop QK0D5S", Field(t, "Reason", "› not needed any more", true)+"\n\n"+Field(t, "Kind", "code", false), w))

	section("Suggestions")
	sw := min(width, 64)
	b.WriteString(strings.Join([]string{
		Suggestion(t, "--state", -1, "", false, sw),
		Suggestion(t, "in-progress", 2, "", true, sw),
		Suggestion(t, "01M48KB1NRFQ1A3VWB9SHDM3TM", 8, "Release 0.2.0", false, sw),
	}, "\n"))

	section("Panes")
	left, right := Split(width, 0.5, 20, 20)
	body := "Requirement prose\n" + t.S.Muted.Render("muted second line") + "\nA line long enough to be clipped at the pane's inner width, ending here."
	a := Pane{Title: "Focused pane", Body: body, Focused: true, Width: left, Height: 6}.View(t)
	if right > 0 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, a, Pane{Title: "Unfocused pane", Body: body, Width: right, Height: 6}.View(t)))
	} else {
		b.WriteString(a)
	}

	section("Activity")
	aw := min(width, 72)
	b.WriteString(Chapter(t, "5ZTDH", "Activity stream, prep focus and prep activity", domain.StateInProgress, aw) + "\n")
	for _, a := range []Activity{
		{Area: "issue", Ago: "now", Verb: "checked criterion 3", Target: "Activity stream", Fresh: true},
		{Area: "knowledge", Ago: "12s", Verb: "searched knowledge", Target: "watch debounce"},
		{Area: "code", Ago: "1m", Verb: "edited", Target: "internal/watch/watch.go"},
		{Area: "code", Ago: "2m", Verb: "ran", Target: "go test ./...", Failed: true},
		{Area: "prep", Ago: "4m", Verb: "read the briefing"},
	} {
		b.WriteString(ActivityLine(t, a, aw) + "\n")
	}
	b.WriteString(Celebrate(t, "done in 23 commands", aw) + "\n")
	b.WriteString(Meter(t, "context", 41_200, 200_000, aw) + "\n")
	b.WriteString(Meter(t, "context", 131_000, 200_000, aw) + "\n")
	b.WriteString(Meter(t, "context", 182_500, 200_000, aw) + "\n")
	b.WriteString(BarRow(t, "knowledge", 2400, 9100, "2.4k", aw) + "\n")
	b.WriteString(BarRow(t, "code", 9100, 9100, "9.1k", aw))

	section("Key help")
	b.WriteString(KeyHelp(t, []Key{{"tab", "next view"}, {"↑/↓", "move"}, {"enter", "focus detail"}, {"y", "copy id"}, {"q", "quit"}}, width))

	section("Empty, loading, error")
	b.WriteString(Empty(t, "No issues in this view", "Issues appear here when they match the view's flags", width, 3) + "\n")
	b.WriteString(Loading(t, "Loading .prep …") + "\n")
	b.WriteString(Fit(Error(t, "issue.md: frontmatter is not valid YAML"), width))

	section("Markdown")
	md := "## Requirement\n\nThe TUI watches `.prep` and **reloads** when files change.\n\n- [x] tokens\n- [ ] gallery\n\n> A quoted note.\n\n1. first\n2. second\n"
	if out, err := Markdown(t, md, width); err != nil {
		b.WriteString(Error(t, err.Error()))
	} else {
		b.WriteString(out)
	}
	return b.String()
}

// wrapItems joins rendered items with sep, starting a new line before an
// item that would exceed width.
func wrapItems(items []string, width int, sep string) string {
	var lines []string
	cur := ""
	for _, it := range items {
		next := it
		if cur != "" {
			next = cur + sep + it
		}
		if cur != "" && lipgloss.Width(next) > width {
			lines = append(lines, cur)
			next = it
		}
		cur = next
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}
