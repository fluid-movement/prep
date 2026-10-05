// Package ui is the TUI component library. Components take a theme and
// plain values and return rendered strings; they hold no state and read no
// project data. Screens own state and compose components.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Fixed column widths, so list rows align.
const (
	BadgeWidth = 14 // "▶ in progress" plus a gap
	KindWidth  = 9  // "research" plus a gap
)

var stateGlyphs = map[domain.State]string{
	domain.StateOpen:       "○",
	domain.StateDefined:    "◐",
	domain.StateReady:      "●",
	domain.StateInProgress: "▶",
	domain.StateDone:       "✓",
	domain.StateDropped:    "✕",
}

// StateLabel is the human label of a state.
func StateLabel(s domain.State) string {
	return strings.ReplaceAll(string(s), "_", " ")
}

// StateBadge renders a state as glyph and label in the state's color,
// padded to BadgeWidth.
func StateBadge(t *theme.Theme, s domain.State) string {
	g, ok := stateGlyphs[s]
	if !ok {
		g = "?"
	}
	st := t.R.NewStyle().Foreground(t.State(s)).Width(BadgeWidth)
	if s == domain.StateDropped {
		return st.Render(g + " " + t.R.NewStyle().Strikethrough(true).Render(StateLabel(s)))
	}
	return st.Render(g + " " + StateLabel(s))
}

// KindTag renders an issue kind in its color, padded to KindWidth.
func KindTag(t *theme.Theme, k domain.Kind) string {
	return t.R.NewStyle().Foreground(t.Kind(k)).Width(KindWidth).Render(string(k))
}

// Progress renders resolved children of a parent as "n/m" with a short bar.
func Progress(t *theme.Theme, done, total int) string {
	if total <= 0 {
		return ""
	}
	const cells = 5
	filled := done * cells / total
	bar := t.R.NewStyle().Foreground(t.C.Success).Render(strings.Repeat("■", filled)) +
		t.R.NewStyle().Foreground(t.C.Subtle).Render(strings.Repeat("■", cells-filled))
	return bar + " " + t.S.Muted.Render(fmt.Sprintf("%d/%d", done, total))
}

// Note renders a short status note such as "blocked" or "stale".
func Note(t *theme.Theme, text string, tone Tone) string {
	return t.R.NewStyle().Foreground(tone.color(t)).Render(text)
}

// Tone is the semantic emphasis of a note or message.
type Tone int

const (
	ToneMuted Tone = iota
	ToneSuccess
	ToneWarning
	ToneError
	ToneAccent
)

func (o Tone) color(t *theme.Theme) lipgloss.TerminalColor {
	switch o {
	case ToneSuccess:
		return t.C.Success
	case ToneWarning:
		return t.C.Warning
	case ToneError:
		return t.C.Error
	case ToneAccent:
		return t.C.Accent
	}
	return t.C.Muted
}

// Tab is one entry of a tabs bar. Count < 0 hides the count.
type Tab struct {
	Label string
	Count int
}

// Tabs renders a one-line tabs bar with the active tab highlighted,
// truncated to width.
func Tabs(t *theme.Theme, tabs []Tab, active, width int) string {
	var parts []string
	for k, tb := range tabs {
		label := tb.Label
		if tb.Count >= 0 {
			label += " " + fmt.Sprint(tb.Count)
		}
		if k == active {
			parts = append(parts, t.R.NewStyle().Foreground(t.C.Accent).Background(t.C.Selection).Bold(true).Padding(0, 1).Render(label))
		} else {
			parts = append(parts, t.R.NewStyle().Foreground(t.C.Muted).Padding(0, 1).Render(label))
		}
	}
	return Fit(strings.Join(parts, t.S.Subtle.Render("│")), width)
}

// Row is the data of one issue list row.
type Row struct {
	ID    string
	State domain.State
	Kind  domain.Kind
	Note  string // progress, "blocked", "stale"; pre-rendered by the caller
	Title string
}

// ListRow renders one issue row: marker, ID, state, kind, note, title. The
// selected row gets the selection background across the full width.
func ListRow(t *theme.Theme, r Row, selected bool, width int) string {
	marker := "  "
	if selected {
		marker = t.R.NewStyle().Foreground(t.C.Accent).Render("▌ ")
	}
	id := t.S.Muted.Render(r.ID)
	if selected {
		id = t.R.NewStyle().Foreground(t.C.Accent).Render(r.ID)
	}
	head := marker + id + " " + StateBadge(t, r.State) + KindTag(t, r.Kind)
	if r.Note != "" {
		head += r.Note + " "
	}
	title := t.S.Body.Render(r.Title)
	if selected {
		title = t.S.Heading.Render(r.Title)
	}
	line := Fit(head+title, width)
	if selected {
		return t.R.NewStyle().Background(t.C.Selection).Width(width).Render(line)
	}
	return line
}

// Pane is a bordered box with a title in its top border.
type Pane struct {
	Title   string
	Body    string
	Focused bool
	Width   int // outer width
	Height  int // outer height
}

// View renders the pane. The body is clipped to the inner size.
func (p Pane) View(t *theme.Theme) string {
	if p.Width < 4 || p.Height < 2 {
		return ""
	}
	bc := t.C.Border
	if p.Focused {
		bc = t.C.Focus
	}
	b := t.Border
	edge := t.R.NewStyle().Foreground(bc)
	inner := p.Width - 2

	title := ""
	if p.Title != "" {
		ts := t.S.Muted
		if p.Focused {
			ts = t.S.Title
		}
		title = " " + ts.Render(ansi.Truncate(p.Title, inner-4, "…")) + " "
	}
	fill := inner - 1 - lipgloss.Width(title)
	if fill < 0 {
		fill = 0
	}
	top := edge.Render(b.TopLeft+b.Top) + title + edge.Render(strings.Repeat(b.Top, fill)+b.TopRight)

	bodyW := inner - 2*theme.Pad
	bodyH := p.Height - 2
	lines := strings.Split(p.Body, "\n")
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	var out []string
	out = append(out, top)
	pad := strings.Repeat(" ", theme.Pad)
	for k := 0; k < bodyH; k++ {
		l := ""
		if k < len(lines) {
			l = Fit(lines[k], bodyW)
		}
		l += strings.Repeat(" ", max(0, bodyW-lipgloss.Width(l)))
		out = append(out, edge.Render(b.Left)+pad+l+pad+edge.Render(b.Right))
	}
	out = append(out, edge.Render(b.BottomLeft+strings.Repeat(b.Bottom, inner)+b.BottomRight))
	return strings.Join(out, "\n")
}

// Key is one key binding for the help footer.
type Key struct {
	Keys, Desc string
}

// KeyHelp renders key bindings on one line, truncated to width.
func KeyHelp(t *theme.Theme, keys []Key, width int) string {
	var parts []string
	for _, k := range keys {
		parts = append(parts, t.S.Key.Render(k.Keys)+" "+t.S.KeyDesc.Render(k.Desc))
	}
	return Fit(strings.Join(parts, t.S.Subtle.Render(" · ")), width)
}

// Empty renders a centered empty state with a title and a hint.
func Empty(t *theme.Theme, title, hint string, width, height int) string {
	body := t.S.Muted.Render(title)
	if hint != "" {
		body += "\n" + t.S.Subtle.Render(hint)
	}
	return t.R.Place(width, height, lipgloss.Center, lipgloss.Center, t.R.NewStyle().Align(lipgloss.Center).Render(body))
}

// Loading renders a loading message.
func Loading(t *theme.Theme, msg string) string {
	return t.R.NewStyle().Foreground(t.C.Accent).Render("◌") + " " + t.S.Muted.Render(msg)
}

// Error renders an error message.
func Error(t *theme.Theme, msg string) string {
	return t.R.NewStyle().Foreground(t.C.Error).Bold(true).Render("✕") + " " + t.R.NewStyle().Foreground(t.C.Error).Render(msg)
}

// Fit truncates a rendered line to width cells, keeping escape codes intact.
func Fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "…")
}
