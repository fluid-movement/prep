package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain points the project registry at a temporary state directory, so
// the commands the tests run never record into the real one.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "prep-state-")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_STATE_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestProjectsRecordsTheProjectsCommandsRunIn(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	h := newHarness(t)
	var got struct {
		Projects []struct {
			Root, Name string
		}
	}
	// projects needs no project: run it from outside one.
	var out strings.Builder
	if code := Main([]string{"projects", "--json", "--root", t.TempDir()}, strings.NewReader(""), &out, &out); code != 0 {
		t.Fatalf("exit %d: %s", code, out.String())
	}
	if err := json.Unmarshal([]byte(out.String()), &got); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(h.dir)
	if len(got.Projects) != 1 || got.Projects[0].Root != abs || got.Projects[0].Name != filepath.Base(abs) {
		t.Fatalf("%+v", got)
	}
	if text := h.ok("projects"); !strings.Contains(text, abs) {
		t.Fatalf("text: %s", text)
	}
	os.RemoveAll(filepath.Join(h.dir, ".prep"))
	if text := h.ok("projects"); strings.Contains(text, abs) {
		t.Fatalf("removed project still listed: %s", text)
	}
}

func TestInitWritesIdentityAndFixRestoresIt(t *testing.T) {
	h := newHarness(t)
	pm := filepath.Join(h.dir, ".prep", "project.md")
	read := func() string { b, _ := os.ReadFile(pm); return string(b) }
	first := read()
	if !strings.Contains(first, "\nid: 01") || !strings.Contains(first, "\nname: ") {
		t.Fatalf("init: %s", first)
	}
	// A project from before IDs: check warns, fix writes them back.
	os.WriteFile(pm, []byte("---\nschema: 1\n---\n\n# Project\n"), 0o644)
	if out := h.ok("check"); !strings.Contains(out, "P008") {
		t.Fatalf("check: %s", out)
	}
	h.ok("fix")
	fixed := read()
	if !strings.Contains(fixed, "\nid: 01") || !strings.Contains(fixed, "\nname: ") || !strings.Contains(fixed, "# Project") {
		t.Fatalf("fix: %s", fixed)
	}
	// --new-id replaces only the ID.
	h.ok("init", "--new-id")
	renewed := read()
	if renewed == fixed || strings.Replace(renewed, idLine(renewed), idLine(fixed), 1) != fixed {
		t.Fatalf("new-id:\n%s\nwas:\n%s", renewed, fixed)
	}
}

func idLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, "id: ") {
			return l
		}
	}
	return ""
}

func TestProjectsShowsIdentityAndAddsAndRemoves(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	h := newHarness(t)
	other := newHarness(t)
	h.ok("projects", "remove", other.dir)
	var got struct {
		Projects []struct{ Root, ID, Name, Remote string }
	}
	if err := json.Unmarshal([]byte(h.ok("projects", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Projects) != 1 || len(got.Projects[0].ID) != 26 || got.Projects[0].Name == "" {
		t.Fatalf("%+v", got)
	}
	h.ok("projects", "add", other.dir)
	if out := h.ok("projects"); !strings.Contains(out, other.dir) {
		t.Fatalf("added: %s", out)
	}
	h.fails("no .prep/project.md", "projects", "add", t.TempDir())
}
