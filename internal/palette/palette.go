// Package palette names the TUI's color tokens, holds the built-in themes
// and resolves custom themes from the configuration. It has no TUI
// dependencies, so the store and the CLI can validate and print themes;
// internal/tui/theme turns a palette into colors and styles.
package palette

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
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

// Palette is a resolved theme: every token's color. Themes are made for
// dark terminals; Light marks the ones made for light terminals (the light
// built-in and custom themes based on it).
type Palette struct {
	Name   string
	Light  bool
	Colors Variant
}

var hexRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

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

// Names lists the built-in themes, then the custom ones by name.
func Names(custom map[string]domain.ThemeDef) []string {
	names := Builtins()
	var own []string
	for n := range custom {
		if _, ok := builtin(n); !ok {
			own = append(own, n)
		}
	}
	sort.Strings(own)
	return append(names, own...)
}

// IsBuiltin reports whether name is a built-in theme.
func IsBuiltin(name string) bool {
	_, ok := builtin(name)
	return ok
}

// Resolve returns the full palette of a theme: a built-in, or a custom
// theme with its base chain applied (a custom theme without base starts
// from the default theme). Empty name means the default.
func Resolve(name string, custom map[string]domain.ThemeDef) (Palette, error) {
	if name == "" {
		name = Default
	}
	return resolve(name, custom, nil)
}

func resolve(name string, custom map[string]domain.ThemeDef, seen []string) (Palette, error) {
	if b, ok := builtin(name); ok {
		return Palette{Name: b.Name, Light: b.Light, Colors: clone(b.Colors)}, nil
	}
	def, ok := custom[name]
	if !ok {
		if len(seen) > 0 {
			return Palette{}, fmt.Errorf("theme %q: unknown base %q", seen[len(seen)-1], name)
		}
		return Palette{}, fmt.Errorf("unknown theme %q; themes: %s", name, strings.Join(Names(custom), ", "))
	}
	if slices.Contains(seen, name) {
		return Palette{}, fmt.Errorf("theme %q: base cycle %s", name, strings.Join(append(seen, name), " → "))
	}
	base := def.Base
	if base == "" {
		base = Default
	}
	p, err := resolve(base, custom, append(seen, name))
	if err != nil {
		return Palette{}, err
	}
	p.Name = name
	for tok, hex := range def.Colors {
		if !slices.Contains(Tokens(), tok) {
			return Palette{}, fmt.Errorf("theme %q: unknown token %s; tokens: base, %s", name, tok, strings.Join(Tokens(), ", "))
		}
		if !hexRe.MatchString(hex) {
			return Palette{}, fmt.Errorf("theme %q: %s must be a color like #A1B2C3, got %q", name, tok, hex)
		}
		p.Colors[tok] = Color{Hex: hex}
	}
	return p, nil
}

// Validate checks every custom theme and the selected one.
func Validate(selected string, custom map[string]domain.ThemeDef) error {
	for _, n := range Names(custom) {
		if IsBuiltin(n) {
			continue
		}
		if _, err := Resolve(n, custom); err != nil {
			return err
		}
	}
	for n := range custom {
		if IsBuiltin(n) {
			return fmt.Errorf("theme %q: a built-in theme cannot be redefined; use base: %s under another name", n, n)
		}
	}
	_, err := Resolve(selected, custom)
	return err
}

// Def returns a custom theme definition holding every token of p, as prep
// theme new writes it. A light palette keeps light as its base, so the
// new theme stays a light theme.
func Def(p Palette) domain.ThemeDef {
	d := domain.ThemeDef{Colors: map[string]string{}}
	if p.Light {
		d.Base = Light
	}
	for _, tok := range Tokens() {
		d.Colors[tok] = p.Colors[tok].Hex
	}
	return d
}

func clone(v Variant) Variant {
	out := make(Variant, len(v))
	for k, c := range v {
		out[k] = c
	}
	return out
}

// hex builds a variant from token → hex pairs in Tokens order.
func hex(values ...string) Variant {
	toks := Tokens()
	if len(values) != len(toks) {
		panic(fmt.Sprintf("palette: %d values for %d tokens", len(values), len(toks)))
	}
	v := make(Variant, len(toks))
	for k, t := range toks {
		v[t] = Color{Hex: values[k]}
	}
	return v
}
