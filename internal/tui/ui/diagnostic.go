package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Diagnostic is the data of one check finding.
type Diagnostic struct {
	Error   bool   // severity: error, else warning
	Code    string // stable code such as I019
	Where   string // issue or entry and file
	Message string
	Fix     string
	Class   string // fixable, guided, manual
}

// DiagnosticRow renders a finding over several lines: severity glyph, code
// and location, then the message and the fix, wrapped to width.
func DiagnosticRow(t *theme.Theme, d Diagnostic, width int) string {
	glyph, tone := "▲", lipgloss.NewStyle().Foreground(t.C.Warning)
	if d.Error {
		glyph, tone = "✕", lipgloss.NewStyle().Foreground(t.C.Error)
	}
	head := tone.Bold(true).Render(glyph+" "+d.Code) + "  " + t.S.Muted.Render(d.Where)
	lines := []string{Fit(head, width)}
	wrap := func(s string, st func(string) string) {
		for _, l := range strings.Split(lipgloss.NewStyle().Width(max(1, width-2)).Render(s), "\n") {
			lines = append(lines, "  "+st(strings.TrimRight(l, " ")))
		}
	}
	wrap(d.Message, func(s string) string { return t.S.Body.Render(s) })
	if d.Fix != "" {
		fix := "fix: " + d.Fix
		if d.Class != "" {
			fix = d.Class + " · " + fix
		}
		wrap(fix, func(s string) string { return t.S.Subtle.Render(s) })
	}
	return strings.Join(lines, "\n")
}
