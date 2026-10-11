package cli

import (
	"bytes"
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

// runIn runs prep as if started in h's project (PREP_ROOT, no --root), so
// --project can choose another one.
func (h *harness) runIn(args ...string) (int, string) {
	h.t.Helper()
	h.t.Setenv("PREP_ROOT", h.dir)
	var out, errb bytes.Buffer
	code := Main(args, strings.NewReader(""), &out, &errb)
	return code, out.String() + errb.String()
}

func TestProjectFlagRunsCommandsInAnotherProject(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	home := newHarness(t)
	api := newHarness(t)
	var reg struct {
		Projects []struct{ Root, ID, Name string }
	}
	home.jsonOf(&reg, "projects")
	apiID := ""
	for _, p := range reg.Projects {
		if abs, _ := filepath.Abs(api.dir); p.Root == abs {
			apiID = p.ID
		}
	}
	if apiID == "" {
		t.Fatalf("api not registered: %+v", reg)
	}

	// A write by ID suffix lands in api, names it, and stages nothing at home.
	code, out := home.runIn("new", "--project", apiID[20:], "--title", "Rate limits", "--kind", "code", "--by", "agent/1", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var w writeResult
	if err := json.Unmarshal([]byte(out), &w); err != nil || w.Project == nil || w.Project.ID != apiID {
		t.Fatalf("result: %v %s", err, out)
	}
	if st, _ := api.state(w.ID); st != "open" {
		t.Fatalf("issue in api: %q", st)
	}
	if code, _ := home.run("show", w.ID); code == 0 {
		t.Fatal("issue in home")
	}
	// Text output names the project; reads work too.
	if code, out := home.runIn("log", w.ID, "first note", "--project", api.dir); code != 0 || !strings.Contains(out, "in project ") {
		t.Fatalf("log: %d %s", code, out)
	}
	if code, out := home.runIn("show", w.ID, "--project", apiID); code != 0 || !strings.Contains(out, "Rate limits") {
		t.Fatalf("show: %d %s", code, out)
	}

	// api's stream has the write; home's notes it without moving focus.
	focused := false
	for _, f := range api.activity().Focus {
		focused = focused || f.Actor == "agent/1" && f.Issue == w.ID
	}
	if !focused {
		t.Fatal("api focus misses agent/1's write")
	}
	hs := home.activity()
	noted := false
	for _, e := range hs.Events {
		if e.Project != "" && e.Issue == w.ID && strings.Contains(e.Target, " › Rate limits") {
			noted = true
		}
	}
	if !noted || len(hs.Focus) != 0 {
		t.Fatalf("home: %+v", hs)
	}

	// Errors: both flags, an unknown project.
	if code, out := home.run("list", "--project", apiID); code != exitUsage || !strings.Contains(out, "--project and --root") {
		t.Fatalf("both: %d %s", code, out)
	}
	if code, out := home.runIn("list", "--project", "nope"); code == 0 || !strings.Contains(out, "no project \"nope\"") {
		t.Fatalf("unknown: %d %s", code, out)
	}
	// PREP_PROJECT works, and --root overrides it.
	t.Setenv("PREP_PROJECT", apiID)
	if code, out := home.runIn("show", w.ID); code != 0 {
		t.Fatalf("env: %d %s", code, out)
	}
	if code, _ := home.run("show", w.ID); code == 0 {
		t.Fatal("--root did not override PREP_PROJECT")
	}
	t.Setenv("PREP_PROJECT", "")

	// prime at home names api.
	var b primeBrief
	home.jsonOf(&b, "prime")
	if len(b.Projects) != 1 {
		t.Fatalf("prime projects: %v", b.Projects)
	}
}
