package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Modal renders a focused box sized to its content, for dialogs drawn over
// the body. Callers center it with Center.
func Modal(t *theme.Theme, title, body string, width int) string {
	h := strings.Count(body, "\n") + 1
	return Pane{Title: title, Body: body, Focused: true, Width: width, Height: h + 2}.View(t)
}

// Center places a block in the middle of an area.
func Center(t *theme.Theme, block string, width, height int) string {
	return t.R.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}

// MenuRow renders one action of a menu: key, label and, for an unavailable
// action, the reason. The selected row is highlighted.
func MenuRow(t *theme.Theme, key, label, reason string, enabled, selected bool, width int) string {
	marker := "  "
	if selected {
		marker = t.R.NewStyle().Foreground(t.C.Accent).Render("▌ ")
	}
	keyStyle, labelStyle := t.R.NewStyle().Foreground(t.C.Accent).Bold(true), t.S.Body
	if !enabled {
		keyStyle, labelStyle = t.S.Subtle, t.S.Subtle
	}
	line := marker + keyStyle.Width(3).Render(key) + labelStyle.Width(22).Render(label)
	if reason != "" {
		line += t.S.Muted.Render(reason)
	}
	line = Fit(line, width)
	if selected {
		return t.R.NewStyle().Background(t.C.Selection).Width(width).Render(line)
	}
	return line
}

// Field renders a labeled form field: the label above the rendered input.
func Field(t *theme.Theme, label, input string, focused bool) string {
	l := t.S.Muted.Render(label)
	if focused {
		l = t.S.Title.Render(label)
	}
	return l + "\n" + input
}
