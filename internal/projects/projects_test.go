package projects

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".prep"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestTouchListAndPrune(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	if ps, err := List(); err != nil || len(ps) != 0 {
		t.Fatalf("empty: %+v %v", ps, err)
	}
	a, b, gone := project(t), project(t), project(t)
	t0 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	for _, step := range []struct {
		root string
		at   time.Time
	}{{a, t0}, {b, t0.Add(time.Hour)}, {gone, t0.Add(2 * time.Hour)}, {a, t0.Add(3 * time.Hour)}} {
		if err := Touch(step.root, step.at); err != nil {
			t.Fatal(err)
		}
	}
	os.RemoveAll(gone)
	ps, err := List()
	if err != nil || len(ps) != 2 || ps[0].Root != a || ps[1].Root != b {
		t.Fatalf("list: %+v %v", ps, err)
	}
	if ps[0].Name != filepath.Base(a) || !ps[0].LastUsed.Equal(t0.Add(3*time.Hour)) {
		t.Fatalf("first: %+v", ps[0])
	}
	if _, err := os.Stat(filepath.Join(state, "prep", "projects.json")); err != nil {
		t.Fatal(err)
	}
	// The pruned registry no longer holds the missing project.
	all, _ := load()
	if len(all) != 2 {
		t.Fatalf("pruned file: %+v", all)
	}
}

func TestTouchWithinAMinuteWritesNothing(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	a := project(t)
	t0 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	Touch(a, t0)
	Touch(a, t0.Add(30*time.Second))
	ps, _ := List()
	if len(ps) != 1 || !ps[0].LastUsed.Equal(t0) {
		t.Fatalf("%+v", ps)
	}
}

func TestDamagedRegistryIsRebuilt(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	os.MkdirAll(filepath.Join(state, "prep"), 0o755)
	os.WriteFile(filepath.Join(state, "prep", "projects.json"), []byte("{"), 0o644)
	a := project(t)
	if err := Touch(a, time.Now()); err != nil {
		t.Fatal(err)
	}
	if ps, err := List(); err != nil || len(ps) != 1 {
		t.Fatalf("%+v %v", ps, err)
	}
}

// identified makes a project whose project.md declares id and name.
func identified(t *testing.T, id, name string) string {
	t.Helper()
	dir := project(t)
	fm := "---\nschema: 1\nid: " + id + "\nname: " + name + "\n---\n\n# P\n"
	if err := os.WriteFile(filepath.Join(dir, ".prep", "project.md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestNormalizeRemote(t *testing.T) {
	for _, u := range []string{
		"git@github.com:acme/api.git",
		"https://github.com/acme/api.git",
		"https://user@GitHub.com/acme/api",
		"ssh://git@github.com:22/acme/api.git",
		"git+ssh://git@github.com/acme/api/",
	} {
		if got := NormalizeRemote(u); got != "github.com/acme/api" {
			t.Errorf("%s: %q", u, got)
		}
	}
	if got := NormalizeRemote("/srv/git/api.git"); got != "/srv/git/api" {
		t.Errorf("path: %q", got)
	}
	if NormalizeRemote(" ") != "" {
		t.Error("empty")
	}
}

func TestSanitizeName(t *testing.T) {
	for in, want := range map[string]string{"Prep": "prep", "my app": "my-app", "-x": "x", "": "project", "a.b_c-d": "a.b_c-d"} {
		if got := SanitizeName(in); got != want {
			t.Errorf("%q: %q", in, got)
		}
	}
}

func TestResolve(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t0 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	api := identified(t, "01M4AAAAAAAAAAAAAAAAAAAP11", "api")
	apiClone := identified(t, "01M4AAAAAAAAAAAAAAAAAAAP11", "api")
	web := identified(t, "01M4BBBBBBBBBBBBBBBBBBBWEB", "web")
	webFork := identified(t, "01M4CCCCCCCCCCCCCCCCCCCWEB", "web")
	Touch(api, t0)
	Touch(apiClone, t0.Add(time.Hour))
	Touch(web, t0)
	Touch(webFork, t0)

	// Clones of one project resolve to the most recently used.
	if p, err := Resolve("api"); err != nil || p.Root != apiClone {
		t.Fatalf("api: %+v %v", p, err)
	}
	// A shared name is ambiguous; the ID or a suffix of it is not.
	if _, err := Resolve("web"); err == nil || !strings.Contains(err.Error(), "several projects") {
		t.Fatalf("web: %v", err)
	}
	if p, err := Resolve("cccweb"); err != nil || p.Root != webFork {
		t.Fatalf("suffix: %+v %v", p, err)
	}
	if _, err := Resolve("WEB"); err == nil {
		t.Fatal("a suffix of two IDs resolved")
	}
	if p, err := Resolve(web); err != nil || p.Root != web {
		t.Fatalf("path: %+v %v", p, err)
	}
	if _, err := Resolve("nope"); err == nil || !strings.Contains(err.Error(), "registered: api, web") {
		t.Fatalf("unknown: %v", err)
	}
	ps, _ := List()
	for _, p := range ps {
		if p.Collision != (p.Name == "web") {
			t.Errorf("collision: %+v", p)
		}
	}
	if p, ok, _ := ByID("01M4AAAAAAAAAAAAAAAAAAAP11"); !ok || p.Root != apiClone {
		t.Fatalf("by id: %+v", p)
	}
}

func TestAddAndRemove(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t0 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	api := identified(t, "01M4AAAAAAAAAAAAAAAAAAAP11", "api")
	clone := identified(t, "01M4AAAAAAAAAAAAAAAAAAAP11", "api")
	if _, err := Add(t.TempDir(), t0); err == nil {
		t.Fatal("added a directory without .prep")
	}
	if p, err := Add(api, t0); err != nil || p.Name != "api" || p.ID == "" {
		t.Fatalf("add: %+v %v", p, err)
	}
	Add(clone, t0)
	// Removing by path forgets that clone only; by name, every clone.
	if roots, err := Remove(api); err != nil || len(roots) != 1 {
		t.Fatalf("remove path: %v %v", roots, err)
	}
	Add(api, t0)
	if roots, err := Remove("api"); err != nil || len(roots) != 2 {
		t.Fatalf("remove name: %v %v", roots, err)
	}
	if ps, _ := List(); len(ps) != 0 {
		t.Fatalf("left: %+v", ps)
	}
	b, _ := os.ReadFile(filepath.Join(api, ".prep", "project.md"))
	if !strings.Contains(string(b), "name: api") {
		t.Fatal("the project changed")
	}
}
