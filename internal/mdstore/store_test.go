package mdstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestCompareAndSwapRejectsConcurrentEdit(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	tree := domain.NewTree(domain.Project{}, nil, nil, nil)
	c, err := tree.PlanNew(domain.NewIssueInput{Title: "A", Kind: domain.KindCode, Body: "A."}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}

	// Two writers load the same state.
	w1, w2 := Open(root), Open(root)
	for _, w := range []*Store{w1, w2} {
		if _, _, _, err := w.Load(); err != nil {
			t.Fatal(err)
		}
	}
	rel := IssueDir(c.IssueID) + "/history.md"
	if err := w1.write(rel, "first\n"); err != nil {
		t.Fatal(err)
	}
	err = w2.write(rel, "second\n")
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(root, rel))
	if string(b) != "first\n" {
		t.Fatalf("history.md = %q", b)
	}
	// No temp files are left behind.
	entries, _ := os.ReadDir(filepath.Join(root, IssueDir(c.IssueID)))
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".md" {
			t.Fatalf("leftover file %s", e.Name())
		}
	}
}

func TestRoundTripIsCanonical(t *testing.T) {
	raw := "---\ntitle: X\nkind: code\nparent: 20261005-120000\ndepends_on:\n  - 20261005-120001\n---\n\nProse line.\n\n## Open questions\n\n- why?\n"
	i := &domain.Issue{}
	if err := parseIssue(raw, i); err != nil {
		t.Fatal(err)
	}
	if got := renderIssue(i); got != raw {
		t.Fatalf("round trip differs:\n%q\n%q", got, raw)
	}
	if i.Prose != "Prose line." || i.OpenQuestions != "- why?" {
		t.Fatalf("prose %q questions %q", i.Prose, i.OpenQuestions)
	}
}
