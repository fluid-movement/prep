package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Modal renders a focused box sized to its content, for dialogs drawn over
// the body. Callers center it with Center.
func Modal(t *theme.Theme, title, body string, width int) string {
	h := strings.Count(body, "\n") + 1
	return Pane{Title: title, Body: body, Focused: true, Width: width, Height: h + 2}.View(t)
}

// Overlay draws a block centered over a background of width×height cells.
// The background keeps its text but loses its colors and is drawn in the
// subtle tone, so a dialog stands out while the screen behind stays readable.
// Both are Lip Gloss layers; the block sits above the background.
func Overlay(t *theme.Theme, background, block string, width, height int) string {
	bg := strings.Split(ansi.Strip(background), "\n")
	for len(bg) < height {
		bg = append(bg, "")
	}
	for y := range bg[:height] {
		bg[y] = t.S.Subtle.Render(ansi.Truncate(bg[y], width, "") + strings.Repeat(" ", max(0, width-ansi.StringWidth(bg[y]))))
	}
	top := max(0, (height-lipgloss.Height(block))/2)
	left := max(0, (width-lipgloss.Width(block))/2)
	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(strings.Join(bg[:height], "\n")),
		lipgloss.NewLayer(block).X(left).Y(top).Z(1),
	)
	return lipgloss.NewCanvas(width, height).Compose(layers).Render()
}

// Center places a block in the middle of an area.
func Center(t *theme.Theme, block string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}

// MenuRow renders one action of a menu: key, label and, for an unavailable
// action, the reason. The selected row is highlighted.
func MenuRow(t *theme.Theme, key, label, reason string, enabled, selected bool, width int) string {
	marker := "  "
	if selected {
		marker = lipgloss.NewStyle().Foreground(t.C.Accent).Render("▌ ")
	}
	keyStyle, labelStyle := lipgloss.NewStyle().Foreground(t.C.Accent).Bold(true), t.S.Body
	if !enabled {
		keyStyle, labelStyle = t.S.Subtle, t.S.Subtle
	}
	// Short labels pad to a column so reasons line up; long ones (a linked
	// issue's title) run on and are cut at the edge, never wrapped.
	const labelW = 22
	line := marker + keyStyle.Width(3).Render(key) + labelStyle.Render(label)
	if w := ansi.StringWidth(label); w < labelW {
		line += strings.Repeat(" ", labelW-w)
	} else if reason != "" {
		line += "  "
	}
	if reason != "" {
		line += t.S.Muted.Render(reason)
	}
	line = Fit(line, width)
	if selected {
		return lipgloss.NewStyle().Background(t.C.Selection).Width(width).Render(line)
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
