package domain

import (
	"strings"
	"testing"
	"time"
)

const dependencyIssueID = "01M4AAAAAAAAAAAAAAAAAAAAAA"
const dependencyTargetID = "01M4BBBBBBBBBBBBBBBBBBBBBB"

func TestUnknownDependencyOverrideIsHumanScopedAndDoesNotWaiveApproval(t *testing.T) {
	ref := "api:" + dependencyTargetID
	i := &Issue{ID: dependencyIssueID, Title: "Dependent", Kind: KindManual, DependsOn: []string{ref}, DependsOnApproval: []string{ref}}
	tree := NewTree(Project{Links: []ProjectLink{{Name: "api", ID: "01M4CCCCCCCCCCCCCCCCCCCCCC"}}}, []*Issue{i}, nil, nil)
	tree.External = map[string]XIssue{ref: {Ref: ref, Link: "api", ID: dependencyTargetID, Status: XUnknown, Outcome: DependencyUnknown, Uncertainty: "source checkout unavailable"}}
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	if _, err := tree.PlanDependencyOverride(i.ID, DependencyOverrideInput{Ref: ref, Actor: "codex", Reason: "Proceed", Now: now}); err == nil || !strings.Contains(err.Error(), "human:") {
		t.Fatalf("agent attribution override error = %v", err)
	}
	change, err := tree.PlanDependencyOverride(i.ID, DependencyOverrideInput{Ref: ref, Actor: "human:alex", Reason: "Proceed while source unavailable", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	withOverride := tree.Apply(change)
	if withOverride.DependencyOverrideStatus(i.ID, ref) != "current" {
		t.Fatal("override should apply to current unknown observation")
	}
	if !withOverride.Blocked(i.ID) {
		t.Fatal("unknown dependency override must not waive required resolution approval")
	}
	if _, err := withOverride.PlanDependencyOverride(i.ID, DependencyOverrideInput{Ref: ref, Actor: "human:alex", Reason: "again", Now: now.Add(time.Minute)}); err == nil {
		t.Fatal("duplicate current override should be rejected")
	}
}

func TestDependencyOverrideStalesWhenLocalConfigurationChanges(t *testing.T) {
	ref := "api:" + dependencyTargetID
	i := &Issue{ID: dependencyIssueID, Title: "Dependent", Kind: KindManual, Body: "first", DependsOn: []string{ref}}
	tree := NewTree(Project{}, []*Issue{i}, nil, nil)
	tree.External = map[string]XIssue{ref: {Ref: ref, Link: "api", ID: dependencyTargetID, Status: XUnknown, Outcome: DependencyUnknown, Uncertainty: "not checked out"}}
	change, err := tree.PlanDependencyOverride(i.ID, DependencyOverrideInput{Ref: ref, Actor: "human:alex", Reason: "Proceed", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	withOverride := tree.Apply(change)
	if withOverride.DependencyOverrideStatus(i.ID, ref) != "current" {
		t.Fatal("override should initially be current")
	}
	changed := *withOverride.Issues[i.ID]
	changed.DependsOnApproval = []string{ref}
	changedTree := NewTree(Project{}, []*Issue{&changed}, nil, nil)
	changedTree.External = withOverride.External
	if changedTree.DependencyOverrideStatus(i.ID, ref) != "stale" {
		t.Fatal("approval prerequisite change must stale override")
	}
	if !changedTree.Blocked(i.ID) {
		t.Fatal("stale override must block")
	}
}

func TestDependencyOverrideStalesWhenSourceObservationChanges(t *testing.T) {
	ref := "api:" + dependencyTargetID
	i := &Issue{ID: dependencyIssueID, Title: "Dependent", Kind: KindManual, DependsOn: []string{ref}}
	tree := NewTree(Project{}, []*Issue{i}, nil, nil)
	tree.External = map[string]XIssue{ref: {Ref: ref, Link: "api", ID: dependencyTargetID, Status: XUnknown, Outcome: DependencyUnknown, Uncertainty: "checkout unavailable"}}
	change, err := tree.PlanDependencyOverride(i.ID, DependencyOverrideInput{Ref: ref, Actor: "human:alex", Reason: "Proceed", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	withOverride := tree.Apply(change)
	withOverride.External[ref] = XIssue{Ref: ref, Link: "api", ID: dependencyTargetID, Status: XUnknown, Outcome: DependencyUnknown, Uncertainty: "registry entry missing"}
	if withOverride.DependencyOverrideStatus(i.ID, ref) != "stale" {
		t.Fatal("changed source observation must stale override")
	}
	if !withOverride.Blocked(i.ID) {
		t.Fatal("stale source override must block")
	}
}

func TestExternalDependencyOutcomeIncludesRequiredReview(t *testing.T) {
	ref := "api:" + dependencyTargetID
	i := &Issue{ID: dependencyIssueID, Title: "Dependent", Kind: KindManual, DependsOn: []string{ref}, DependsOnApproval: []string{ref}}
	tree := NewTree(Project{}, []*Issue{i}, nil, nil)
	tree.External = map[string]XIssue{ref: {Ref: ref, Link: "api", ID: dependencyTargetID, Status: XFound, State: StateDone, Outcome: DependencySatisfied, ReviewStatus: ReviewPending}}
	if got := tree.ExternalDependency(i.ID, ref); got.Outcome != DependencyUnsatisfied || !got.RequiresApproval {
		t.Fatalf("pending required review = %#v", got)
	}
	tree.External[ref] = XIssue{Ref: ref, Link: "api", ID: dependencyTargetID, Status: XFound, State: StateDone, Outcome: DependencySatisfied, ReviewStatus: ReviewApproved}
	if got := tree.ExternalDependency(i.ID, ref); got.Outcome != DependencySatisfied {
		t.Fatalf("approved required review = %#v", got)
	}
	i.DependsOnApproval = nil
	if got := tree.ExternalDependency(i.ID, ref); got.Outcome != DependencySatisfied || got.RequiresApproval {
		t.Fatalf("default done dependency = %#v", got)
	}
}
