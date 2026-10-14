package domain

import (
	"strings"
	"testing"
	"time"
)

func TestKnowledgeReviewsBindEntryWorkAndNoImpactToCompletion(t *testing.T) {
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	i := &Issue{ID: idA, Title: "Work", Kind: KindManual, Body: "Work", Prose: "Work", Criteria: []Criterion{{Text: "Observed", Checked: true}}, Baselines: []Baseline{{Name: "20261014-100000", Kind: KindManual, Requirement: "Work"}}}
	entry := &Entry{Path: "/overview.md", Type: "overview", Title: "Overview", Description: "Project overview", Body: "Relevant content"}
	policy := &CompletionPolicy{Verification: "Inspect", Documentation: "Review", By: "test", At: now}
	tr := NewTree(Project{Policy: policy}, []*Issue{i}, []*Entry{entry}, nil)
	i.Ready = &Ready{Baseline: "20261014-100000", Contract: tr.Contract(i.ID)}

	if _, problems := tr.KnowledgeReviewsForCompletion(i.ID, []string{entry.Path}, ""); len(problems) != 1 {
		t.Fatalf("bare docs reference passed: %v", problems)
	}
	r := KnowledgeReview{Kind: KnowledgeUpdated, Entry: entry.Path, Observation: "Compared the changed API and updated this entry", By: "agent/test", At: now, Scopes: []string{"internal/domain"}, Snapshot: strings.Repeat("a", 64)}
	change, err := tr.PlanKnowledgeReview(i.ID, r)
	if err != nil {
		t.Fatal(err)
	}
	withReview := tr.Apply(change)
	review := withReview.Issues[i.ID].KnowledgeReviews[0]
	withReview.Issues[i.ID].KnowledgeReviewCurrent = map[string]bool{review.ID: true}
	selected, problems := withReview.KnowledgeReviewsForCompletion(i.ID, []string{entry.Path}, "")
	if len(problems) != 0 || len(selected) != 1 || selected[0].ID != review.ID {
		t.Fatalf("current report not selected: %#v %v", selected, problems)
	}
	withReview.Knowledge[entry.Path].Body = "Later edited content"
	if _, problems := withReview.KnowledgeReviewsForCompletion(i.ID, []string{entry.Path}, ""); len(problems) != 1 {
		t.Fatalf("changed entry reused a stale report: %v", problems)
	}
	withReview.Knowledge[entry.Path].Body = "Relevant content"
	withReview.Issues[i.ID].KnowledgeReviewCurrent[review.ID] = false
	if _, problems := withReview.KnowledgeReviewsForCompletion(i.ID, []string{entry.Path}, ""); len(problems) != 1 {
		t.Fatalf("changed work reused a stale report: %v", problems)
	}

	noImpact := KnowledgeReview{Kind: KnowledgeNoImpact, Observation: "No knowledge change: internal implementation only", By: "agent/test", At: now.Add(time.Second), Scopes: []string{"internal/domain"}, Snapshot: strings.Repeat("b", 64)}
	change, err = tr.PlanKnowledgeReview(i.ID, noImpact)
	if err != nil {
		t.Fatal(err)
	}
	withNoImpact := tr.Apply(change)
	ni := withNoImpact.Issues[i.ID].KnowledgeReviews[0]
	completedIssue := withNoImpact.Issues[i.ID]
	completedIssue.KnowledgeReviewCurrent = map[string]bool{ni.ID: true}
	if _, problems = withNoImpact.KnowledgeReviewsForCompletion(i.ID, nil, ni.Observation); len(problems) != 0 {
		t.Fatalf("matching no-impact report rejected: %v", problems)
	}
	if _, problems = withNoImpact.KnowledgeReviewsForCompletion(i.ID, nil, "different reason"); len(problems) != 1 {
		t.Fatalf("misaligned no-impact report accepted: %v", problems)
	}
	completedIssue.Claim = &Claim{By: "agent/test", At: now}
	completedIssue.EvidenceCurrent = map[string]bool{}
	for _, requirement := range completedIssue.Ready.Contract.Requirements {
		e := VerificationEvidence{Contract: completedIssue.Ready.Contract.ID, Requirement: requirement.ID, Result: "pass", Method: "manual", Observation: "Checked", By: "agent/test", At: now, Provenance: "reported", Scopes: []string{"internal/domain"}, Snapshot: strings.Repeat("c", 64)}
		e.ID = e.ContentID()
		completedIssue.Verification = append(completedIssue.Verification, e)
		completedIssue.EvidenceCurrent[e.ID] = true
	}
	completion, err := withNoImpact.Plan(i.ID, OpComplete, Input{Actor: "agent/test", Now: now, NoImpact: ni.Observation})
	if err != nil {
		t.Fatal(err)
	}
	if len(completion.Resolution.KnowledgeReviews) != 1 || completion.Resolution.KnowledgeReviews[0].ID != ni.ID {
		t.Fatalf("resolution did not snapshot selected report: %#v", completion.Resolution.KnowledgeReviews)
	}
}
