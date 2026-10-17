package palette

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuiltinsDefineEveryToken(t *testing.T) {
	if got := strings.Join(Tokens(), " "); got != "text muted subtle accent selection success warning error" {
		t.Fatalf("tokens = %s", got)
	}
	want := []string{Default, Light}
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
			if !regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`).MatchString(b.Colors[tok].Hex) {
				t.Errorf("%s %s = %q", b.Name, tok, b.Colors[tok].Hex)
			}
		}
	}
}

func TestPickFollowsTheTerminalOnlyWhenUnset(t *testing.T) {
	if Pick("", true) != Default || Pick("", false) != Light || Pick(Default, false) != Default {
		t.Fatal("Pick")
	}
}

func TestResolveBuiltins(t *testing.T) {
	for _, name := range []string{"", Default, Light} {
		p, err := Resolve(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "" && p.Name != Default {
			t.Fatal("empty selection must use default")
		}
		p.Colors["text"] = Color{Hex: "#000000"}
		again, _ := Resolve(name)
		if again.Colors["text"].Hex == "#000000" {
			t.Fatal("Resolve modified the built-in")
		}
		if err := Validate(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := Validate("removed"); err == nil {
		t.Fatal("unknown theme accepted")
	}
}
