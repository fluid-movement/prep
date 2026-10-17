// Package palette names the TUI color tokens and holds the two built-in themes.
// It has no TUI dependencies, so the store can validate the selected theme;
// internal/tui/theme turns a palette into colors and styles.
package palette

import (
	"fmt"
	"slices"
	"strings"
)

// Default and Light are the themes used when the configuration selects
// none, on dark and on light terminals.
const (
	Default = "default"
	Light   = "light"
)

// Pick returns the theme to show: the configured one, else the default for
// the terminal's background.
func Pick(configured string, darkTerminal bool) string {
	switch {
	case configured != "":
		return configured
	case darkTerminal:
		return Default
	}
	return Light
}

// tokens are the palette: a gray ramp, the accent, the selection
// background and the status colors. Everything else (borders, states,
// kinds) is drawn from these.
var tokens = []string{"text", "muted", "subtle", "accent", "selection", "success", "warning", "error"}

// Tokens lists every token name in display order.
func Tokens() []string { return slices.Clone(tokens) }

// Color is one token's color: a true-color hex value, and optionally
// hand-picked 256- and 16-color values. Without them, the theme package
// downsamples the hex value for the terminal's profile.
type Color struct{ Hex, ANSI256, ANSI string }

// Variant maps every token to its color.
type Variant map[string]Color

// Palette holds every token color. Light marks the light-terminal theme.
type Palette struct {
	Name   string
	Light  bool
	Colors Variant
}

// Builtins lists the built-in theme names in display order.
func Builtins() []string {
	names := make([]string, len(builtins))
	for k, b := range builtins {
		names[k] = b.Name
	}
	return names
}

func builtin(name string) (Palette, bool) {
	for _, b := range builtins {
		if b.Name == name {
			return b, true
		}
	}
	return Palette{}, false
}

// Resolve returns a built-in palette. Empty name means the default.
func Resolve(name string) (Palette, error) {
	if name == "" {
		name = Default
	}
	if b, ok := builtin(name); ok {
		return Palette{Name: b.Name, Light: b.Light, Colors: clone(b.Colors)}, nil
	}
	return Palette{}, fmt.Errorf("unknown theme %q; themes: %s", name, strings.Join(Builtins(), ", "))
}

// Validate checks the selected theme.
func Validate(selected string) error {
	_, err := Resolve(selected)
	return err
}

func clone(v Variant) Variant {
	out := make(Variant, len(v))
	for k, c := range v {
		out[k] = c
	}
	return out
}
