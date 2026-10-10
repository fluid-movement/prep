package projects

import (
	"os"
	"path/filepath"
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
