// Package theme holds the TUI design tokens and text styles. Screens and
// components take their looks from a Theme and never define colors or
// styles of their own, so visual changes happen here.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
)

// Colors are the resolved tokens for one background and color profile.
type Colors struct {
	Text, Muted, Subtle, Accent, Border, Focus, Selection, Success, Warning, Error color.Color
}

// Styles are the text styles. Components start from these.
type Styles struct {
	Title, Heading, Body, Muted, Subtle, Code, Key, KeyDesc lipgloss.Style
}

// Spacing tokens, in cells.
const (
	Gap     = 1 // between inline elements
	Pad     = 1 // inside panes
	Section = 1 // blank lines between sections
)

// Theme is the design system resolved for one palette and color profile.
type Theme struct {
	Palette palette.Palette
	// Dark is true for themes made for dark terminals (all but light ones).
	Dark    bool
	Profile colorprofile.Profile
	C       Colors
	S       Styles
	Border  lipgloss.Border
}

// New builds the built-in theme for a terminal background when nothing is
// configured: default on dark terminals, light on light ones. Programs
// start dark and follow the terminal's reported background; tests pass
// fixed values so output is deterministic.
func New(dark bool, profile colorprofile.Profile) *Theme {
	p, _ := palette.Resolve(palette.Pick("", dark), nil)
	return From(p, profile)
}

// From builds the theme of a palette for a color profile.
func From(p palette.Palette, profile colorprofile.Profile) *Theme {
	t := &Theme{Palette: p, Dark: !p.Light, Profile: profile, Border: lipgloss.RoundedBorder()}
	t.C = Colors{
		Text: t.color("text"), Muted: t.color("muted"), Subtle: t.color("subtle"), Accent: t.color("accent"),
		Border: t.color("subtle"), Focus: t.color("accent"), Selection: t.color("selection"),
		Success: t.color("success"), Warning: t.color("warning"), Error: t.color("error"),
	}
	st := lipgloss.NewStyle
	t.S = Styles{
		Title:   st().Foreground(t.C.Accent).Bold(true),
		Heading: st().Foreground(t.C.Text).Bold(true),
		Body:    st().Foreground(t.C.Text),
		Muted:   st().Foreground(t.C.Muted),
		Subtle:  st().Foreground(t.C.Subtle),
		Code:    st().Foreground(t.C.Accent),
		Key:     st().Foreground(t.C.Text).Bold(true),
		KeyDesc: st().Foreground(t.C.Muted),
	}
	return t
}

// Rebuild returns the same palette for another color profile.
func (t *Theme) Rebuild(profile colorprofile.Profile) *Theme { return From(t.Palette, profile) }

// color resolves a token for the profile: hand-picked 256- and 16-color
// values when the palette has them, else the hex value downsampled.
func (t *Theme) color(token string) color.Color {
	v, ok := t.Palette.Colors[token]
	if !ok {
		v = t.Palette.Colors["muted"]
	}
	hex := lipgloss.Color(v.Hex)
	ansi256, ansi := colorprofile.ANSI256.Convert(hex), colorprofile.ANSI.Convert(hex)
	if v.ANSI256 != "" {
		ansi256 = lipgloss.Color(v.ANSI256)
	}
	if v.ANSI != "" {
		ansi = lipgloss.Color(v.ANSI)
	}
	return lipgloss.Complete(t.Profile)(ansi, ansi256, hex)
}

// stateTokens draws each state from the palette: the glyph tells states
// apart, the color only marks what needs attention (ready, in progress)
// and lets finished work fade.
var stateTokens = map[domain.State]string{
	domain.StateOpen:       "muted",
	domain.StateDefined:    "text",
	domain.StateReady:      "success",
	domain.StateInProgress: "accent",
	domain.StateDone:       "subtle",
	domain.StateDropped:    "subtle",
}

// State returns the color of an issue state.
func (t *Theme) State(s domain.State) color.Color {
	if tok, ok := stateTokens[s]; ok {
		return t.color(tok)
	}
	return t.C.Muted
}

// Kind returns the color of an issue kind: kinds are written out, so they
// stay muted.
func (t *Theme) Kind(domain.Kind) color.Color { return t.C.Muted }

// Hex returns the true-color value of a token for the theme's background,
// for libraries that take color strings (Glamour).
func (t *Theme) Hex(name string) string { return t.Palette.Colors[name].Hex }

// Token names and their colors, in display order, for the gallery.
func (t *Theme) Tokens() []struct {
	Name  string
	Color color.Color
} {
	var out []struct {
		Name  string
		Color color.Color
	}
	for _, tok := range palette.Tokens() {
		out = append(out, struct {
			Name  string
			Color color.Color
		}{tok, t.color(tok)})
	}
	return out
}
