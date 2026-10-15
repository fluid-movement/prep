package mdstore

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

// This covers the two parallel slices together: the override's content chain
// and its audit history must survive the generic transaction boundary.
func TestDependencyOverrideRecoveryAndConcurrentHistory(t *testing.T) {
	s := Open(t.TempDir())
	id := transactionIssue(t, s)
	load := func(store *Store) *domain.Tree {
		t.Helper()
		p, issues, _, err := store.Load()
		if err != nil {
			t.Fatal(err)
		}
		return domain.NewTree(p, issues, nil, nil)
	}
	load(s)
	links := []domain.ProjectLink{{Name: "api", ID: "01M4BBBBBBBBBBBBBBBBBBBBBB"}}
	if _, err := s.Apply(&domain.Change{Links: &links}); err != nil {
		t.Fatal(err)
	}
	tr := load(s)
	ref := "api:01M4CCCCCCCCCCCCCCCCCCCCCC"
	deps := []string{ref}
	edit, err := tr.PlanEdit(id, domain.EditInput{Actor: "test", Now: time.Now(), DependsOn: &deps})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(edit); err != nil {
		t.Fatal(err)
	}
	tr = load(s)
	change, err := tr.PlanDependencyOverride(id, domain.DependencyOverrideInput{Ref: ref, Actor: "human:reviewer", Reason: "Accept this observed uncertainty", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	interrupted := errors.New("interrupted override batch")
	s.txHook = func(stage string, index int) error {
		if stage == "after-install" && index == 0 {
			return interrupted
		}
		return nil
	}
	if _, err := s.Apply(change); !errors.Is(err, interrupted) {
		t.Fatalf("interruption = %v", err)
	}
	restarted := Open(s.Root)
	tr = load(restarted)
	i := tr.Issues[id]
	if len(i.DependencyOverrides) != 1 || !strings.Contains(i.History, "Accept this observed uncertainty") {
		t.Fatal("recovery did not restore override and audit history together")
	}
	if err := domain.ValidateDependencyOverrides(i.DependencyOverrides); err != nil {
		t.Fatal(err)
	}
	if tr.DependencyOverrideStatus(id, ref) != "current" {
		t.Fatal("recovered override is not applicable to its original observation")
	}
	// A requirement edit stales the first override; both writers now plan a
	// replacement against the same existing chain and history.
	body := "A revised requirement."
	edit, err = tr.PlanEdit(id, domain.EditInput{Actor: "test", Now: time.Now(), Body: &body})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.Apply(edit); err != nil {
		t.Fatal(err)
	}
	a, b := Open(s.Root), Open(s.Root)
	first, err := load(a).PlanDependencyOverride(id, domain.DependencyOverrideInput{Ref: ref, Actor: "human:first", Reason: "First decision", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	second, err := load(b).PlanDependencyOverride(id, domain.DependencyOverrideInput{Ref: ref, Actor: "human:second", Reason: "Second decision", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Apply(first); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Apply(second); err == nil {
		t.Fatal("stale override writer replaced the chain")
	} else {
		var conflict *domain.Error
		if !errors.As(err, &conflict) || conflict.Code != domain.ErrConflict {
			t.Fatalf("conflict = %v", err)
		}
	}
	i = load(Open(s.Root)).Issues[id]
	if len(i.DependencyOverrides) != 2 || i.DependencyOverrides[1].Previous != i.DependencyOverrides[0].ID {
		t.Fatal("override content chain lost integrity")
	}
	if err := domain.ValidateDependencyOverrides(i.DependencyOverrides); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(i.History, "First decision") || strings.Contains(i.History, "Second decision") {
		t.Fatal("stale writer altered audit history")
	}
}
