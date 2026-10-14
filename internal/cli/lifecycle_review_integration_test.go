package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestResolutionReviewCoversSelectedKnowledgeWorkAndEntries(t *testing.T) {
	for _, changed := range []string{"knowledge-work.txt", "knowledge-checks.txt", ".prep/knowledge/overview.md"} {
		t.Run(changed, func(t *testing.T) {
			h := newHarness(t)
			id := h.newIssue("--title", "Review the combined lifecycle", "--kind", "manual", "--body", "Inspect work and its knowledge.")
			h.ok("define", id)
			h.ok("criterion", id, "--add", "Work inspected")
			h.ok("ready", id)
			h.ok("claim", id)
			h.ok("criterion", id, "--check", "1")
			for _, path := range []string{"work.txt", "verification-checks.txt", "knowledge-work.txt", "knowledge-checks.txt"} {
				if err := os.WriteFile(filepath.Join(h.dir, path), []byte("original"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var policy struct {
				Bound *domain.CompletionContract `json:"bound"`
			}
			h.jsonOf(&policy, "policy", "show", id)
			for _, requirement := range policy.Bound.Requirements {
				h.ok("evidence", "record", id, "--requirement", requirement.ID, "--result", "pass", "--method", "manual", "--observation", "Fixture work inspected.", "--scope", "work.txt", "--config-scope", "verification-checks.txt")
			}
			h.ok("knowledge", "review", "record", id, "--entry", "/overview.md", "--kind", "unchanged", "--observation", "Reviewed the overview for this work.", "--scope", "knowledge-work.txt", "--config-scope", "knowledge-checks.txt")
			h.ok("complete", id, "--docs", "/overview.md")
			h.ok("review", "approve", id, "--reason", "Reviewed the work and named knowledge.", "--by", "human:fixture")
			var view resolutionReviewView
			h.jsonOf(&view, "review", "show", id)
			if view.Status != domain.ReviewApproved || !slices.Contains(view.WorkScopes, ".prep/knowledge/overview.md") || !slices.Contains(view.WorkScopes, "knowledge-work.txt") || !slices.Contains(view.ConfigScopes, "knowledge-checks.txt") {
				t.Fatalf("review omitted selected knowledge coverage: %+v", view)
			}
			if changed == ".prep/knowledge/overview.md" {
				h.ok("knowledge", "update", "/overview.md", "--body", "Changed knowledge after resolution review.")
			} else if err := os.WriteFile(filepath.Join(h.dir, changed), []byte("changed"), 0o644); err != nil {
				t.Fatal(err)
			}
			h.jsonOf(&view, "review", "show", id)
			if view.Status != domain.ReviewStale {
				t.Fatalf("changed selected knowledge content/work/config did not invalidate approval: %+v", view)
			}
			// Reopening preserves the complete selected knowledge snapshot in
			// history, and an earlier approval cannot approve a new generation.
			h.ok("reopen", id, "--reason", "User requested follow-up on changed work.", "--by", "agent/fixture")
			var result struct {
				Issue *domain.Issue `json:"issue"`
			}
			h.jsonOf(&result, "show", id)
			if result.Issue.Resolution != nil || len(result.Issue.ResolutionHistory) != 1 || len(result.Issue.ResolutionHistory[0].KnowledgeReviews) != 1 || len(result.Issue.ResolutionReviews) != 1 {
				t.Fatalf("reopening lost the reviewed knowledge snapshot: %+v", result.Issue)
			}
			h.ok("fmt", "--check")
		})
	}
}
