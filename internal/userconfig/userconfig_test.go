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
