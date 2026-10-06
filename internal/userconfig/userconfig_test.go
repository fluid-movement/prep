package userconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoundTripAndLocation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	c, err := Load()
	if err != nil || len(c.Harnesses) != 0 {
		t.Fatalf("missing file: %+v %v", c, err)
	}
	p, err := Save(Config{Harnesses: []string{"claude-code"}})
	if err != nil || p != filepath.Join(dir, "prep", "config.yaml") {
		t.Fatalf("save: %s %v", p, err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "harnesses:\n  - claude-code\n") {
		t.Fatalf("file:\n%s", b)
	}
	c, err = Load()
	if err != nil || len(c.Harnesses) != 1 || c.Harnesses[0] != "claude-code" {
		t.Fatalf("load: %+v %v", c, err)
	}
	os.WriteFile(p, []byte("harnesses: [a]\ntheme: dark\n"), 0o644)
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "theme") {
		t.Fatalf("unknown keys should be rejected: %v", err)
	}
}

func TestDefaultLocation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	if p, _ := Path(); p != filepath.Join(home, ".config", "prep", "config.yaml") {
		t.Fatalf("path = %s", p)
	}
}

func TestTUIMouseChoice(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	off := false
	p, err := Save(Config{Harnesses: []string{"claude-code"}, TUI: TUI{Mouse: &off}})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "tui:\n  mouse: false\n") {
		t.Fatalf("file:\n%s", b)
	}
	c, err := Load()
	if err != nil || c.TUI.Mouse == nil || *c.TUI.Mouse || len(c.Harnesses) != 1 {
		t.Fatalf("load: %+v %v", c, err)
	}
	// Unset stays out of the file, so the default (on) can change later.
	if _, err := Save(Config{}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); strings.Contains(string(b), "tui") {
		t.Fatalf("unset mouse written:\n%s", b)
	}
}
