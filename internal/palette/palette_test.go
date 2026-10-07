package palette

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestBuiltinsDefineEveryToken(t *testing.T) {
	want := []string{"default", "high-contrast", "monochrome", "pastel", "catppuccin", "nord", "gruvbox"}
	if got := strings.Join(Builtins(), " "); got != strings.Join(want, " ") {
		t.Fatalf("built-ins = %s", got)
	}
	for _, b := range builtins {
		for _, v := range []struct {
			bg   string
			vals Variant
		}{{"dark", b.Dark}, {"light", b.Light}} {
			if len(v.vals) != len(Tokens()) {
				t.Errorf("%s %s: %d tokens, want %d", b.Name, v.bg, len(v.vals), len(Tokens()))
			}
			for _, tok := range Tokens() {
				if !hexRe.MatchString(v.vals[tok].Hex) {
					t.Errorf("%s %s.%s = %q", b.Name, v.bg, tok, v.vals[tok].Hex)
				}
			}
		}
	}
}

func TestResolveCustomThemes(t *testing.T) {
	custom := map[string]domain.ThemeDef{
		"mine":   {Base: "nord", Dark: map[string]string{"accent": "#123456"}},
		"deeper": {Base: "mine", Light: map[string]string{"state.done": "#ABCDEF"}},
		"plain":  {Dark: map[string]string{"text": "#FFFFFF"}},
	}
	p, err := Resolve("deeper", custom)
	if err != nil {
		t.Fatal(err)
	}
	nord, _ := builtin("nord")
	if p.Name != "deeper" || p.Dark["accent"].Hex != "#123456" || p.Light["state.done"].Hex != "#ABCDEF" || p.Dark["text"] != nord.Dark["text"] {
		t.Fatalf("deeper = %+v", p)
	}
	if nord.Dark["accent"].Hex == "#123456" {
		t.Fatal("resolving changed the built-in")
	}
	if p, err := Resolve("plain", custom); err != nil || p.Light["text"].Hex != "#1E1E26" {
		t.Fatalf("a theme without base starts from default: %v %+v", err, p.Light["text"])
	}
	if p, err := Resolve("", nil); err != nil || p.Name != Default {
		t.Fatalf("empty name: %v %s", err, p.Name)
	}
	if got := strings.Join(Names(custom), " "); !strings.HasSuffix(got, "gruvbox deeper mine plain") {
		t.Fatalf("names = %s", got)
	}
	if d := Def(p); len(d.Dark) != len(Tokens()) || d.Dark["text"] != "#ECEFF4" || d.Light["state.done"] != "#ABCDEF" {
		t.Fatalf("Def = %+v", d)
	}
}

func TestValidateRejectsBrokenThemes(t *testing.T) {
	for _, c := range []struct {
		selected string
		custom   map[string]domain.ThemeDef
		want     string
	}{
		{"nope", nil, `unknown theme "nope"`},
		{"", map[string]domain.ThemeDef{"a": {Dark: map[string]string{"accnt": "#000000"}}}, "unknown token dark.accnt"},
		{"", map[string]domain.ThemeDef{"a": {Light: map[string]string{"text": "red"}}}, "light.text must be a color"},
		{"", map[string]domain.ThemeDef{"a": {Base: "b"}}, `unknown base "b"`},
		{"", map[string]domain.ThemeDef{"a": {Base: "b"}, "b": {Base: "a"}}, "base cycle"},
		{"", map[string]domain.ThemeDef{"nord": {}}, "built-in theme cannot be redefined"},
	} {
		err := Validate(c.selected, c.custom)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Validate(%q, %v) = %v, want %q", c.selected, c.custom, err, c.want)
		}
	}
	if err := Validate("catppuccin", nil); err != nil {
		t.Fatal(err)
	}
}
