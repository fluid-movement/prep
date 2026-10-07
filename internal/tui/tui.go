// Package tui is prep's terminal UI, the human client next to the harness.
// Looks come from the theme and ui packages; screens here own state.
package tui

import (
	"io"
	"os"
	"slices"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// RunGallery shows every component in a scrollable view that reflows on
// resize, without project data. It starts with the configured theme and
// t/T switch through the built-in themes and the config's custom ones, so
// a theme can be judged before choosing it.
func RunGallery(in io.Reader, out io.Writer, cfg domain.Config) error {
	g := newGallery(cfg, theme.New(true, colorprofile.Detect(out, os.Environ())))
	_, err := tea.NewProgram(g, tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}

type gallery struct {
	th     *theme.Theme
	theme  string // the configured theme; empty follows the background
	themes map[string]domain.ThemeDef
	names  []string
	chosen bool // the user switched themes; the background no longer picks
	vp     viewport.Model
	ready  bool
	width  int
}

func newGallery(cfg domain.Config, th *theme.Theme) *gallery {
	g := &gallery{th: th, theme: cfg.Theme, themes: cfg.Themes, names: palette.Names(cfg.Themes)}
	g.show(palette.Pick(cfg.Theme, true))
	return g
}

// show switches to a theme by name and re-renders.
func (g *gallery) show(name string) {
	p, err := palette.Resolve(name, g.themes)
	if err != nil {
		return
	}
	*g.th = *theme.From(p, g.th.Profile)
	if g.ready {
		g.vp.SetContent(ui.Gallery(g.th, g.width-2))
	}
}

// switchTheme shows the next or previous theme.
func (g *gallery) switchTheme(back bool) {
	cur := slices.Index(g.names, g.th.Palette.Name)
	step := 1
	if back {
		step = len(g.names) - 1
	}
	g.chosen = true
	g.show(g.names[(cur+step)%len(g.names)])
}

var galleryKeys = []ui.Key{{Keys: "↑/↓ pgup/pgdn", Desc: "scroll"}, {Keys: "t/T", Desc: "next/previous theme"}, {Keys: "q", Desc: "quit"}}

func (g *gallery) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (g *gallery) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		if !g.chosen {
			g.show(palette.Pick(g.theme, msg.IsDark()))
		}
	case tea.ColorProfileMsg:
		*g.th = *g.th.Rebuild(msg.Profile)
		g.vp.SetContent(ui.Gallery(g.th, g.width-2))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return g, tea.Quit
		case "t", "T":
			g.switchTheme(msg.String() == "T")
			return g, nil
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
	header := g.th.S.Title.Render("prep design system") + "  " + g.th.S.Subtle.Render("gallery · theme ") + g.th.S.Heading.Render(g.th.Palette.Name)
	footer := ui.KeyHelp(g.th, galleryKeys, g.width)
	body := lipgloss.NewStyle().PaddingLeft(1).Render(g.vp.View())
	return header + "\n\n" + body + "\n\n" + footer
}
