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
	expect := func(what string) Change {
		t.Helper()
		select {
		case c := <-ch:
			return c
		case <-time.After(2 * time.Second):
			t.Fatalf("no change reported after %s", what)
		}
		return Change{}
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
	if c := expect("writing a file in a new directory"); !c.Tree || c.Local {
		t.Fatalf("issue write: %+v", c)
	}
	if err := os.MkdirAll(filepath.Join(dir, Local), 0o755); err != nil {
		t.Fatal(err)
	}
	expect("creating the local directory")
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, Local, "activity.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c := expect("appending activity"); c.Tree || !c.Local {
		t.Fatalf("activity write: %+v", c)
	}
}
