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
