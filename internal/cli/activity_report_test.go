package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func TestActivityReport(t *testing.T) {
	h := newHarness(t)
	t.Setenv("PREP_SESSION", "s1")
	id := h.newIssue("--title", "Export", "--kind", "code", "--by", "claude-code")
	h.ok("guide", id, "--by", "claude-code")
	h.ok("show", id, "--by", "claude-code")
	h.fails("no issue matches", "guide", "NOPE", "--by", "claude-code")
	h.fails("", "activity", "report", "--since", "yesterday")
	t.Setenv("PREP_SESSION", "")

	prepDir := filepath.Join(h.dir, mdstore.Dir)
	add := func(e activity.Event) {
		t.Helper()
		e.At, e.Actor = clock().UTC(), "claude-code"
		if err := activity.Append(prepDir, e); err != nil {
			t.Fatal(err)
		}
	}
	add(activity.Event{Kind: activity.KindHook, Session: "s1", Op: "prompt", Verb: "sent", Chars: 200})
	add(activity.Event{Kind: activity.KindTool, Session: "s1", Op: "Edit", Area: activity.AreaCode, Target: "main.go"})
	add(activity.Event{Kind: activity.KindHook, Session: "s1", Op: "finish", Verb: "asked", Target: "4 files"})
	add(activity.Event{Kind: activity.KindPrep, Session: "s1", Op: "decide", Area: activity.AreaKnowledge})
	add(activity.Event{Kind: activity.KindHook, Session: "s1", Op: "finish", Verb: "skipped", Target: "1 files"})
	add(activity.Event{Kind: activity.KindHook, Session: "s1", Op: "before-tool", Verb: "denied write", Target: ".prep/x"})
	add(activity.Event{Kind: activity.KindRequest, Session: "s1", Tokens: &activity.Tokens{Input: 10, Output: 5}, CostUSD: 0.5})
	add(activity.Event{Kind: activity.KindTool, Session: "s2", Op: "Write", Area: activity.AreaCode, Target: "other.go"})
	// An archived stream from before the harness clock's day.
	archived := `{"at":"2026-10-01T10:00:00Z","actor":"claude-code","kind":"hook","session":"s0","op":"prompt","verb":"sent"}
{"at":"2026-10-01T10:01:00Z","actor":"claude-code","kind":"prep","session":"s0","op":"claim","verb":"E_GATE","failed":true}
`
	if err := os.WriteFile(filepath.Join(prepDir, activity.Dir, "activity-2026-10-02.jsonl"), []byte(archived), 0o644); err != nil {
		t.Fatal(err)
	}

	out := h.ok("activity", "report")
	for _, want := range []string{
		"2 sessions, 3 turns",
		"\nHooks\n", "\nRule following\n", "\nHand-off\n", "\nCommands\n", "\nCost\n",
		"Stop question: asked 1, skipped 1, followed by a decision or finding 1 of 1 (100%)",
		"3-8 changed files:       asked 1, skipped 0, followed 1 of 1 (100%)",
		"<3 changed files:        asked 0, skipped 1, followed 0 of 0",
		"before-tool denials: 1 (0 reads, 1 writes)",
		"code edits: 2, with no issue named earlier in the session 1 of 2 (50%)",
		"per turn: 2 edited code; no issue named earlier in the session 1 of 2 (50%)",
		"per session: 2 edited code; an edit before any issue was named 1 of 2 (50%)",
		"per issue: 1 had code edits",
		"guides followed by show of the same issue: 1 of 1 (100%)",
		"briefs: 0", "hand-backs: accepted 0, refused 0", "failed checks (accepted hand-backs with no completion after): 0",
		"guide                2  1 failed",
		"guide            E_NOT_FOUND            1",
		"request events in 1 of 2 sessions",
		"reported: $0.50, 1 requests",
		"no figure for sessions and subagents without request events (1 and 0)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("report lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "E_GATE") {
		t.Fatalf("the current stream's report includes the archive:\n%s", out)
	}

	// --since reads the archives too.
	out = h.ok("activity", "report", "--since", "2026-09-30")
	if !strings.Contains(out, "3 sessions, 4 turns") || !strings.Contains(out, "claim            E_GATE                 1") {
		t.Fatalf("--since:\n%s", out)
	}
	if out = h.ok("activity", "report", "--since", "2026-10-02T00:00:00Z"); strings.Contains(out, "E_GATE") || !strings.Contains(out, "2 sessions, 3 turns") {
		t.Fatalf("--since after the archive:\n%s", out)
	}
	if out = h.ok("activity", "report", "--since", "2027-01-01"); !strings.Contains(out, "no activity recorded in this period") {
		t.Fatalf("empty period:\n%s", out)
	}

	// --json carries the same numbers.
	var r activity.Report
	h.jsonOf(&r, "activity", "report", "--since", "2026-09-30", "--json")
	if r.Sessions != 3 || r.Turns != 4 || r.Hooks.Stop != (activity.StopCount{Asked: 1, Skipped: 1, Followed: 1}) || r.Hooks.DeniedWrites != 1 {
		t.Fatalf("json hooks: %+v", r)
	}
	if r.Rules.Edits != 2 || r.Rules.EditsNoIssue != 1 || r.Rules.Turns.Edited != 2 || r.Rules.Sessions.NoIssue != 1 || r.Rules.Issues.Edited != 1 || r.Rules.GuidesThenShow != 1 {
		t.Fatalf("json rules: %+v", r.Rules)
	}
	if r.Cost.Sessions != 3 || r.Cost.SessionsWithRequests != 1 || r.Cost.CostUSD != 0.5 || r.Cost.Requests != 1 {
		t.Fatalf("json cost: %+v", r.Cost)
	}
	if len(r.Failures) != 3 || r.Handoff.RefusedBy == nil {
		t.Fatalf("json failures: %+v, handoff %+v", r.Failures, r.Handoff)
	}
	raw := h.ok("activity", "report", "--json")
	for _, key := range []string{`"stop_by_files"`, `"handbacks_refused_by_reason"`, `"failed_checks"`, `"sessions_with_requests"`, `"code_edits_without_issue"`} {
		if !strings.Contains(raw, key) {
			t.Fatalf("json lacks %s:\n%s", key, raw)
		}
	}
}
