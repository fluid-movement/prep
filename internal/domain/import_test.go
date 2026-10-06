package domain

import (
	"errors"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func TestPlanImport(t *testing.T) {
	tree := NewTree(Project{}, nil, nil, nil)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	cs, err := tree.PlanImport("test/1", now, rand.NewChaCha8([32]byte{}))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cs {
		if err := tree.CheckWrite(c); err != nil {
			t.Fatalf("%s: %v", c.Op, err)
		}
		tree = tree.Apply(c)
	}
	id := cs[0].IssueID
	i := tree.Issues[id]
	if i == nil || i.Title != "Import existing work" || i.Kind != KindResearch || !contains(i.Tags, ImportTag) || tree.State(id) != StateOpen {
		t.Fatalf("import issue = %+v", i)
	}
	if tree.ImportOpen() != id {
		t.Fatalf("ImportOpen = %q, want %q", tree.ImportOpen(), id)
	}
	ctx := i.Context
	for _, want := range []string{"prep findings " + id, `prep list --text "Source: <source>"`, "prep new --title", "Source: <source>", "prep complete " + id} {
		if !strings.Contains(ctx, want) {
			t.Errorf("context lacks %q", want)
		}
	}
	if len(i.Criteria) != 3 {
		t.Fatalf("criteria = %+v", i.Criteria)
	}

	// A second import waits until the first is resolved.
	_, err = tree.PlanImport("test/1", now, nil)
	var e *Error
	if !errors.As(err, &e) || e.Code != ErrInvalid || !strings.Contains(e.Message, "prep guide "+id) {
		t.Fatalf("second import: %v", err)
	}
	c, err := tree.Plan(id, OpDrop, Input{Actor: "test/1", Now: now, Reason: "not now"})
	if err != nil {
		t.Fatal(err)
	}
	tree = tree.Apply(c)
	if tree.ImportOpen() != "" {
		t.Fatal("a dropped import is still open")
	}
	if _, err := tree.PlanImport("test/1", now, nil); err != nil {
		t.Fatalf("import after the first was dropped: %v", err)
	}
}
