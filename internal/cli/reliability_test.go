package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewSimilarIssuesBoundedAndNonBlocking(t *testing.T) {
	h := newHarness(t)
	first := h.newIssue("--title", "Payment refund reconciliation", "--kind", "code")
	for range 6 {
		h.newIssue("--title", "Payment refund webhook", "--kind", "code")
	}
	closed := h.newIssue("--title", "Payment refund duplicate", "--kind", "code")
	h.ok("drop", closed, "--reason", "already resolved")
	var result writeResult
	h.jsonOf(&result, "new", "--title", "Fix payment refunds", "--kind", "code")
	if !result.OK || result.ID == first || len(result.Similar) != 5 {
		t.Fatalf("creation/hints: %+v", result)
	}
	for _, s := range result.Similar {
		if s.ID == closed || s.ID == result.ID {
			t.Fatalf("terminal/self hint: %+v", s)
		}
	}
	if text := h.ok("new", "--title", "Payment refund events", "--kind", "code"); !strings.Contains(text, "Similar unresolved issues") {
		t.Fatalf("no text hints: %s", text)
	}
	out := h.ok("new", "--title", "Keyboard shortcuts", "--kind", "code", "--json")
	if strings.Contains(out, "similar_issues") {
		t.Fatalf("unrelated hint: %s", out)
	}
	if out := h.ok("new", "--title", "Add support", "--kind", "code"); strings.Contains(out, "Similar unresolved") {
		t.Fatalf("generic-title hint: %s", out)
	}
}

func TestPrimeReportsTagsAndBoundedBlockingFacts(t *testing.T) {
	h := newHarness(t)
	ready := h.newIssue("--title", "Pricing spelling correction", "--kind", "code", "--tag", "quick", "--body", "Correct the visible pricing typo.")
	h.ok("define", ready)
	h.ok("context", ready, "--body", "Fix the visible typo.")
	h.ok("criterion", ready, "--add", "Typo corrected.")
	h.ok("ready", ready)
	dep := h.newIssue("--title", "Authentication migration", "--kind", "code")
	for range 3 {
		h.newIssue("--title", "Hosting cutover", "--kind", "code", "--depends-on", dep)
	}
	var brief primeBrief
	h.jsonOf(&brief, "prime", "--max", "1")
	if len(brief.Next) != 1 || len(brief.Next[0].Tags) != 1 || brief.Next[0].Tags[0] != "quick" || brief.BlockedCount != 3 || len(brief.Blocked) != 1 {
		t.Fatalf("brief: %+v", brief)
	}
	out := h.ok("prime", "--max", "1")
	if !strings.Contains(out, "#quick") || !strings.Contains(out, "Blocked (unresolved dependencies)") || !strings.Contains(out, dep) || !strings.Contains(out, "2 more blocked") {
		t.Fatalf("missing facts: %s", out)
	}
	h.fails("nonnegative", "prime", "--max", "-1")
}

func TestSharedGuidancePolicyAndScope(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	for _, harness := range []string{"codex", "claude-code"} {
		out := runHook(t, "prompt", harness, hookEvent{Cwd: h.dir, Prompt: "Fix the calculator."})
		var decoded struct {
			Output struct {
				Context string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(out), &decoded); err != nil {
			t.Fatal(err)
		}
		for _, phrase := range []string{"implementation of any size", "explicitly opts out", "Questions, exploration and planning", "ask which one", "never infer effort", "requirement edits alone", "no-record-write"} {
			if !strings.Contains(decoded.Output.Context, phrase) {
				t.Fatalf("policy missing %q: %s", phrase, out)
			}
		}
		if len(decoded.Output.Context) > 1100 {
			t.Fatal("small reminder grew beyond its budget")
		}
	}
}
