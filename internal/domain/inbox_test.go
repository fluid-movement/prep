package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	inboxID1  = "01M4AAAAAAAAAAAAAAAAAAAAAA"
	inboxID2  = "01M4BBBBBBBBBBBBBBBBBBBBBB"
	inboxID3  = "01M4CCCCCCCCCCCCCCCCCCCCCC"
	inboxID4  = "01M4DDDDDDDDDDDDDDDDDDDDDD"
	inboxID5  = "01M4EEEEEEEEEEEEEEEEEEEEEE"
	inboxID6  = "01M4GGGGGGGGGGGGGGGGGGGGGG"
	inboxBase = "20260101-000000"
)

func inboxPolicy() *CompletionPolicy {
	return &CompletionPolicy{Verification: "run tests", Documentation: "review knowledge", By: "test", At: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func inboxDefined(id string, kind Kind) *Issue {
	return &Issue{ID: id, Title: "Issue " + id, Kind: kind, Body: "Requirement", Prose: "Requirement",
		Criteria: []Criterion{{Text: "Works"}}, Baselines: []Baseline{{Name: inboxBase, Kind: kind, Requirement: "Requirement"}}}
}

func inboxResolved(id string, at time.Time, priority Priority) *Issue {
	return &Issue{ID: id, Title: "Done " + id, Kind: KindManual, Priority: priority,
		Resolution: &Resolution{Outcome: OutcomeDone, By: "agent", At: at}}
}

func TestInboxCompletionReviewStatusesAndOrdering(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	pendingNew := inboxResolved(inboxID1, now, PriorityMedium)
	pendingOld := inboxResolved(inboxID2, now.Add(-time.Hour), PriorityHigh)
	stale := inboxResolved(inboxID3, now.Add(-2*time.Hour), PriorityCritical)
	stale.ResolutionReviews = []ResolutionReview{{ResolutionID: "old"}}
	approved := inboxResolved(inboxID4, now.Add(time.Hour), PriorityCritical)
	r := ResolutionReview{ResolutionID: approved.Resolution.ContentID(), Generation: 1, Reason: "Reviewed", By: "human:test",
		At: now, Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	r.ID = r.ContentID()
	approved.ResolutionReviews = []ResolutionReview{r}
	approved.ResolutionReviewCurrent = map[string]bool{r.ID: true}
	dropped := &Issue{ID: inboxID5, Title: "Dropped", Kind: KindManual, Resolution: &Resolution{Outcome: OutcomeDropped, At: now}}
	ordinary := &Issue{ID: "01M4FFFFFFFFFFFFFFFFFFFFFF", Title: "Backlog", Kind: KindManual}

	tree := NewTree(Project{}, []*Issue{pendingOld, approved, dropped, pendingNew, ordinary, stale}, nil, nil)
	got := tree.Inbox()
	want := []InboxItem{
		{IssueID: inboxID3, Category: InboxCompletionReview, Reason: "completed work needs review", ReviewStatus: ReviewStale},
		{IssueID: inboxID2, Category: InboxCompletionReview, Reason: "completed work needs review", ReviewStatus: ReviewPending},
		{IssueID: inboxID1, Category: InboxCompletionReview, Reason: "completed work needs review", ReviewStatus: ReviewPending},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Inbox() = %#v, want %#v", got, want)
	}
}

func TestInboxReadyQuestionsStaleAndPolicyPrecedence(t *testing.T) {
	policy := inboxPolicy()
	ready := inboxDefined(inboxID1, KindManual)

	questions := &Issue{ID: inboxID2, Title: "Questions", Kind: KindManual, OpenQuestions: "Which behavior should we choose?"}
	staleAndQuestions := inboxDefined(inboxID3, KindManual)
	staleAndQuestions.Body = "Changed requirement"
	staleAndQuestions.OpenQuestions = "Which behavior should we choose?"
	unknown := &Issue{ID: inboxID4, Title: "Unknown dependency", Kind: KindManual, DependsOn: []string{"api:01M4FFFFFFFFFFFFFFFFFFFFFF"}}

	changedPolicyIssue := inboxDefined(inboxID5, KindManual)
	oldTree := NewTree(Project{Policy: policy}, []*Issue{changedPolicyIssue}, nil, nil)
	changedPolicyIssue.Ready = &Ready{Baseline: inboxBase, Contract: oldTree.Contract(inboxID5)}
	changedPolicy := *policy
	changedPolicy.Verification = "run the full test suite"
	stalePolicyIssue := inboxDefined(inboxID6, KindManual)
	oldPolicyTree := NewTree(Project{Policy: policy}, []*Issue{stalePolicyIssue}, nil, nil)
	stalePolicyIssue.Ready = &Ready{Baseline: inboxBase, Contract: oldPolicyTree.Contract(inboxID6)}
	stalePolicyIssue.Body = "Changed requirement"

	tree := NewTree(Project{Policy: &changedPolicy, Links: []ProjectLink{{Name: "api"}}},
		[]*Issue{ready, questions, staleAndQuestions, unknown, changedPolicyIssue, stalePolicyIssue}, nil, nil)
	got := tree.Inbox()
	want := []InboxItem{
		{IssueID: inboxID1, Category: InboxReadySignoff, Reason: "issue is ready for human sign-off"},
		{IssueID: inboxID2, Category: InboxOpenQuestions, Reason: "open questions need answers"},
		{IssueID: inboxID3, Category: InboxOpenQuestions, Reason: "open questions need answers"},
		{IssueID: inboxID6, Category: InboxStaleRequirement, Reason: "requirement or kind changed since the latest baseline"},
		{IssueID: inboxID5, Category: InboxStalePolicy, Reason: "bound completion policy changed and needs review"},
		{IssueID: inboxID4, Category: InboxUnknownDependency, Reason: "external dependency state is unavailable: api:01M4FFFFFFFFFFFFFFFFFFFFFF"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Inbox() = %#v, want %#v", got, want)
	}
}

func TestInboxReadyGatesAndStaleRequirement(t *testing.T) {
	blockedByContext := inboxDefined(inboxID1, KindCode)
	noPolicy := inboxDefined(inboxID2, KindManual)
	stale := inboxDefined(inboxID3, KindManual)
	stale.Body = "Changed requirement"

	tree := NewTree(Project{}, []*Issue{blockedByContext, noPolicy, stale}, nil, nil)
	got := tree.Inbox()
	want := []InboxItem{{IssueID: inboxID3, Category: InboxStaleRequirement,
		Reason: "requirement or kind changed since the latest baseline"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Inbox() = %#v, want %#v", got, want)
	}
	if unmet, _ := tree.Gates(inboxID1, OpReady, nil); len(unmet) == 0 {
		t.Fatal("fixture should fail the ready context gate")
	}
	if unmet, _ := tree.Gates(inboxID2, OpReady, nil); len(unmet) == 0 {
		t.Fatal("fixture should fail the ready policy gate")
	}
}

func TestInboxUnknownExternalDependencyBlocksClaim(t *testing.T) {
	i := &Issue{ID: inboxID1, Title: "External dependency", Kind: KindManual, DependsOn: []string{"api:01M4FFFFFFFFFFFFFFFFFFFFFF"}}
	tree := NewTree(Project{Links: []ProjectLink{{Name: "api"}}}, []*Issue{i}, nil, nil)
	if !tree.Blocked(i.ID) {
		t.Fatal("unknown external dependency must block claims until observed or overridden")
	}
	got := tree.Inbox()
	if len(got) != 1 || got[0].Category != InboxUnknownDependency || got[0].IssueID != i.ID {
		t.Fatalf("Inbox() = %#v", got)
	}
}

func TestInboxDoesNotMutateTree(t *testing.T) {
	i := inboxResolved(inboxID1, time.Now(), PriorityHigh)
	i.ResolutionReviews = []ResolutionReview{{ResolutionID: "old"}}
	tree := NewTree(Project{}, []*Issue{i, inboxDefined(inboxID2, KindManual)}, nil, nil)
	before, err := json.Marshal(struct {
		Project   Project
		Issues    map[string]*Issue
		Knowledge map[string]*Entry
		External  map[string]XIssue
	}{tree.Project, tree.Issues, tree.Knowledge, tree.External})
	if err != nil {
		t.Fatal(err)
	}
	_ = tree.Inbox()
	after, err := json.Marshal(struct {
		Project   Project
		Issues    map[string]*Issue
		Knowledge map[string]*Entry
		External  map[string]XIssue
	}{tree.Project, tree.Issues, tree.Knowledge, tree.External})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("Inbox mutated the tree")
	}
}
