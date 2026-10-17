package mdstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestHeldLockSpansLoadAndWriteAndBlocksSecondWriter(t *testing.T) {
	root := t.TempDir()
	id := transactionIssue(t, Open(root))
	first, second := Open(root), Open(root)
	release, err := first.Hold()
	if err != nil {
		t.Fatal(err)
	}
	// Load and Apply must not take the lock again while it is held.
	if _, _, _, err := first.Load(); err != nil {
		t.Fatal(err)
	}
	started, done := make(chan struct{}), make(chan error, 1)
	var seen string
	go func() {
		done <- second.Locked(func() error {
			close(started)
			_, issues, _, err := second.Load()
			if err != nil {
				return err
			}
			for _, i := range issues {
				if i.ID == id {
					seen = i.Context
				}
			}
			return nil
		})
	}()
	select {
	case <-started:
		t.Fatal("second writer got the lock while the first still held it")
	case <-time.After(150 * time.Millisecond):
	}
	ctx := "First writer's context."
	if _, err := first.Apply(&domain.Change{IssueID: id, Context: &ctx}); err != nil {
		t.Fatal(err)
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seen, ctx) {
		t.Fatalf("second writer planned on %q, not on the first writer's result", seen)
	}
}

func TestApplyWritesStateDefiningFileLast(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	ctx := "Done."
	at := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	touched, err := s.Apply(&domain.Change{
		IssueID:     id,
		Resolution:  &domain.Resolution{Outcome: "done", By: "human:test", At: at, Note: "ok"},
		RemoveClaim: true,
		History:     "- completed",
		Context:     &ctx,
	})
	if err != nil {
		t.Fatal(err)
	}
	dir := IssueDir(id)
	// resolution.md is written before claim.md is removed: an interruption
	// between them leaves the issue done, not back in ready.
	want := []string{dir + "/context.md", dir + "/history.md", dir + "/resolution.md", dir + "/claim.md"}
	if strings.Join(touched, ",") != strings.Join(want, ",") {
		t.Fatalf("write order = %v, want %v", touched, want)
	}
	if _, err := os.Stat(filepath.Join(root, dir, "resolution.md")); err != nil {
		t.Fatal(err)
	}
}

func TestOrderForWriteKeepsStateFilesLast(t *testing.T) {
	body := "x"
	touched := []string{"d/resolution.md", "d/baselines/b.md", "d/context.md", "d/ready.md", "d/history.md", "d/claim.md"}
	contents := []*string{nil, &body, &body, &body, &body, &body}
	orderForWrite(touched, contents)
	// Writes by rank (supporting records, then baselines, ready and claim),
	// then removals: reopen removes resolution.md after its history line.
	want := "d/context.md,d/history.md,d/baselines/b.md,d/ready.md,d/claim.md,d/resolution.md"
	if got := strings.Join(touched, ","); got != want {
		t.Fatalf("order = %s, want %s", got, want)
	}
}
