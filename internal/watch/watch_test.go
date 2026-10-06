package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchReportsChanges(t *testing.T) {
	dir := t.TempDir()
	ch, stop, err := Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	expect := func(what string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("no change reported after %s", what)
		}
	}
	sub := filepath.Join(dir, "issues", "01KDYYYSM08C9CDRA20DXH61HN")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	expect("creating directories")
	time.Sleep(50 * time.Millisecond) // let the watcher add the new directory
	if err := os.WriteFile(filepath.Join(sub, "issue.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	expect("writing a file in a new directory")
}
