package domain

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const artifactIssueID = "01M4AAAAAAAAAAAAAAAAAAAAAA"

func TestArtifactReferenceValidation(t *testing.T) {
	valid := ArtifactReference{Source: "tracker:team/project", Location: "https://tracker.example/item/42", Title: "Upstream item"}
	if err := ValidateArtifactReferences([]ArtifactReference{valid}); err != nil {
		t.Fatalf("valid reference rejected: %v", err)
	}

	for name, ref := range map[string]ArtifactReference{
		"empty source":     {Location: "https://tracker.example/item/42"},
		"empty location":   {Source: "tracker"},
		"source control":   {Source: "tracker\nteam", Location: "/item/42"},
		"location control": {Source: "tracker", Location: "/item/42\x00"},
		"title control":    {Source: "tracker", Location: "/item/42", Title: "line one\nline two"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateArtifactReferences([]ArtifactReference{ref}); err == nil {
				t.Fatal("invalid reference was accepted")
			}
		})
	}
	if err := ValidateArtifactReferences([]ArtifactReference{valid, valid}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate reference error = %v", err)
	}
}

func TestArtifactChangeDoesNotSatisfyCompletionGates(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	baseline := Baseline{Name: "20261015-120000", By: "human:test", At: now, Kind: KindManual, Requirement: "Inspect the behavior."}
	i := &Issue{ID: artifactIssueID, Title: "Inspect behavior", Kind: KindManual, Body: baseline.Requirement,
		Prose: baseline.Requirement, Criteria: []Criterion{{Text: "Behavior is understood", Checked: true}}, Baselines: []Baseline{baseline}}
	tree := NewTree(Project{}, []*Issue{i}, nil, nil)
	i.Ready = &Ready{By: "human:test", At: now, Baseline: baseline.Name}
	i.Claim = &Claim{By: "agent:test", At: now}
	i.Criteria[0].Checked = false
	beforeState, beforeStale := tree.State(i.ID), tree.Stale(i.ID)
	in := Input{Actor: "agent:test", Now: now, NoImpact: "Reference metadata only."}
	before, _ := tree.Gates(i.ID, OpComplete, &in)
	if !hasGate(before, GateUnchecked) {
		t.Fatalf("fixture should require checked criteria: %#v", before)
	}

	change, err := tree.PlanArtifact(i.ID, ArtifactInput{Reference: ArtifactReference{
		Source: "tracker:team/project", Location: "https://tracker.example/item/42", Title: "Upstream item",
	}, Actor: "agent:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if change.Resolution != nil {
		t.Fatalf("artifact change contains a resolution: %#v", change)
	}
	afterTree := tree.Apply(change)
	after, _ := afterTree.Gates(i.ID, OpComplete, &in)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("artifact changed completion gates:\nbefore %#v\nafter  %#v", before, after)
	}
	if afterTree.State(i.ID) != beforeState || afterTree.Stale(i.ID) != beforeStale {
		t.Fatalf("artifact changed lifecycle state: %s stale=%v; want %s stale=%v", afterTree.State(i.ID), afterTree.Stale(i.ID), beforeState, beforeStale)
	}
}

func hasGate(unmet []Unmet, code string) bool {
	for _, u := range unmet {
		if u.Code == code {
			return true
		}
	}
	return false
}
