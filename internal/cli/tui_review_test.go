package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/tui"
)

func doneTUIReviewFixture(t *testing.T) (*harness, string) {
	t.Helper()
	h := newHarness(t)
	id := h.newIssue("--title", "Review me", "--kind", "code", "--body", "Review this work.")
	h.ok("define", id)
	h.enrich(id)
	h.ok("ready", id)
	h.ok("claim", id)
	h.replace(id, "acceptance.md", "- [ ]", "- [x]")
	h.verify(id, "No knowledge change was needed.")
	h.ok("complete", id, "--no-impact", "No knowledge change was needed.")
	return h, id
}

func doneTUIReviewConfigFixture(t *testing.T) (*harness, string) {
	t.Helper()
	h := newHarness(t)
	id := h.newIssue("--title", "Review configured work", "--kind", "code", "--body", "Review this configured work.")
	h.ok("define", id)
	h.enrich(id)
	h.ok("ready", id)
	h.ok("claim", id)
	h.replace(id, "acceptance.md", "- [ ]", "- [x]")
	h.verify(id, "No knowledge change was needed.")
	checks := filepath.Join(h.dir, "verification-checks.txt")
	if err := os.WriteFile(checks, []byte("check set one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Bound *domain.CompletionContract `json:"bound"`
	}
	h.jsonOf(&policy, "policy", "show", id)
	if policy.Bound == nil || len(policy.Bound.Requirements) == 0 {
		t.Fatal("fixture has no completion requirement")
	}
	requirement := policy.Bound.Requirements[0]
	h.ok("evidence", "record", id, "--requirement", requirement.ID, "--result", "pass", "--method", "manual",
		"--observation", "Inspected configured checks.", "--scope", ".prep/knowledge", "--config-scope", "verification-checks.txt", "--by", "agent/test")
	h.ok("complete", id, "--no-impact", "No knowledge change was needed.")
	return h, id
}

func writeTUIReview(root string, plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
	st := mdstore.Open(root)
	t, err := domain.Load(st, &okf.Store{Root: root}, false)
	if err != nil {
		return "", err
	}
	c, err := plan(t)
	if err != nil {
		return "", err
	}
	if err := t.CheckWrite(c); err != nil {
		return "", err
	}
	if _, err := st.Apply(c); err != nil {
		return "", err
	}
	return c.IssueID, nil
}

func TestTUIReviewApprovalWithoutCommitUsesDisplayedSnapshot(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Evidence != "" || !strings.Contains(snapshot.ChangesNote, "not in a Git work tree") {
		t.Fatalf("missing Git evidence was not reported honestly: evidence=%q note=%q", snapshot.Evidence, snapshot.ChangesNote)
	}
	if snapshot.Fingerprint == "" || snapshot.Status != domain.ReviewPending {
		t.Fatalf("snapshot = fingerprint %q status %q", snapshot.Fingerprint, snapshot.Status)
	}
	for report, current := range snapshot.KnowledgeCurrent {
		if !current {
			t.Fatalf("fixture knowledge report %s unexpectedly stale", report)
		}
	}
	writes := 0
	idOut, err := approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		writes++
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Reviewed the displayed scoped work."})
	if err != nil {
		t.Fatal(err)
	}
	if writes != 1 || idOut != id {
		t.Fatalf("approval writes=%d issue=%q", writes, idOut)
	}
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.ResolutionReviewStatus(id); got != domain.ReviewApproved {
		t.Fatalf("review status = %q", got)
	}
}

func TestTUIReviewDerivedGitEvidenceDoesNotChangeResolutionIdentity(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	beforeTree, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	before := beforeTree.Issues[id].Resolution.ContentID()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "review@example.test"}, {"config", "user.name", "Review Test"}, {"add", "."}, {"commit", "-qm", "fixture completion"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = h.dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Evidence == "" || snapshot.Issue.Resolution.Evidence != "" {
		t.Fatalf("derived evidence=%q mutated recorded evidence=%q", snapshot.Evidence, snapshot.Issue.Resolution.Evidence)
	}
	if got := snapshot.Issue.Resolution.ContentID(); got != before || snapshot.ResolutionID != before {
		t.Fatalf("derived evidence changed resolution identity: before=%s issue=%s snapshot=%s", before, got, snapshot.ResolutionID)
	}
	if !strings.Contains(snapshot.Changes, "fixture completion") {
		t.Fatalf("commit diff not included in review: %s", snapshot.Changes)
	}
}

func TestTUIReviewLegacyRequiresExplicitScope(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	// Produce a legacy shape by creating a done resolution without scoped
	// applicability: the manual resolution's knowledge contract still applies,
	// so use a separate handcrafted completed resolution fixture with empty scope.
	path := h.path(id, "resolution.md")
	text := "---\noutcome: done\nby: agent/test\nat: 2026-10-05T12:00:00Z\n---\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	without, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if without.Fingerprint != "" || !strings.Contains(without.ChangesNote, "explicit scope") {
		t.Fatalf("legacy review snapshot = %+v", without)
	}
	if _, err := approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: without, Reason: "Reviewed."}); err == nil || !strings.Contains(err.Error(), "explicit legacy scopes") {
		t.Fatalf("empty-scope approval error = %v", err)
	}
	legacyPath := filepath.Join(h.dir, "legacy.txt")
	if err := os.WriteFile(legacyPath, []byte("tracked by fingerprint\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	with, err := loadTUIReview(h.dir, id, []string{"legacy.txt"})
	if err != nil || with.Fingerprint == "" || !slices.Contains(with.WorkScopes, "legacy.txt") {
		t.Fatalf("explicit-scope snapshot = %+v, err %v", with, err)
	}
}

func TestTUIReviewApprovalRejectsChangedScopeAndBadAttribution(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	// The report includes .prep/knowledge, so this edit invalidates the exact
	// snapshot displayed to the reviewer.
	entry := filepath.Join(h.dir, ".prep", "knowledge", "overview.md")
	contents, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, append(contents, []byte("\nChanged after display.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	currentSnapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	for report, current := range currentSnapshot.KnowledgeCurrent {
		if current {
			t.Fatalf("knowledge report %s remained current after its entry changed", report)
		}
	}
	writes := 0
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		writes++
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Reviewed."})
	if err == nil || !strings.Contains(err.Error(), "knowledge-review applicability changed") || writes != 1 {
		t.Fatalf("changed-scope approval err=%v writes=%d", err, writes)
	}
	if _, err := approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "agent/test", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Reviewed."}); err == nil || !strings.Contains(err.Error(), "human:<name>") {
		t.Fatalf("agent attribution error = %v", err)
	}
	if _, err := approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "  "}); err == nil || !strings.Contains(err.Error(), "requires a reason") {
		t.Fatalf("blank reason error = %v", err)
	}
}

func TestTUIReviewRejectsCodeScopeChangedAfterDisplay(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	codePath := filepath.Join(h.dir, "src", "reviewed.go")
	if err := os.MkdirAll(filepath.Dir(codePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codePath, []byte("package reviewed\n\nconst Version = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot, err := loadTUIReview(h.dir, id, []string{"src/reviewed.go"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(snapshot.WorkScopes, "src/reviewed.go") {
		t.Fatalf("explicit code scope was not captured: %v", snapshot.WorkScopes)
	}
	if err := os.WriteFile(codePath, []byte("package reviewed\n\nconst Version = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale code view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "review scopes changed since they were displayed") {
		t.Fatalf("changed code scope error = %v", err)
	}
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.Issues[id].ResolutionReviews); got != 0 {
		t.Fatalf("stale code view appended an approval; history length=%d", got)
	}
}

func TestTUIReviewRejectsConfigScopeChangedAfterDisplay(t *testing.T) {
	h, id := doneTUIReviewConfigFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(snapshot.ConfigScopes, "verification-checks.txt") {
		t.Fatalf("resolution config scope not captured: %v", snapshot.ConfigScopes)
	}
	checks := filepath.Join(h.dir, "verification-checks.txt")
	if err := os.WriteFile(checks, []byte("check set two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale configuration view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "review scopes changed since they were displayed") {
		t.Fatalf("changed config scope error = %v", err)
	}
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.Issues[id].ResolutionReviews); got != 0 {
		t.Fatalf("stale configuration view appended an approval; history length=%d", got)
	}
}

func TestTUIReviewRejectsResolutionChangedAfterDisplay(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ok("reopen", id, "--reason", "Agent will revise after feedback.", "--by", "agent/test")
	writes := 0
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		writes++
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Reviewed."})
	if err == nil || !strings.Contains(err.Error(), "no longer has a done resolution") || writes != 1 {
		t.Fatalf("resolution-change approval err=%v writes=%d", err, writes)
	}
}

func TestTUIReviewRejectsStatusChangedByAnotherApproval(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	writer := func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}
	if _, err := approveTUIReview(h.dir, writer, "human:first-reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "First reviewer approved."}); err != nil {
		t.Fatal(err)
	}
	_, err = approveTUIReview(h.dir, writer, "human:second-reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "resolution review status changed from pending to approved") {
		t.Fatalf("stale pending snapshot error = %v", err)
	}
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.Issues[id].ResolutionReviews); got != 1 {
		t.Fatalf("stale view appended another approval; history length=%d", got)
	}
}

func TestTUIReviewRejectsKnowledgeApplicabilityChangedByPolicy(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.KnowledgeCurrent) == 0 {
		t.Fatal("fixture has no selected knowledge reports")
	}
	h.ok("policy", "set", "--verification", "Use the revised verification policy.", "--documentation", "Update affected knowledge or explain no impact.")
	current, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for report, wasCurrent := range snapshot.KnowledgeCurrent {
		changed = changed || wasCurrent != current.KnowledgeCurrent[report]
	}
	if !changed {
		t.Fatal("policy update did not change selected report applicability")
	}
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale policy view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "current completion contract changed") {
		t.Fatalf("stale knowledge applicability error = %v", err)
	}
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.Issues[id].ResolutionReviews); got != 0 {
		t.Fatalf("stale policy view appended an approval; history length=%d", got)
	}
}

func TestTUIReviewRejectsRequirementEditAfterDisplay(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	snapshot, err := loadTUIReview(h.dir, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	h.replace(id, "issue.md", "Review this work.", "Review this work and its consequences.")
	loaded, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Issues[id].Resolution.ContentID() != snapshot.ResolutionID {
		t.Fatal("fixture changed the resolution identity instead of only issue display content")
	}
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale requirement view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "issue requirements or acceptance content changed") {
		t.Fatalf("stale requirement error = %v", err)
	}
	if got := len(loaded.Issues[id].ResolutionReviews); got != 0 {
		t.Fatalf("stale requirement view appended an approval; history length=%d", got)
	}
}

func TestTUIReviewRejectsCurrentPolicyChangeForLegacyResolution(t *testing.T) {
	h, id := doneTUIReviewFixture(t)
	// Historical resolutions without knowledge reports still capture the
	// effective contract shown to the reviewer.
	if err := os.WriteFile(h.path(id, "resolution.md"), []byte("---\noutcome: done\nby: agent/test\nat: 2026-10-05T12:00:00Z\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	legacyPath := filepath.Join(h.dir, "legacy.txt")
	if err := os.WriteFile(legacyPath, []byte("legacy work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot, err := loadTUIReview(h.dir, id, []string{"legacy.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.KnowledgeCurrent) != 0 || snapshot.CurrentContractID == "" {
		t.Fatalf("legacy snapshot contract=%q knowledge=%v", snapshot.CurrentContractID, snapshot.KnowledgeCurrent)
	}
	h.ok("policy", "set", "--verification", "A new policy for legacy work.", "--documentation", "Update affected knowledge or explain no impact.")
	current, err := loadTUIReviewTree(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	if current.Issues[id].Resolution.ContentID() != snapshot.ResolutionID {
		t.Fatal("policy update changed the legacy resolution identity")
	}
	_, err = approveTUIReview(h.dir, func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		return writeTUIReview(h.dir, plan)
	}, "human:reviewer", tui.ReviewApprovalRequest{Snapshot: snapshot, Reason: "Stale policy view should be refused."})
	if err == nil || !strings.Contains(err.Error(), "current completion contract changed") {
		t.Fatalf("stale legacy policy error = %v", err)
	}
	if got := len(current.Issues[id].ResolutionReviews); got != 0 {
		t.Fatalf("stale policy view appended an approval; history length=%d", got)
	}
}
