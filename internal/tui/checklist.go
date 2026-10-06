package tui

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// ChecklistItem is one choice in a checklist.
type ChecklistItem struct {
	Label   string
	Detail  string // shown muted after the label
	Checked bool
}

// Checklist is a small inline multi-select: arrows move, space toggles,
// enter confirms, esc or q cancels. It renders below the cursor without
// taking over the screen, so it fits an install script's output.
type Checklist struct {
	th        *theme.Theme
	title     string
	items     []ChecklistItem
	cursor    int
	width     int
	done      bool
	cancelled bool
}

// NewChecklist builds a checklist; RunChecklist runs it on a terminal.
func NewChecklist(th *theme.Theme, title string, items []ChecklistItem) *Checklist {
	return &Checklist{th: th, title: title, items: append([]ChecklistItem(nil), items...), width: 80}
}

// RunChecklist asks the user and returns the final items, or ok false when
// the user cancelled.
func RunChecklist(in io.Reader, out io.Writer, title string, items []ChecklistItem) ([]ChecklistItem, bool, error) {
	c := NewChecklist(theme.New(lipgloss.NewRenderer(out)), title, items)
	if _, err := tea.NewProgram(c, tea.WithInput(in), tea.WithOutput(out)).Run(); err != nil {
		return nil, false, err
	}
	return c.items, !c.cancelled, nil
}

func (c *Checklist) Init() tea.Cmd { return nil }

func (c *Checklist) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			c.cursor = clamp(c.cursor-1, 0, len(c.items)-1)
		case "down", "j":
			c.cursor = clamp(c.cursor+1, 0, len(c.items)-1)
		case " ", "x":
			if len(c.items) > 0 {
				c.items[c.cursor].Checked = !c.items[c.cursor].Checked
			}
		case "enter":
			c.done = true
			return c, tea.Quit
		case "esc", "q", "ctrl+c":
			c.done, c.cancelled = true, true
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c *Checklist) View() string {
	if c.done {
		return ""
	}
	t := c.th
	lines := []string{t.S.Title.Render(c.title), ""}
	for k, it := range c.items {
		box := t.S.Subtle.Render("[ ]")
		if it.Checked {
			box = t.R.NewStyle().Foreground(t.C.Accent).Bold(true).Render("[x]")
		}
		marker := "  "
		label := t.S.Body.Render(it.Label)
		if k == c.cursor {
			marker = t.R.NewStyle().Foreground(t.C.Accent).Render("▌ ")
			label = t.S.Heading.Render(it.Label)
		}
		line := marker + box + " " + label
		if it.Detail != "" {
			line += "  " + t.S.Muted.Render(it.Detail)
		}
		lines = append(lines, ui.Fit(line, c.width))
	}
	lines = append(lines, "", ui.KeyHelp(t, []ui.Key{{Keys: "↑↓", Desc: "move"}, {Keys: "space", Desc: "toggle"}, {Keys: "enter", Desc: "apply"}, {Keys: "esc", Desc: "cancel"}}, c.width))
	return strings.Join(lines, "\n") + "\n"
}
