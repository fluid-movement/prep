package mdstore

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestConcurrentApprovalDoesNotOverwriteHistoryOrRevisionState(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(domain.NewIssueInput{Title: "Work", Kind: domain.KindManual, Body: "Work."}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	id := c.IssueID
	load := func(s *Store) *domain.Tree {
		t.Helper()
		p, issues, diagnostics, err := s.Load()
		if err != nil {
			t.Fatal(err)
		}
		return domain.NewTree(p, issues, nil, diagnostics)
	}
	c, err = load(s).PlanRecord(id, domain.OpDoD, domain.RecordInput{Actor: "agent/test", Now: now, Acceptance: []domain.AcceptanceOp{{Op: "dod-opt-out", Text: "prep check passes", Reason: "Alternate verifier"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	w1, w2 := Open(root), Open(root)
	first, second := load(w1), load(w2)
	approve := func(tr *domain.Tree, actor string) *domain.Change {
		t.Helper()
		c, err := tr.PlanApproveOptOut(id, "prep check passes", domain.Input{Actor: actor, Now: now, Reason: "Reviewed"})
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	if _, err := w1.Apply(approve(first, "human:first")); err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join(root, IssueDir(id), "approvals.md")
	before, err := os.ReadFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	_, err = w2.Apply(approve(second, "human:second"))
	var conflict *domain.Error
	if !errors.As(err, &conflict) || conflict.Code != domain.ErrConflict {
		t.Fatalf("concurrent approval: %v", err)
	}
	after, err := os.ReadFile(rel)
	if err != nil || string(after) != string(before) {
		t.Fatal("concurrent writer changed approved history")
	}
	current := load(Open(root)).Issues[id]
	if len(current.Approvals) != 1 || current.DoDRevision == "" || current.OptOutRevisions["prep check passes"] == "" {
		t.Fatal("approval write lost request revision state")
	}
}

func approvalRecord(t *testing.T) domain.CompletionApproval {
	t.Helper()
	i := &domain.Issue{ID: "01M4AAAAAAAAAAAAAAAAAAAAAA", DoDOptOuts: []domain.OptOut{{Item: "Tests pass", Reason: "Alternate verifier"}}}
	tr := domain.NewTree(domain.Project{DoD: []string{"Tests pass"}}, []*domain.Issue{i}, nil, nil)
	c, err := tr.PlanApproveOptOut(i.ID, "Tests pass", domain.Input{Actor: "human:reviewer", Now: time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC), Reason: "Reviewed alternate verifier"})
	if err != nil {
		t.Fatal(err)
	}
	return (*c.Approvals)[0]
}

func TestApprovalHistoryRoundTripAndIntegrity(t *testing.T) {
	a := approvalRecord(t)
	next := a
	next.Previous, next.Reason = a.ID, "Reviewed again"
	next.ID = next.ContentID()
	raw := renderApprovals([]domain.CompletionApproval{a, next})
	loaded, err := parseApprovals(raw)
	if err != nil || len(loaded) != 2 || renderApprovals(loaded) != raw {
		t.Fatalf("approval roundtrip: %v", err)
	}
	if _, err := parseApprovals(strings.Replace(raw, "Reviewed alternate verifier", "Unrecorded change", 1)); err == nil {
		t.Fatal("accepted changed approval")
	}
	if _, err := parseApprovals(renderApprovals([]domain.CompletionApproval{next, a})); err == nil {
		t.Fatal("accepted reordered history")
	}
	for name, mutate := range map[string]func(*domain.CompletionApproval){
		"agent actor":    func(a *domain.CompletionApproval) { a.By = "agent/test" },
		"missing reason": func(a *domain.CompletionApproval) { a.Reason = "" },
		"changed item":   func(a *domain.CompletionApproval) { a.Item = "Different check" },
		"unknown kind":   func(a *domain.CompletionApproval) { a.Kind = "unknown" },
	} {
		t.Run(name, func(t *testing.T) {
			bad := a
			mutate(&bad)
			bad.ID = bad.ContentID()
			if _, err := parseApprovals(renderApprovals([]domain.CompletionApproval{bad})); err == nil {
				t.Fatal("accepted invalid approval with consistent content ID")
			}
		})
	}
}

func TestPolicyAndContractHistorySurviveCanonicalization(t *testing.T) {
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	old := domain.CompletionPolicy{Verification: "Old check", Documentation: "Review", By: "test", At: now}
	current := old
	current.Verification = "New check"
	p := domain.Project{Schema: 1, Policy: &current, PolicyHistory: []domain.CompletionPolicy{old}}
	raw := renderProject(p, "Project prose.")
	loaded, err := parseProject(raw)
	if err != nil || len(loaded.PolicyHistory) != 1 || loaded.PolicyHistory[0] != old || canonicalProject(raw) != raw {
		t.Fatalf("policy history: %v", err)
	}
	c := &domain.CompletionContract{Verification: "Old check", Documentation: "Review"}
	c.ID = c.ContentID()
	next := *c
	next.Verification = "New check"
	next.ID = next.ContentID()
	r := &domain.Ready{Contract: &next, PreviousContracts: []domain.CompletionContract{*c}, By: "test", At: now, Baseline: "20261014-100000"}
	raw = renderReady(r)
	ready, err := parseReady(raw)
	if err != nil || len(ready.PreviousContracts) != 1 || ready.PreviousContracts[0].ID != c.ID || renderReady(ready) != raw {
		t.Fatalf("ready history: %v", err)
	}
}
