package domain

import (
	"strings"
	"testing"
	"time"
)

func TestResolutionReviewRequiresHumanReasonAndScope(t *testing.T) {
	issue := &Issue{ID: idA, Title: "Done", Kind: KindManual, Resolution: &Resolution{Outcome: OutcomeDone, Generation: 1}}
	tr := NewTree(Project{}, []*Issue{issue}, nil, nil)
	base := Input{Actor: "human:reviewer", Now: time.Now(), Reason: "Reviewed", Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	for _, bad := range []Input{
		{Actor: "agent/1", Now: base.Now, Reason: base.Reason, Scopes: base.Scopes, Snapshot: base.Snapshot},
		{Actor: base.Actor, Now: base.Now, Reason: " ", Scopes: base.Scopes, Snapshot: base.Snapshot},
		{Actor: base.Actor, Now: base.Now, Reason: base.Reason, Snapshot: base.Snapshot},
	} {
		if _, err := tr.Plan(issue.ID, OpResolutionReview, bad); err == nil {
			t.Fatalf("accepted invalid review input: %+v", bad)
		}
	}
	change, err := tr.Plan(issue.ID, OpResolutionReview, base)
	if err != nil {
		t.Fatal(err)
	}
	got := (*change.ResolutionReviews)[0]
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestResolutionReviewStatusRequiresValidCurrentReview(t *testing.T) {
	res := &Resolution{Outcome: OutcomeDone, Generation: 2}
	i := &Issue{ID: idA, Title: "Done", Kind: KindManual, Resolution: res}
	tr := NewTree(Project{}, []*Issue{i}, nil, nil)
	if got := tr.ResolutionReviewStatus(idA); got != ReviewPending {
		t.Fatalf("status = %s", got)
	}
	r := ResolutionReview{ResolutionID: res.ContentID(), Generation: 2, Reason: "Reviewed", By: "human:reviewer", At: time.Now(), Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	r.ID = r.ContentID()
	i.ResolutionReviews = []ResolutionReview{r}
	i.ResolutionReviewCurrent = map[string]bool{r.ID: true}
	if got := tr.ResolutionReviewStatus(idA); got != ReviewApproved {
		t.Fatalf("status = %s, want approved", got)
	}
	res.Generation++
	if got := tr.ResolutionReviewStatus(idA); got != ReviewStale {
		t.Fatalf("new generation status = %s, want stale", got)
	}
	res.Generation--
	r.Reason = "tampered"
	i.ResolutionReviews = []ResolutionReview{r}
	i.ResolutionReviewCurrent[r.ID] = true
	if got := tr.ResolutionReviewStatus(idA); got != ReviewStale {
		t.Fatalf("invalid latest review status = %s, want stale", got)
	}
}

func TestReopenArchivesResolutionAndUsesNewGeneration(t *testing.T) {
	res := &Resolution{Outcome: OutcomeDone, Generation: 3, Note: "previous result"}
	i := &Issue{ID: idA, Title: "Done", Kind: KindManual, Resolution: res, Claim: &Claim{By: "agent/1"}}
	tr := NewTree(Project{}, []*Issue{i}, nil, nil)
	c, err := tr.Plan(idA, OpReopen, Input{Actor: "agent/1", Now: time.Now(), Reason: "new evidence"})
	if err != nil {
		t.Fatal(err)
	}
	reopened := tr.Apply(c)
	if reopened.Issues[idA].Resolution != nil || reopened.Issues[idA].Claim != nil || len(reopened.Issues[idA].ResolutionHistory) != 1 {
		t.Fatalf("reopen did not archive and clear active resolution/claim: %+v", reopened.Issues[idA])
	}
	if got := nextResolutionGeneration(reopened.Issues[idA]); got != 4 {
		t.Fatalf("next generation = %d, want 4", got)
	}
}

func TestReviewScopesIncludeExceptionApprovals(t *testing.T) {
	r := &Resolution{Verification: []VerificationEvidence{{Scopes: []string{"verified"}, ConfigScopes: []string{"checks"}}}, Approvals: []CompletionApproval{{Kind: ApprovalException, Scopes: []string{"waived"}, ConfigScopes: []string{"waiver-check"}}}}
	if got := strings.Join(ReviewWorkScopes(r), ","); got != "verified,waived" {
		t.Fatalf("work scopes = %s", got)
	}
	if got := strings.Join(ReviewConfigScopes(r), ","); got != "checks,waiver-check" {
		t.Fatalf("config scopes = %s", got)
	}
}
