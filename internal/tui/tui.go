// Package tui is prep's terminal UI, the human client next to the harness.
// Looks come from the theme and ui packages; screens here own state.
package tui

import (
	"io"
	"os"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// RunGallery shows every component in a scrollable view that reflows on
// resize, without project data.
func RunGallery(in io.Reader, out io.Writer) error {
	th := theme.New(true, colorprofile.Detect(out, os.Environ()))
	_, err := tea.NewProgram(&gallery{th: th}, tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}

type gallery struct {
	th    *theme.Theme
	vp    viewport.Model
	ready bool
	width int
}

var galleryKeys = []ui.Key{{Keys: "↑/↓ pgup/pgdn", Desc: "scroll"}, {Keys: "q", Desc: "quit"}}

func (g *gallery) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (g *gallery) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		*g.th = *theme.New(msg.IsDark(), g.th.Profile)
		g.vp.SetContent(ui.Gallery(g.th, g.width-2))
	case tea.ColorProfileMsg:
		*g.th = *theme.New(g.th.Dark, msg.Profile)
		g.vp.SetContent(ui.Gallery(g.th, g.width-2))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return g, tea.Quit
		}
	case tea.WindowSizeMsg:
		g.width = msg.Width
		h := ui.Stack(msg.Height, 2, 2)
		if !g.ready {
			g.vp = newViewport()
			g.ready = true
		}
		g.vp.SetWidth(msg.Width)
		g.vp.SetHeight(h)
		g.vp.SetContent(ui.Gallery(g.th, msg.Width-2))
	}
	var cmd tea.Cmd
	g.vp, cmd = g.vp.Update(msg)
	return g, cmd
}

func (g *gallery) View() tea.View {
	v := tea.NewView(g.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (g *gallery) render() string {
	if !g.ready {
		return ui.Loading(g.th, "Rendering gallery …")
	}
	header := g.th.S.Title.Render("prep design system") + "  " + g.th.S.Subtle.Render("gallery")
	footer := ui.KeyHelp(g.th, galleryKeys, g.width)
	body := lipgloss.NewStyle().PaddingLeft(1).Render(g.vp.View())
	return header + "\n\n" + body + "\n\n" + footer
}
