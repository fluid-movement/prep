package domain

import (
	"strings"
	"testing"
	"time"
)

func TestEvidenceIsBoundToRequirementsAndLatestResult(t *testing.T) {
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	i := &Issue{ID: "01M4AAAAAAAAAAAAAAAAAAAAAA", Title: "Work", Kind: KindManual, Body: "Work", Prose: "Work",
		Criteria: []Criterion{{Text: "Observed", Checked: true}}, Baselines: []Baseline{{Name: "20261014-100000", Kind: KindManual, Requirement: "Work"}}}
	tr := NewTree(Project{Policy: &CompletionPolicy{Verification: "Inspect", Documentation: "Review impact", By: "test", At: now}}, []*Issue{i}, nil, nil)
	ready, err := tr.Plan(i.ID, OpReady, Input{Actor: "test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	tr = tr.Apply(ready)
	e := VerificationEvidence{Requirement: tr.Issues[i.ID].Ready.Contract.Requirements[0].ID, Result: "pass", Method: "manual",
		Observation: "Observed", By: "test", At: now, Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	pass, err := tr.PlanEvidence(i.ID, e)
	if err != nil {
		t.Fatal(err)
	}
	withPass := tr.Apply(pass)
	if len(tr.Issues[i.ID].Verification) != 0 {
		t.Fatal("recording mutated previous tree")
	}
	if got := LatestEvidence(withPass.Issues[i.ID], e.Requirement); got == nil || got.Result != "pass" || got.Provenance != "reported" {
		t.Fatal("missing attributed report")
	}
	e.Result = "fail"
	fail, err := withPass.PlanEvidence(i.ID, e)
	if err != nil {
		t.Fatal(err)
	}
	withFailure := withPass.Apply(fail)
	if got := LatestEvidence(withFailure.Issues[i.ID], e.Requirement); got == nil || got.Result != "fail" || got.Previous != (*pass.Verification)[0].ID {
		t.Fatal("old pass hid the failed report")
	}
	withFailure.Project.Policy.Verification = "New check configuration"
	if _, err := withFailure.PlanEvidence(i.ID, e); err == nil {
		t.Fatal("accepted report against changed policy")
	}
	adopt, err := withFailure.PlanAdoptPolicy(i.ID, Input{Actor: "test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	adopted := withFailure.Apply(adopt)
	if got := LatestEvidence(adopted.Issues[i.ID], e.Requirement); got != nil {
		t.Fatal("reused old contract evidence after adoption")
	}
	e.Provenance = "verified"
	if _, err := tr.PlanEvidence(i.ID, e); err == nil {
		t.Fatal("agent report claimed independently verified provenance")
	}
}
