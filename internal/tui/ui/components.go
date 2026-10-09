// Package ui is the TUI component library. Components take a theme and
// plain values and return rendered strings; they hold no state and read no
// project data. Screens own state and compose components.
package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
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
	return strings.ReplaceAll(string(s), "-", " ")
}

// StateBadge renders a state as glyph and label in the state's color,
// padded to BadgeWidth.
func StateBadge(t *theme.Theme, s domain.State) string {
	g, ok := stateGlyphs[s]
	if !ok {
		g = "?"
	}
	fg := lipgloss.NewStyle().Foreground(t.State(s))
	label := fg
	if s == domain.StateDropped {
		// Sibling styles, not nested ones: a styled string inside another
		// style loses the outer color after its reset.
		label = label.Strikethrough(true)
	}
	return lipgloss.NewStyle().Width(BadgeWidth).Render(fg.Render(g+" ") + label.Render(StateLabel(s)))
}

// KindTag renders an issue kind in its color, padded to KindWidth.
func KindTag(t *theme.Theme, k domain.Kind) string {
	return lipgloss.NewStyle().Foreground(t.Kind(k)).Width(KindWidth).Render(string(k))
}

// Progress renders resolved children of a parent as "n/m" with a short bar.
func Progress(t *theme.Theme, done, total int) string {
	if total <= 0 {
		return ""
	}
	const cells = 5
	filled := done * cells / total
	bar := lipgloss.NewStyle().Foreground(t.C.Success).Render(strings.Repeat("■", filled)) +
		lipgloss.NewStyle().Foreground(t.C.Subtle).Render(strings.Repeat("■", cells-filled))
	return bar + " " + t.S.Muted.Render(fmt.Sprintf("%d/%d", done, total))
}

// Note renders a short status note such as "blocked" or "stale".
func Note(t *theme.Theme, text string, tone Tone) string {
	return lipgloss.NewStyle().Foreground(tone.color(t)).Render(text)
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

func (o Tone) color(t *theme.Theme) color.Color {
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

// Tabs renders a screen's views, the second navigation tier, in the
// grammar of Nav: the active view accent and bold, the others muted, plain
// labels SpaceS apart inside their SpaceS padding; truncated to width.
func Tabs(t *theme.Theme, tabs []Tab, active, width int) string {
	var parts []string
	for k, tb := range tabs {
		parts = append(parts, navStyle(t, k == active).Padding(0, theme.SpaceS).Render(tb.label()))
	}
	return Fit(strings.Join(parts, strings.Repeat(" ", theme.SpaceS)), width)
}

func (tb Tab) label() string {
	if tb.Count >= 0 {
		return tb.Label + " " + fmt.Sprint(tb.Count)
	}
	return tb.Label
}

// TabSpans returns where each tab of a Tabs bar starts and how wide it is,
// in cells from the bar's left edge, so callers can tell which tab a
// pointer is on. Tabs cut off by the width are left out.
func TabSpans(tabs []Tab, width int) (x, w []int) {
	at := 0
	for _, tb := range tabs {
		tw := ansi.StringWidth(tb.label()) + 2*theme.SpaceS // the padding
		if at+tw > width {
			break
		}
		x, w = append(x, at), append(w, tw)
		at += tw + theme.SpaceS // the space between tabs
	}
	return x, w
}

// Inner is the width of a pane's content: the pane less its border and its
// padding on both sides.
func Inner(w int) int { return w - 2 - 2*theme.SpaceS }

// Row is the data of one issue list row.
type Row struct {
	ID     string
	State  domain.State
	Kind   domain.Kind
	Note   string // progress, "blocked", "stale"; pre-rendered by the caller
	Title  string
	Tree   string   // tree lines before the title, such as "│  ├─ "; see TreePrefix
	Dimmed bool     // context row: shown for structure, not part of the result
	Tags   []string // shown after the title
	// Priority marks the title unless it is medium (or unset).
	Priority domain.Priority
}

// PriorityMark renders a priority before a title: critical in the error
// tone, high in the warning tone, low subtle; medium has no mark.
func PriorityMark(t *theme.Theme, p domain.Priority) string {
	switch p.Effective() {
	case domain.PriorityCritical:
		return Note(t, "!crit", ToneError) + " "
	case domain.PriorityHigh:
		return Note(t, "!high", ToneWarning) + " "
	case domain.PriorityLow:
		return t.S.Subtle.Render("low") + " "
	}
	return ""
}

// ListRow renders one issue row: marker, ID, state, kind, tree lines, note,
// title. The selected row gets the selection background across the full
// width; a dimmed row renders entirely in the subtle color.
func ListRow(t *theme.Theme, r Row, selected bool, width int) string {
	marker := "  "
	if selected {
		marker = lipgloss.NewStyle().Foreground(t.C.Accent).Render("▌ ")
	}
	tree := t.S.Subtle.Render(r.Tree)
	var head, title string
	switch {
	case r.Dimmed:
		// A context row carries the same information as a result row, all
		// in the subtle tone: the grey says it is only there for structure.
		g := stateGlyphs[r.State]
		plain := fmt.Sprintf("%s %-*s%-*s", r.ID, BadgeWidth, g+" "+StateLabel(r.State), KindWidth, r.Kind)
		head = marker + t.S.Subtle.Render(plain) + tree
		if r.Note != "" {
			head += t.S.Subtle.Render(ansi.Strip(r.Note)) + " "
		}
		title = t.S.Subtle.Render(ansi.Strip(PriorityMark(t, r.Priority)) + r.Title)
	default:
		id := t.S.Muted.Render(r.ID)
		if selected {
			id = lipgloss.NewStyle().Foreground(t.C.Accent).Render(r.ID)
		}
		head = marker + id + " " + StateBadge(t, r.State) + KindTag(t, r.Kind) + tree
		if r.Note != "" {
			head += r.Note + " "
		}
		title = t.S.Body.Render(r.Title)
		if selected {
			title = t.S.Heading.Render(r.Title)
		}
		title = PriorityMark(t, r.Priority) + title
	}
	if len(r.Tags) > 0 {
		title += t.S.Subtle.Render("  #" + strings.Join(r.Tags, " #"))
	}
	line := Fit(head+title, width)
	if selected {
		return lipgloss.NewStyle().Background(t.C.Selection).Width(width).Render(line)
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
	edge := lipgloss.NewStyle().Foreground(bc)
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

	bodyW := inner - 2*theme.SpaceS
	bodyH := p.Height - 2
	lines := strings.Split(p.Body, "\n")
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	var out []string
	out = append(out, top)
	pad := strings.Repeat(" ", theme.SpaceS)
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
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Align(lipgloss.Center).Render(body))
}

// Loading renders a loading message.
func Loading(t *theme.Theme, msg string) string {
	return lipgloss.NewStyle().Foreground(t.C.Accent).Render("◌") + " " + t.S.Muted.Render(msg)
}

// Error renders an error message.
func Error(t *theme.Theme, msg string) string {
	return lipgloss.NewStyle().Foreground(t.C.Error).Bold(true).Render("✕") + " " + lipgloss.NewStyle().Foreground(t.C.Error).Render(msg)
}

// Fit truncates a rendered line to width cells, keeping escape codes intact.
func Fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

// Link is one navigable issue in the detail's relations block. Its Lead
// says how it relates by shape rather than by a label: tree lines for
// children, an arrow for dependencies.
type Link struct {
	Lead     string // pre-rendered: "├─ ", "← needs ", ...
	State    domain.State
	ID       string
	Priority domain.Priority
	Note     string // pre-rendered progress or status, after the ID
	Title    string
	Key      string // shown before the link in link mode: the key that follows it
}

// LinkLine renders a Link with the list's row grammar: state glyph, ID,
// note, priority mark, title. The selected one gets the selection marker
// and background.
func LinkLine(t *theme.Theme, l Link, selected bool, width int) string {
	marker := "  "
	idStyle, titleStyle := t.S.Muted, t.S.Body
	if selected {
		marker = lipgloss.NewStyle().Foreground(t.C.Accent).Render("▌ ")
		idStyle, titleStyle = lipgloss.NewStyle().Foreground(t.C.Accent), t.S.Heading
	}
	if l.Key != "" {
		marker = LinkKey(t, l.Key)
	}
	g, ok := stateGlyphs[l.State]
	if !ok {
		g = "?"
	}
	line := marker + l.Lead + lipgloss.NewStyle().Foreground(t.State(l.State)).Render(g) + " " + idStyle.Render(l.ID) + " "
	if l.Note != "" {
		line += l.Note + " "
	}
	line += PriorityMark(t, l.Priority) + titleStyle.Render(l.Title)
	line = Fit(line, width)
	if selected {
		return lipgloss.NewStyle().Background(t.C.Selection).Width(width).Render(line)
	}
	return line
}

// LinkKey renders the key that follows a link in link mode, two cells wide.
func LinkKey(t *theme.Theme, key string) string {
	return lipgloss.NewStyle().Foreground(t.C.Accent).Bold(true).Render(fmt.Sprintf("%-2s", key))
}

// Selected puts a pre-rendered line on the selection background.
func Selected(t *theme.Theme, line string, width int) string {
	return lipgloss.NewStyle().Background(t.C.Selection).Width(width).Render(Fit(line, width))
}

// LinkWidth is the label column of a link row.
const LinkWidth = 11

// LinkRow renders a compact navigable link to an issue: relation label,
// state glyph, ID and title. Used where a full list row is too wide.
func LinkRow(t *theme.Theme, label string, state domain.State, id, title string, selected bool, width int) string {
	marker := "  "
	idStyle, titleStyle := t.S.Muted, t.S.Body
	if selected {
		marker = lipgloss.NewStyle().Foreground(t.C.Accent).Render("▌ ")
		idStyle, titleStyle = lipgloss.NewStyle().Foreground(t.C.Accent), t.S.Heading
	}
	g, ok := stateGlyphs[state]
	if !ok {
		g = "?"
	}
	line := marker + t.S.Subtle.Width(LinkWidth).Render(label) +
		lipgloss.NewStyle().Foreground(t.State(state)).Render(g) + " " +
		idStyle.Render(id) + " " + titleStyle.Render(title)
	line = Fit(line, width)
	if selected {
		return lipgloss.NewStyle().Background(t.C.Selection).Width(width).Render(line)
	}
	return line
}
