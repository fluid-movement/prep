package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

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
		swatches = append(swatches, t.R.NewStyle().Foreground(tok.Color).Render("██")+" "+t.S.Muted.Render(tok.Name))
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

	section("Tabs")
	tabs := []Tab{{"Attention", 1}, {"Actionable", 0}, {"In progress", 2}, {"To enrich", 3}, {"To define", 12}, {"All", 25}}
	b.WriteString(Tabs(t, tabs, 0, width) + "\n")
	b.WriteString(Tabs(t, tabs, 2, width) + "\n")
	b.WriteString(Tabs(t, tabs, 5, min(width, 40)))

	section("List rows")
	rows := []Row{
		{ID: "152616", State: domain.StateOpen, Kind: domain.KindCode, Note: Progress(t, 1, 4), Title: "TUI: human client next to the harness"},
		{ID: "152617", State: domain.StateDefined, Kind: domain.KindCode, Note: Note(t, "blocked", ToneWarning), Title: "TUI issue views: prep tui, tabs, list, detail pane, live reload"},
		{ID: "190543", State: domain.StateInProgress, Kind: domain.KindCode, Title: "TUI design system: tokens, styles, components, gallery"},
		{ID: "152623", State: domain.StateReady, Kind: domain.KindDecision, Note: Note(t, "stale", ToneError), Title: "Decide the install and update flow"},
		{ID: "185825", State: domain.StateDone, Kind: domain.KindResearch, Title: "Cloud sessions without the prep binary"},
		{ID: "152629", State: domain.StateDropped, Kind: domain.KindManual, Title: "Free-form tags"},
	}
	for k, r := range rows {
		b.WriteString(ListRow(t, r, k == 2, width) + "\n")
	}

	section("Panes")
	left, right := Split(width, 0.5, 20, 20)
	body := "Requirement prose\n" + t.S.Muted.Render("muted second line") + "\nA line long enough to be clipped at the pane's inner width, ending here."
	a := Pane{Title: "Focused pane", Body: body, Focused: true, Width: left, Height: 6}.View(t)
	if right > 0 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, a, Pane{Title: "Unfocused pane", Body: body, Width: right, Height: 6}.View(t)))
	} else {
		b.WriteString(a)
	}

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
