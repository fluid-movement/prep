// Package theme holds the TUI design tokens and text styles. Screens and
// components take their looks from a Theme and never define colors or
// styles of their own, so visual changes happen here.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/fluid-movement/prep/internal/domain"
)

// value is a color for one background: true color, 256-color and 16-color.
type value struct{ trueColor, ansi256, ansi string }

// token is one color token for dark and light backgrounds.
type token struct{ dark, light value }

func c(darkTrue, dark256, dark16, lightTrue, light256, light16 string) token {
	return token{dark: value{darkTrue, dark256, dark16}, light: value{lightTrue, light256, light16}}
}

// The palette. Change values here; names describe meaning, not hue.
var (
	text      = c("#E4E4EA", "254", "15", "#1E1E26", "235", "0")
	muted     = c("#A0A0AE", "247", "7", "#5C5C6A", "241", "8")
	subtle    = c("#5E5E6C", "240", "8", "#A4A4B2", "248", "7")
	accent    = c("#FF9E5E", "215", "11", "#C2410C", "166", "3")
	border    = c("#3A3A48", "237", "8", "#D2D2DC", "252", "7")
	selection = c("#2A2F3A", "236", "0", "#E8ECF2", "255", "15")
	success   = c("#5BD68A", "78", "10", "#167A3E", "28", "2")
	warning   = c("#F2C14E", "221", "11", "#9A5B00", "130", "3")
	errorC    = c("#FF6B81", "204", "9", "#BE123C", "161", "1")

	stateColors = map[domain.State]token{
		domain.StateOpen:       c("#A0A0AE", "247", "7", "#5C5C6A", "241", "8"),
		domain.StateDefined:    c("#6CB6FF", "75", "12", "#1F5FAD", "25", "4"),
		domain.StateReady:      c("#3DD6C6", "43", "14", "#0B7A70", "30", "6"),
		domain.StateInProgress: c("#F2C14E", "221", "11", "#9A5B00", "130", "3"),
		domain.StateDone:       c("#5BD68A", "78", "10", "#167A3E", "28", "2"),
		domain.StateDropped:    c("#6E6E7C", "242", "8", "#9A9AA8", "247", "7"),
	}
	kindColors = map[domain.Kind]token{
		domain.KindCode:     c("#8AB4F8", "111", "12", "#1D5FBF", "26", "4"),
		domain.KindManual:   c("#F58FC6", "211", "13", "#A3246C", "125", "5"),
		domain.KindResearch: c("#62C7F5", "81", "14", "#0F6A99", "24", "6"),
		domain.KindDecision: c("#E3C58E", "180", "11", "#7A5A12", "94", "3"),
	}
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

// Theme is the design system resolved for one terminal background and
// color profile.
type Theme struct {
	Dark    bool
	Profile colorprofile.Profile
	C       Colors
	S       Styles
	Border  lipgloss.Border
}

// New builds the theme for a background and color profile. Programs start
// with dark and rebuild the theme when the terminal reports its background;
// tests pass fixed values so output is deterministic.
func New(dark bool, profile colorprofile.Profile) *Theme {
	t := &Theme{Dark: dark, Profile: profile, Border: lipgloss.RoundedBorder()}
	t.C = Colors{
		Text: t.color(text), Muted: t.color(muted), Subtle: t.color(subtle), Accent: t.color(accent),
		Border: t.color(border), Focus: t.color(accent), Selection: t.color(selection),
		Success: t.color(success), Warning: t.color(warning), Error: t.color(errorC),
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

func (t *Theme) color(c token) color.Color {
	v := c.light
	if t.Dark {
		v = c.dark
	}
	return lipgloss.Complete(t.Profile)(lipgloss.Color(v.ansi), lipgloss.Color(v.ansi256), lipgloss.Color(v.trueColor))
}

// State returns the color of an issue state.
func (t *Theme) State(s domain.State) color.Color {
	if c, ok := stateColors[s]; ok {
		return t.color(c)
	}
	return t.C.Muted
}

// Kind returns the color of an issue kind.
func (t *Theme) Kind(k domain.Kind) color.Color {
	if c, ok := kindColors[k]; ok {
		return t.color(c)
	}
	return t.C.Muted
}

// Hex returns the true-color value of a token for the theme's background,
// for libraries that take color strings (Glamour).
func (t *Theme) Hex(name string) string {
	m := map[string]token{"text": text, "muted": muted, "subtle": subtle, "accent": accent, "border": border,
		"selection": selection, "success": success, "warning": warning, "error": errorC}
	c, ok := m[name]
	if !ok {
		return ""
	}
	if t.Dark {
		return c.dark.trueColor
	}
	return c.light.trueColor
}

// Token names and their colors, in display order, for the gallery.
func (t *Theme) Tokens() []struct {
	Name  string
	Color color.Color
} {
	return []struct {
		Name  string
		Color color.Color
	}{
		{"text", t.C.Text}, {"muted", t.C.Muted}, {"subtle", t.C.Subtle}, {"accent", t.C.Accent},
		{"border", t.C.Border}, {"selection", t.C.Selection}, {"success", t.C.Success},
		{"warning", t.C.Warning}, {"error", t.C.Error},
	}
}
