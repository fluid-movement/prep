package domain

import (
	"testing"
	"time"
)

func TestContractTracksEffectivePolicyNotBookkeeping(t *testing.T) {
	p := &Issue{ID: "01M4AAAAAAAAAAAAAAAAAAAAAA", Title: "Parent", Kind: KindManual, DoDAdd: []string{"parent check"}}
	i := &Issue{ID: "01M4BBBBBBBBBBBBBBBBBBBBBB", Title: "Child", Kind: KindManual, Parent: p.ID, Body: "Work", Prose: "Work", Criteria: []Criterion{{Text: "Works"}}, Baselines: []Baseline{{Name: "20260101-000000", Kind: KindManual, Requirement: "Work"}}}
	policy := &CompletionPolicy{Verification: "Test behavior", Documentation: "Review knowledge", By: "test", At: time.Now()}
	tr := NewTree(Project{Policy: policy, DoD: []string{"project check"}}, []*Issue{p, i}, nil, nil)
	c, err := tr.Plan(i.ID, OpReady, Input{Actor: "test", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	tr = tr.Apply(c)
	initial := tr.Contract(i.ID)
	if tr.PolicyChanged(i.ID) {
		t.Fatal("fresh signoff already stale")
	}
	tr.Project.Policy.By = "another actor"
	tr.Project.Policy.At = time.Now().Add(time.Hour)
	tr.Project.Config.Theme = "nord"
	tr.Issues[i.ID].Criteria[0].Checked = true
	if tr.Contract(i.ID).ID != initial.ID {
		t.Fatal("bookkeeping changed contract")
	}
	tr.Issues[p.ID].DoDAdd = append(tr.Issues[p.ID].DoDAdd, "new ancestor check")
	if !tr.PolicyChanged(i.ID) || tr.Actionable(i.ID) {
		t.Fatal("ancestor policy change unnoticed")
	}
	if _, err := tr.Plan(i.ID, OpClaim, Input{Actor: "test", Now: time.Now()}); err == nil {
		t.Fatal("claimed against changed policy")
	}
	adoption, err := tr.PlanAdoptPolicy(i.ID, Input{Actor: "test", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	next := tr.Apply(adoption)
	if next.PolicyChanged(i.ID) || !tr.PolicyChanged(i.ID) {
		t.Fatal("adoption mutated old tree or failed")
	}
	next.Issues[i.ID].Parent = ""
	if !next.PolicyChanged(i.ID) {
		t.Fatal("reparenting unnoticed")
	}
	// Stable requirement identities survive unrelated policy changes.
	old := map[string]string{}
	for _, r := range initial.Requirements {
		old[r.Kind+":"+r.Text] = r.ID
	}
	for _, r := range next.Contract(i.ID).Requirements {
		if want, ok := old[r.Kind+":"+r.Text]; ok && r.ID != want {
			t.Fatal("unstable requirement ID")
		}
	}
}
