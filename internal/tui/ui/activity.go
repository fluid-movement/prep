package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Activity is one line of an agent's activity feed.
type Activity struct {
	Area   string // issue, knowledge, code, prep, other
	Ago    string // relative time, such as "4m"
	Verb   string
	Target string
	Failed bool
	Fresh  bool // just arrived: marked until it settles
}

// areaGlyphs mark what an activity touched.
var areaGlyphs = map[string]string{
	"issue":     "◆",
	"knowledge": "✦",
	"code":      "›",
	"prep":      "▸",
}

// AreaGlyph returns the glyph of an activity area in its tone.
func AreaGlyph(t *theme.Theme, area string, failed bool) string {
	g, ok := areaGlyphs[area]
	if !ok {
		g = "·"
	}
	c := t.C.Muted
	switch {
	case failed:
		g, c = "✗", t.C.Error
	case area == "issue":
		c = t.C.Accent
	case area == "knowledge":
		c = t.C.Success
	case area == "code":
		c = t.C.Text
	}
	return lipgloss.NewStyle().Foreground(c).Render(g)
}

// ActivityLine renders one feed line: a fresh marker, the area glyph, how
// long ago, what happened and to what.
func ActivityLine(t *theme.Theme, a Activity, width int) string {
	marker := "  "
	verb := t.S.Body
	if a.Fresh {
		marker = lipgloss.NewStyle().Foreground(t.C.Accent).Render("▌ ")
		verb = t.S.Heading
	}
	line := marker + AreaGlyph(t, a.Area, a.Failed) + " " + t.S.Subtle.Render(fmt.Sprintf("%-4s", a.Ago)) + " " + verb.Render(a.Verb)
	if a.Target != "" {
		line += t.S.Muted.Render(" · " + a.Target)
	}
	return Fit(line, width)
}

// Chapter heads the feed's lines for one issue: state glyph, ID, title and
// a rule to the edge.
func Chapter(t *theme.Theme, id, title string, s domain.State, width int) string {
	g, ok := stateGlyphs[s]
	if !ok {
		g = "·"
	}
	head := lipgloss.NewStyle().Foreground(t.State(s)).Render(g) + " " + t.S.Subtle.Render(id) + " " + t.S.Heading.Render(title) + " "
	if rest := width - lipgloss.Width(head); rest > 0 {
		head += t.S.Subtle.Render(strings.Repeat("─", rest))
	}
	return Fit(head, width)
}

// Celebrate renders a line marking something finished.
func Celebrate(t *theme.Theme, text string, width int) string {
	return Fit(lipgloss.NewStyle().Foreground(t.C.Success).Bold(true).Render("✓ "+text), width)
}

// Meter renders how full something is: a label, a bar that warns as it
// fills, and the figures.
func Meter(t *theme.Theme, label string, used, total int, width int) string {
	if total <= 0 {
		return Fit(t.S.Muted.Render(fmt.Sprintf("%-10s", label))+t.S.Body.Render(Count(used)), width)
	}
	frac := float64(used) / float64(total)
	c := t.C.Success
	switch {
	case frac >= 0.85:
		c = t.C.Error
	case frac >= 0.6:
		c = t.C.Warning
	}
	figures := fmt.Sprintf(" %s/%s %d%%", Count(used), Count(total), int(frac*100+0.5))
	cells := max(4, min(20, width-10-len(figures)))
	filled := min(cells, int(frac*float64(cells)+0.5))
	bar := lipgloss.NewStyle().Foreground(c).Render(strings.Repeat("■", filled)) + t.S.Subtle.Render(strings.Repeat("■", cells-filled))
	return Fit(t.S.Muted.Render(fmt.Sprintf("%-10s", label))+bar+t.S.Muted.Render(figures), width)
}

// BarRow renders one value of a set as a label, a bar against the largest
// value and its figure.
func BarRow(t *theme.Theme, label string, value, largest int, figure string, width int) string {
	cells := max(4, min(16, width-10-len(figure)-1))
	filled := 0
	if largest > 0 {
		filled = max(1, value*cells/largest)
		if value == 0 {
			filled = 0
		}
	}
	bar := lipgloss.NewStyle().Foreground(t.C.Accent).Render(strings.Repeat("■", filled)) + t.S.Subtle.Render(strings.Repeat("·", cells-filled))
	return Fit(t.S.Muted.Render(fmt.Sprintf("%-10s", label))+bar+" "+t.S.Body.Render(figure), width)
}

// Count renders a number compactly: 950, 12.3k, 1.2M.
func Count(n int) string {
	switch {
	case n >= 1_000_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1_000_000)) + "M"
	case n >= 10_000:
		return fmt.Sprintf("%dk", (n+500)/1000)
	case n >= 1000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1000)) + "k"
	}
	return fmt.Sprint(n)
}

func trimZero(s string) string { return strings.TrimSuffix(s, ".0") }
