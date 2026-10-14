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
	if err == nil || !strings.Contains(err.Error(), "changed since they were displayed") || writes != 1 {
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
