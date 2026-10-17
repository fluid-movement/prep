package mdstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func transactionIssue(t *testing.T, s *Store) string {
	t.Helper()
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Write transaction", Kind: domain.KindCode, Body: "Exercise recovery."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	return c.IssueID
}

func TestLoadOutsideProjectDoesNotCreatePrepOrLock(t *testing.T) {
	root := t.TempDir()
	if _, _, _, err := Open(root).Load(); err == nil {
		t.Fatal("Load without .prep unexpectedly succeeded")
	}
	if _, err := os.Stat(filepath.Join(root, Dir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read created .prep: %v", err)
	}
}

func TestApplyAllowsBaselinePath(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	project, issues, diagnostics, err := Open(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	tree := domain.NewTree(project, issues, nil, diagnostics)
	c, err := tree.Plan(id, domain.OpDefine, domain.Input{Actor: "human:test", Now: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root).Apply(c); err != nil {
		t.Fatalf("baseline transaction: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, IssueDir(id), "baselines", c.Baseline.Name+".md")); err != nil {
		t.Fatalf("baseline was not written: %v", err)
	}
}
