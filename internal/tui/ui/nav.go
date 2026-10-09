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
}

// navGap separates the screens: wide, so the bar reads as a few places,
// not as a row of tabs.
var navGap = strings.Repeat(" ", theme.SpaceL)

// Nav renders the screens bar in the grammar Tabs uses for views: the
// current screen accent and bold, the others muted, plain text.
func Nav(t *theme.Theme, items []NavItem, active int) string {
	var parts []string
	for k, it := range items {
		parts = append(parts, navStyle(t, k == active).Render(it.Label))
	}
	return strings.Join(parts, navGap)
}

// navStyle is how both navigation tiers draw an item.
func navStyle(t *theme.Theme, current bool) lipgloss.Style {
	if current {
		return lipgloss.NewStyle().Foreground(t.C.Accent).Bold(true)
	}
	return lipgloss.NewStyle().Foreground(t.C.Muted)
}

// NavSpans returns where each item of a Nav bar starts and how wide it is,
// in cells from the bar's left edge, for hit-testing.
func NavSpans(items []NavItem) (x, w []int) {
	at := 0
	for _, it := range items {
		iw := ansi.StringWidth(it.Label)
		x, w = append(x, at), append(w, iw)
		at += iw + len(navGap)
	}
	return x, w
}
