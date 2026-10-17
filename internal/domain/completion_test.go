package domain

import (
	"slices"
	"strings"
	"testing"
)

func TestCompleteSourceStates(t *testing.T) {
	for _, kind := range Kinds {
		for _, parent := range []bool{false, true} {
			for _, state := range []State{StateOpen, StateDefined, StateReady, StateInProgress, StateDone, StateDropped} {
				t.Run(string(kind)+"/"+string(state)+map[bool]string{true: "/parent", false: "/leaf"}[parent], func(t *testing.T) {
					i := &Issue{ID: idA, Kind: kind, Body: "Do it.", Prose: "Do it."}
					if state != StateOpen {
						i.Baselines = []Baseline{{Name: "20261017-120000", Kind: kind, Requirement: i.Body}}
					}
					if state == StateReady || state == StateInProgress {
						i.Ready = &Ready{Baseline: i.Baselines[0].Name}
					}
					if state == StateInProgress {
						i.Claim = &Claim{By: "agent"}
					}
					if state == StateDone {
						i.Resolution = &Resolution{Outcome: OutcomeDone}
					}
					if state == StateDropped {
						i.Resolution = &Resolution{Outcome: OutcomeDropped}
					}
					issues := []*Issue{i}
					if parent {
						issues = append(issues, &Issue{ID: idB, Parent: idA, Resolution: &Resolution{Outcome: OutcomeDone}})
					}
					tree := NewTree(Project{}, issues, nil, nil)
					// Parents complete from ready (they are never claimed); decision and
					// research leaves also complete from defined.
					want := state == StateInProgress || (state == StateReady && parent) || (state == StateDefined && !parent && (kind == KindDecision || kind == KindResearch))
					if got := tree.Applicable(idA, OpComplete); got != want {
						t.Fatalf("applicable = %v, want %v", got, want)
					}
					unmet, _ := tree.Gates(idA, OpComplete, &Input{Actor: "human:andre"})
					if got := slices.Contains(codes(unmet), GateState); got == want {
						t.Fatalf("state gates: %+v", unmet)
					}
				})
			}
		}
	}
}

func TestDefinedResultCompletionGatesAndGuide(t *testing.T) {
	for _, kind := range []Kind{KindDecision, KindResearch} {
		t.Run(string(kind), func(t *testing.T) {
			i := &Issue{ID: idA, Kind: kind, Body: "Answer it.", Prose: "Answer it.", Baselines: []Baseline{{Name: "20261017-120000", Kind: kind, Requirement: "Answer it."}}}
			tree := NewTree(Project{}, []*Issue{i}, nil, nil)
			resultGate, command := GateOutcome, "prep decide "
			if kind == KindResearch {
				resultGate, command = GateFindings, "prep finding "
			}
			in := Input{Actor: "human:andre"}
			if got := codes(must(tree.Gates(idA, OpComplete, &in))); !slices.Equal(got, []string{resultGate}) {
				t.Fatalf("missing result: %v", got)
			}
			guide := tree.BuildGuide(idA, func(id string) string { return id }, nil)
			text := strings.Join(guide.Instructions, "\n")
			if guide.Step != "complete" || !strings.Contains(text, command+idA) || !strings.Contains(text, "prep complete "+idA) || len(guide.Outputs) < 2 {
				t.Fatalf("guide: %+v", guide)
			}
			typ := "decision"
			if kind == KindResearch {
				typ = "finding"
			}
			tree.Knowledge["/"+typ+"s/"+idB+".md"] = &Entry{ID: idB, Path: "/" + typ + "s/" + idB + ".md", Type: typ, Issue: idA, Outcome: true}
			if unmet, _ := tree.Gates(idA, OpComplete, nil); len(unmet) != 0 {
				t.Fatalf("guide completion blocked: %+v", unmet)
			}
			guide = tree.BuildGuide(idA, func(id string) string { return id }, nil)
			found := false
			for _, transition := range guide.Transitions {
				if transition.Op == OpComplete {
					found = true
					if !transition.Allowed {
						t.Fatalf("guide completion blocked: %+v", transition)
					}
				}
			}
			if !found {
				t.Fatal("guide omits completion transition")
			}
			c, err := tree.Plan(idA, OpComplete, in)
			if err != nil {
				t.Fatal(err)
			}
			if c.RemoveClaim || c.Claim != nil || c.Resolution == nil || tree.Apply(c).State(idA) != StateDone {
				t.Fatalf("completion without claim: %+v", c)
			}
			i.Criteria = []Criterion{{Text: "Answered"}}
			if got := codes(must(tree.Gates(idA, OpComplete, &in))); !slices.Equal(got, []string{GateUnchecked}) {
				t.Fatalf("unchecked: %v", got)
			}
			i.Criteria[0].Checked = true
			if got := codes(must(tree.Gates(idA, OpComplete, &Input{}))); len(got) != 0 {
				t.Fatalf("completion: %v", got)
			}
			i.Body = "Changed question."
			if got := codes(must(tree.Gates(idA, OpComplete, &in))); !slices.Equal(got, []string{GateStale}) {
				t.Fatalf("stale: %v", got)
			}
		})
	}
}
