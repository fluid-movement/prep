package domain

import (
	"testing"
	"time"
)

func TestExplicitPolicyGatesOnlyNewReadiness(t *testing.T) {
	i := &Issue{ID: "01M4AAAAAAAAAAAAAAAAAAAAAA", Title: "Work", Kind: KindManual, Body: "Work.", Prose: "Work.", Criteria: []Criterion{{Text: "Verified", Checked: true}}, Baselines: []Baseline{{Name: "20260101-000000", Kind: KindManual, Requirement: "Work."}}}
	tr := NewTree(Project{DoD: []string{"legacy tests pass"}}, []*Issue{i}, nil, nil)
	unmet, _ := tr.Gates(i.ID, OpReady, nil)
	if len(unmet) != 1 || unmet[0].Code != GatePolicy {
		t.Fatalf("legacy DoD must not imply policy: %+v", unmet)
	}
	c, err := tr.PlanPolicy(CompletionPolicy{Verification: " Manually inspect behavior. ", Documentation: "Review impact.", By: "test/1", At: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	next := tr.Apply(c)
	if tr.Project.Policy != nil {
		t.Fatal("mutated original tree")
	}
	if got, _ := next.Gates(i.ID, OpReady, nil); len(got) != 0 {
		t.Fatalf("configured policy: %+v", got)
	}
	i.Ready = &Ready{Baseline: i.Baselines[0].Name}
	i.Claim = &Claim{By: "agent"}
	// Existing completions are not retroactively given evidence or a policy.
	if got, _ := tr.Gates(i.ID, OpComplete, &Input{Actor: "agent", NoImpact: "No documentation impact."}); len(got) != 0 {
		t.Fatalf("legacy completion changed: %+v", got)
	}
	if _, err := tr.PlanPolicy(CompletionPolicy{Verification: "Only one method"}); err == nil {
		t.Fatal("accepted incomplete policy")
	}
}
