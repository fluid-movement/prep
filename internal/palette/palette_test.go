package palette

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestBuiltinsDefineEveryToken(t *testing.T) {
	if got := strings.Join(Tokens(), " "); got != "text muted subtle accent selection success warning error" {
		t.Fatalf("tokens = %s", got)
	}
	want := []string{"default", "high-contrast", "monochrome", "pastel", "catppuccin", "nord", "gruvbox", "light"}
	if got := strings.Join(Builtins(), " "); got != strings.Join(want, " ") {
		t.Fatalf("built-ins = %s", got)
	}
	for _, b := range builtins {
		if b.Light != (b.Name == Light) {
			t.Errorf("%s: light = %v; light is the only light theme", b.Name, b.Light)
		}
		if len(b.Colors) != len(Tokens()) {
			t.Errorf("%s: %d tokens, want %d", b.Name, len(b.Colors), len(Tokens()))
		}
		for _, tok := range Tokens() {
			if !hexRe.MatchString(b.Colors[tok].Hex) {
				t.Errorf("%s %s = %q", b.Name, tok, b.Colors[tok].Hex)
			}
		}
	}
}

func TestPickFollowsTheTerminalOnlyWhenUnset(t *testing.T) {
	if Pick("", true) != Default || Pick("", false) != Light || Pick("nord", false) != "nord" {
		t.Fatal("Pick")
	}
}

func TestResolveCustomThemes(t *testing.T) {
	custom := map[string]domain.ThemeDef{
		"mine":   {Base: "nord", Colors: map[string]string{"accent": "#123456"}},
		"deeper": {Base: "mine", Colors: map[string]string{"success": "#ABCDEF"}},
		"plain":  {Colors: map[string]string{"text": "#FFFFFF"}},
		"paper":  {Base: Light, Colors: map[string]string{"accent": "#0000AA"}},
	}
	p, err := Resolve("deeper", custom)
	if err != nil {
		t.Fatal(err)
	}
	nord, _ := builtin("nord")
	if p.Name != "deeper" || p.Light || p.Colors["accent"].Hex != "#123456" || p.Colors["success"].Hex != "#ABCDEF" || p.Colors["text"] != nord.Colors["text"] {
		t.Fatalf("deeper = %+v", p)
	}
	if nord.Colors["accent"].Hex == "#123456" {
		t.Fatal("resolving changed the built-in")
	}
	if p, err := Resolve("plain", custom); err != nil || p.Colors["muted"].Hex != "#A0A0AE" || p.Colors["text"].Hex != "#FFFFFF" {
		t.Fatalf("a theme without base starts from default: %v %+v", err, p.Colors)
	}
	paper, err := Resolve("paper", custom)
	if err != nil || !paper.Light {
		t.Fatalf("a theme based on light is light: %v %+v", err, paper)
	}
	if d := Def(paper); d.Base != Light || len(d.Colors) != len(Tokens()) || d.Colors["accent"] != "#0000AA" {
		t.Fatalf("Def(paper) = %+v", d)
	}
	if d := Def(p); d.Base != "" || d.Colors["text"] != "#ECEFF4" || d.Colors["success"] != "#ABCDEF" {
		t.Fatalf("Def(deeper) = %+v", d)
	}
	if p, err := Resolve("", nil); err != nil || p.Name != Default {
		t.Fatalf("empty name: %v %s", err, p.Name)
	}
	if got := strings.Join(Names(custom), " "); !strings.HasSuffix(got, "light deeper mine paper plain") {
		t.Fatalf("names = %s", got)
	}
}

func TestValidateRejectsBrokenThemes(t *testing.T) {
	for _, c := range []struct {
		selected string
		custom   map[string]domain.ThemeDef
		want     string
	}{
		{"nope", nil, `unknown theme "nope"`},
		{"", map[string]domain.ThemeDef{"a": {Colors: map[string]string{"accnt": "#000000"}}}, "unknown token accnt"},
		{"", map[string]domain.ThemeDef{"a": {Colors: map[string]string{"text": "red"}}}, "text must be a color"},
		{"", map[string]domain.ThemeDef{"a": {Colors: map[string]string{"border": "#000000"}}}, "unknown token border"},
		{"", map[string]domain.ThemeDef{"a": {Colors: map[string]string{"dark": "#000000"}}}, "unknown token dark"},
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
