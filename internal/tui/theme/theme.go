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

// Theme is the design system resolved for one palette, terminal
// background and color profile.
type Theme struct {
	Palette palette.Palette
	Dark    bool
	Profile colorprofile.Profile
	C       Colors
	S       Styles
	Border  lipgloss.Border
}

// New builds the default theme for a background and color profile.
// Programs start with dark and rebuild the theme when the terminal reports
// its background; tests pass fixed values so output is deterministic.
func New(dark bool, profile colorprofile.Profile) *Theme {
	p, _ := palette.Resolve(palette.Default, nil)
	return From(p, dark, profile)
}

// From builds the theme of a palette for a background and color profile.
func From(p palette.Palette, dark bool, profile colorprofile.Profile) *Theme {
	t := &Theme{Palette: p, Dark: dark, Profile: profile, Border: lipgloss.RoundedBorder()}
	t.C = Colors{
		Text: t.color("text"), Muted: t.color("muted"), Subtle: t.color("subtle"), Accent: t.color("accent"),
		Border: t.color("border"), Focus: t.color("accent"), Selection: t.color("selection"),
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

// variant is the palette's tokens for the theme's background.
func (t *Theme) variant() palette.Variant {
	if t.Dark {
		return t.Palette.Dark
	}
	return t.Palette.Light
}

// Rebuild returns the same palette for another background or profile.
func (t *Theme) Rebuild(dark bool, profile colorprofile.Profile) *Theme {
	return From(t.Palette, dark, profile)
}

// color resolves a token for the profile: hand-picked 256- and 16-color
// values when the palette has them, else the hex value downsampled.
func (t *Theme) color(token string) color.Color {
	v, ok := t.variant()[token]
	if !ok {
		v = t.variant()["muted"]
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

// State returns the color of an issue state.
func (t *Theme) State(s domain.State) color.Color { return t.color("state." + string(s)) }

// Kind returns the color of an issue kind.
func (t *Theme) Kind(k domain.Kind) color.Color { return t.color("kind." + string(k)) }

// Hex returns the true-color value of a token for the theme's background,
// for libraries that take color strings (Glamour).
func (t *Theme) Hex(name string) string { return t.variant()[name].Hex }

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
