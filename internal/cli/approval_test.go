package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func approvalFixture(t *testing.T) (*harness, string, *domain.CompletionContract) {
	t.Helper()
	h := newHarness(t)
	id := h.newIssue("--title", "Approve work", "--kind", "manual", "--body", "Inspect work.")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "Observed")
	h.ok("ready", id)
	h.ok("claim", id)
	h.ok("criterion", id, "--check", "1")
	for _, path := range []string{"work.txt", "checks.txt"} {
		if err := os.WriteFile(filepath.Join(h.dir, path), []byte("original"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var policy struct {
		Bound *domain.CompletionContract `json:"bound"`
	}
	h.jsonOf(&policy, "policy", "show", id)
	return h, id, policy.Bound
}

func TestHumanExceptionRequiresReasonAndCurrentWorkAndReport(t *testing.T) {
	h, id, contract := approvalFixture(t)
	h.verify(id)
	r := contract.Requirements[0]
	recordFailure := func() {
		h.ok("evidence", "record", id, "--requirement", r.ID, "--result", "fail", "--method", "manual", "--observation", "Observed failure.", "--scope", "work.txt", "--config-scope", "checks.txt")
	}
	approve := []string{"exception", "approve", id, "--requirement", r.ID, "--reason", "Accept the known limitation for this work", "--scope", "work.txt", "--config-scope", "checks.txt"}
	recordFailure()
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "none", "--json")
	h.fails("human", approve...)
	h.fails("reason", "exception", "approve", id, "--requirement", r.ID, "--scope", "work.txt", "--by", "human:reviewer")
	h.ok(append(approve, "--by", "human:reviewer")...)
	var report struct{ Records []struct{ Applicable bool } }
	h.jsonOf(&report, "exception", "show", id)
	if len(report.Records) != 1 || !report.Records[0].Applicable {
		t.Fatal("approval not applicable")
	}
	before := h.read(id, "approvals.md")
	h.ok("exception", "show", id)
	if before != h.read(id, "approvals.md") {
		t.Fatal("read rewrote approval history")
	}
	for _, path := range []string{"work.txt", "checks.txt"} {
		if err := os.WriteFile(filepath.Join(h.dir, path), []byte("changed"), 0o644); err != nil {
			t.Fatal(err)
		}
		h.fails("G_EVIDENCE", "complete", id, "--no-impact", "none", "--json")
		if err := os.WriteFile(filepath.Join(h.dir, path), []byte("original"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A later result is not the failure the human previously reviewed.
	recordFailure()
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "none", "--json")
	h.ok(append(approve, "--by", "human:reviewer")...)
	h.ok("complete", id, "--no-impact", "none")
	var closed struct {
		Issue struct {
			Resolution *domain.Resolution `json:"resolution"`
		} `json:"issue"`
	}
	h.jsonOf(&closed, "show", id)
	if closed.Issue.Resolution == nil || len(closed.Issue.Resolution.Approvals) != 1 || closed.Issue.Resolution.Approvals[0].Kind != domain.ApprovalException {
		t.Fatal("completion omitted the waiver")
	}
	if closed.Issue.Resolution.Evidence != "" {
		t.Fatal("completion required a commit")
	}
	h.ok("fmt", "--check")
}

func TestHumanExceptionCanCoverMissingReportAndExpiresOnAdoption(t *testing.T) {
	h, id, contract := approvalFixture(t)
	r := contract.Requirements[0]
	for _, requirement := range contract.Requirements[1:] {
		h.ok("evidence", "record", id, "--requirement", requirement.ID, "--result", "pass", "--method", "manual", "--observation", "Observed.", "--scope", "work.txt")
	}
	h.ok("exception", "approve", id, "--requirement", r.ID, "--reason", "Human explicitly waives this missing observation", "--scope", "work.txt", "--by", "human:reviewer")
	if got := h.ok("guide", id, "--json"); strings.Contains(got, `"code": "G_EVIDENCE"`) {
		t.Fatal("waiver did not cover missing report")
	}
	h.ok("policy", "set", "--verification", "Additional verification", "--documentation", "Update affected knowledge or explain no impact.")
	h.fails("G_POLICY", "complete", id, "--no-impact", "none", "--json")
	h.ok("policy", "adopt", id)
	h.fails("G_EVIDENCE", "complete", id, "--no-impact", "none", "--json")
	if got := h.ok("exception", "show", id, "--json"); !strings.Contains(got, `"applicable": false`) {
		t.Fatal("old-contract waiver survived adoption")
	}
	h.ok("fmt", "--check")
}

func TestWaivedAcceptanceRemainsVisiblyUnmet(t *testing.T) {
	h, id, contract := approvalFixture(t)
	h.verify(id)
	h.ok("criterion", id, "--uncheck", "1")
	r := contract.Requirements[len(contract.Requirements)-1]
	h.ok("evidence", "record", id, "--requirement", r.ID, "--result", "fail", "--method", "manual", "--observation", "Acceptance requirement is not met.", "--scope", "work.txt")
	h.fails("G_UNCHECKED", "complete", id, "--no-impact", "none", "--json")
	h.ok("exception", "approve", id, "--requirement", r.ID, "--reason", "Human explicitly accepts this unmet requirement", "--scope", "work.txt", "--by", "human:reviewer")
	h.ok("complete", id, "--no-impact", "none")
	var closed struct {
		Issue *domain.Issue `json:"issue"`
	}
	h.jsonOf(&closed, "show", id)
	if closed.Issue.Criteria[0].Checked || len(closed.Issue.Resolution.Approvals) != 1 {
		t.Fatal("waiver concealed the unmet criterion")
	}
}

func TestPolicyRetentionRequiresHumanAndExpiresOnAnotherChange(t *testing.T) {
	h, id, _ := approvalFixture(t)
	h.verify(id)
	old := h.read(id, "ready.md")
	h.ok("policy", "set", "--verification", "New verification policy", "--documentation", "New documentation policy")
	h.fails("G_POLICY", "complete", id, "--no-impact", "none", "--json")
	h.fails("human", "policy", "retain", id, "--reason", "Finish underway work under previous policy")
	h.fails("reason", "policy", "retain", id, "--by", "human:reviewer")
	h.ok("policy", "retain", id, "--reason", "Finish underway work under previous policy", "--by", "human:reviewer")
	if old != h.read(id, "ready.md") {
		t.Fatal("retention rewrote bound contract")
	}
	if got := h.ok("evidence", "show", id, "--json"); strings.Contains(got, `"applicable": false`) {
		t.Fatal("retained policy invalidated otherwise unchanged reports")
	}
	if got := h.ok("policy", "show", id, "--json"); !strings.Contains(got, `"retained": true`) || !strings.Contains(got, `"review_required": false`) {
		t.Fatal(got)
	}
	h.ok("policy", "set", "--verification", "Third verification policy", "--documentation", "New documentation policy")
	var projectPolicy struct {
		History []domain.CompletionPolicy `json:"history"`
	}
	h.jsonOf(&projectPolicy, "policy", "show")
	if len(projectPolicy.History) != 2 {
		t.Fatal("project policy revisions were lost")
	}
	h.fails("G_POLICY", "complete", id, "--no-impact", "none", "--json")
	h.ok("policy", "retain", id, "--reason", "Review and retain after the third policy", "--by", "human:reviewer")
	h.ok("complete", id, "--no-impact", "none")
	var closed struct {
		Issue struct {
			Resolution *domain.Resolution `json:"resolution"`
		} `json:"issue"`
	}
	h.jsonOf(&closed, "show", id)
	if len(closed.Issue.Resolution.Approvals) != 1 || closed.Issue.Resolution.Approvals[0].Kind != domain.ApprovalRetention {
		t.Fatal("completion omitted retention approval")
	}
	h.ok("fmt", "--check")
}

func TestRetainedPolicyRestoresClaimability(t *testing.T) {
	h, id, _ := approvalFixture(t)
	h.ok("release", id)
	h.ok("policy", "set", "--verification", "Changed policy", "--documentation", "Review docs")
	h.fails("G_POLICY", "claim", id, "--json")
	if ids := listIDs(h, "next"); len(ids) != 0 {
		t.Fatal("changed policy remained claimable")
	}
	h.ok("policy", "retain", id, "--reason", "Continue previously signed-off work", "--by", "human:reviewer")
	if ids := listIDs(h, "next"); len(ids) != 1 || ids[0] != id {
		t.Fatal("approved retention did not restore claimability")
	}
	h.ok("claim", id)
}

func TestRetentionDoesNotReviveWhenPolicyTextReverts(t *testing.T) {
	h, id, _ := approvalFixture(t)
	set := func(verification string) {
		h.ok("policy", "set", "--verification", verification, "--documentation", "Review docs")
	}
	set("Policy A")
	h.ok("policy", "retain", id, "--reason", "Review policy A", "--by", "human:reviewer")
	// Attribution-only re-recording is not a new policy change.
	set("Policy A")
	if got := h.ok("policy", "show", id, "--json"); !strings.Contains(got, `"retained": true`) {
		t.Fatal("unchanged methods invalidated retention")
	}
	set("Policy B")
	set("Policy A")
	h.fails("G_POLICY", "complete", id, "--no-impact", "none", "--json")
	if got := h.ok("policy", "show", id, "--json"); !strings.Contains(got, `"retained": false`) {
		t.Fatal("old retention revived after a policy reversion")
	}
}

func TestRetentionDoesNotReviveWhenInheritedDoDReverts(t *testing.T) {
	h, id, _ := approvalFixture(t)
	p := h.newIssue("--title", "Parent policy", "--kind", "manual", "--body", "Parent work")
	h.ok("edit", id, "--parent", p)
	h.ok("policy", "adopt", id)
	h.ok("dod", p, "--add", "Policy A check")
	h.ok("policy", "retain", id, "--reason", "Review added policy A check", "--by", "human:reviewer")
	h.ok("dod", p, "--remove", "Policy A check")
	h.ok("dod", p, "--add", "Policy A check")
	h.fails("G_POLICY", "complete", id, "--no-impact", "none", "--json")
}

func TestOptOutApprovalDoesNotReviveAfterRequestRecreation(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "Exception request", "--kind", "manual", "--body", "Work")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "Observed")
	request := func(reason string) { h.ok("dod", id, "--opt-out", "prep check passes", "--reason", reason) }
	request("Reason A")
	h.ok("dod", "approve", id, "--item", "prep check passes", "--reason", "Reviewed reason A", "--by", "human:reviewer")
	h.ok("ready", id)
	h.ok("dod", id, "--remove", "prep check passes")
	request("Reason B")
	h.ok("dod", id, "--remove", "prep check passes")
	request("Reason A")
	h.fails("approval", "policy", "adopt", id)
	h.fails("G_APPROVAL", "ready", id, "--json")
	h.ok("dod", "approve", id, "--item", "prep check passes", "--reason", "Reviewed recreated request", "--by", "human:reviewer")
	h.ok("policy", "adopt", id)
	h.ok("fmt", "--check")
}

func TestRetentionCannotBypassAcceptanceChangesAndAdoptionKeepsHistory(t *testing.T) {
	h, id, contract := approvalFixture(t)
	h.ok("criterion", id, "--add", "New acceptance requirement")
	h.fails("acceptance", "policy", "retain", id, "--reason", "Keep old criteria", "--by", "human:reviewer")
	h.ok("policy", "adopt", id)
	var shown struct {
		Issue struct {
			Ready *domain.Ready `json:"ready"`
		} `json:"issue"`
	}
	h.jsonOf(&shown, "show", id)
	if len(shown.Issue.Ready.PreviousContracts) != 1 || shown.Issue.Ready.PreviousContracts[0].ID != contract.ID {
		t.Fatal("adoption lost original contract")
	}
	h.ok("policy", "adopt", id)
	h.jsonOf(&shown, "show", id)
	if len(shown.Issue.Ready.PreviousContracts) != 1 {
		t.Fatal("unchanged adoption duplicated history")
	}
}

func TestRetainedInheritedDoDUsesTheBoundSnapshot(t *testing.T) {
	h, id, _ := approvalFixture(t)
	p := h.newIssue("--title", "Parent policy", "--kind", "manual", "--body", "Parent work")
	h.ok("edit", id, "--parent", p)
	h.fails("ancestry", "policy", "retain", id, "--reason", "Try to keep old ancestry", "--by", "human:reviewer")
	h.ok("policy", "adopt", id)
	h.verify(id)
	h.ok("dod", p, "--add", "New inherited check")
	h.ok("policy", "retain", id, "--reason", "Apply this new check to future work", "--by", "human:reviewer")
	h.ok("complete", id, "--no-impact", "none")
	var closed struct {
		Issue struct {
			Resolution *domain.Resolution `json:"resolution"`
		} `json:"issue"`
	}
	h.jsonOf(&closed, "show", id)
	for _, item := range closed.Issue.Resolution.DoD {
		if item == "New inherited check" {
			t.Fatal("retained completion falsely snapshotted the new policy")
		}
	}
}

func TestInheritedOptOutNeedsApprovalForExactItemAndReason(t *testing.T) {
	h := newHarness(t)
	p := h.newIssue("--title", "Parent", "--kind", "manual", "--body", "Parent work")
	id := h.newIssue("--title", "Child", "--kind", "manual", "--body", "Child work", "--parent", p)
	h.ok("dod", p, "--opt-out", "prep check passes", "--reason", "This subtree uses a different verifier")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "Observed")
	h.fails("G_APPROVAL", "ready", id, "--json")
	h.fails("human", "dod", "approve", p, "--item", "prep check passes", "--reason", "Approved substitute")
	h.ok("dod", "approve", p, "--item", "prep check passes", "--reason", "Approved substitute", "--by", "human:reviewer")
	h.ok("ready", id)
	h.ok("claim", id)
	h.ok("criterion", id, "--check", "1")
	h.verify(id)
	// Changing the requested reason must not borrow approval of an older one.
	h.ok("dod", p, "--remove", "prep check passes")
	h.ok("dod", p, "--opt-out", "prep check passes", "--reason", "Different exception")
	h.fails("G_APPROVAL", "complete", id, "--no-impact", "none", "--json")
	h.fails("approval", "policy", "adopt", id)
	h.ok("dod", "approve", p, "--item", "prep check passes", "--reason", "Reviewed new reason", "--by", "human:reviewer")
	h.ok("policy", "adopt", id)
	h.verify(id)
	h.ok("complete", id, "--no-impact", "none")
	var closed struct {
		Issue struct {
			Resolution *domain.Resolution `json:"resolution"`
		} `json:"issue"`
	}
	h.jsonOf(&closed, "show", id)
	r := closed.Issue.Resolution
	if len(r.DoDOptOuts) != 1 || r.DoDOptOuts[0].ApprovalID == "" || len(r.Approvals) != 1 || r.Approvals[0].Kind != domain.ApprovalOptOut {
		t.Fatal("resolution omitted inherited opt-out approval")
	}
	h.ok("fmt", "--check")
}
