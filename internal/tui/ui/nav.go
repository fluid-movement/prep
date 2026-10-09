package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// NavItem is one screen in the navigation bar, the first tier above a
// screen's views (Tabs).
type NavItem struct {
	Label string
	// Dot puts a status dot before the label in its tone, such as an
	// agent that is active (accent) or idle (subtle); nil shows none.
	Dot *Tone
}

// navGap separates the screens: wide, so the bar reads as a few places,
// not as a row of tabs.
var navGap = strings.Repeat(" ", theme.SpaceL)

// Nav renders the screens bar: the current screen bold and underlined in
// the accent, the others muted, so it never reads as a row of view tabs.
func Nav(t *theme.Theme, items []NavItem, active int) string {
	var parts []string
	for k, it := range items {
		style := lipgloss.NewStyle().Foreground(t.C.Muted)
		if k == active {
			style = lipgloss.NewStyle().Foreground(t.C.Accent).Bold(true).Underline(true)
		}
		label := style.Render(it.Label)
		if it.Dot != nil {
			label = Note(t, "●", *it.Dot) + " " + label
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, navGap)
}

// NavSpans returns where each item of a Nav bar starts and how wide it is,
// in cells from the bar's left edge, for hit-testing.
func NavSpans(items []NavItem) (x, w []int) {
	at := 0
	for _, it := range items {
		iw := ansi.StringWidth(it.Label)
		if it.Dot != nil {
			iw += 2
		}
		x, w = append(x, at), append(w, iw)
		at += iw + len(navGap)
	}
	return x, w
}
