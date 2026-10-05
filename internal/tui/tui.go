// Package tui is prep's terminal UI, the human client next to the harness.
// Looks come from the theme and ui packages; screens here own state.
package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// RunGallery shows every component in a scrollable view that reflows on
// resize, without project data.
func RunGallery(in io.Reader, out io.Writer) error {
	th := theme.New(lipgloss.NewRenderer(out))
	_, err := tea.NewProgram(&gallery{th: th}, tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}

type gallery struct {
	th    *theme.Theme
	vp    viewport.Model
	ready bool
	width int
}

var galleryKeys = []ui.Key{{Keys: "↑/↓ pgup/pgdn", Desc: "scroll"}, {Keys: "q", Desc: "quit"}}

func (g *gallery) Init() tea.Cmd { return nil }

func (g *gallery) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return g, tea.Quit
		}
	case tea.WindowSizeMsg:
		g.width = msg.Width
		h := ui.Stack(msg.Height, 2, 2)
		if !g.ready {
			g.vp = viewport.New(msg.Width, h)
			g.ready = true
		} else {
			g.vp.Width, g.vp.Height = msg.Width, h
		}
		g.vp.SetContent(ui.Gallery(g.th, msg.Width-2))
	}
	var cmd tea.Cmd
	g.vp, cmd = g.vp.Update(msg)
	return g, cmd
}

func (g *gallery) View() string {
	if !g.ready {
		return ui.Loading(g.th, "Rendering gallery …")
	}
	header := g.th.S.Title.Render("prep design system") + "  " + g.th.S.Subtle.Render("gallery")
	footer := ui.KeyHelp(g.th, galleryKeys, g.width)
	body := lipgloss.NewStyle().PaddingLeft(1).Render(g.vp.View())
	return header + "\n\n" + body + "\n\n" + footer
}
