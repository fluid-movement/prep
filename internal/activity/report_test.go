package activity

import (
	"reflect"
	"testing"
	"time"
)

var reportDay = time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)

// stream builds events one minute apart from reportDay.
func stream(events ...Event) []Event {
	for k := range events {
		events[k].At = reportDay.Add(time.Duration(k) * time.Minute)
		if events[k].Actor == "" {
			events[k].Actor = "claude-code"
		}
	}
	return events
}

func hook(session, op, verb, target string) Event {
	return Event{Kind: KindHook, Session: session, Op: op, Verb: verb, Target: target}
}

func prep(session, op, issue string) Event {
	return Event{Kind: KindPrep, Session: session, Op: op, Issue: issue}
}

func edit(session string) Event {
	return Event{Kind: KindTool, Session: session, Op: "Edit", Area: AreaCode, Target: "main.go"}
}

func TestReportStopQuestions(t *testing.T) {
	r := BuildReport(stream(
		hook("s1", "prompt", "sent", ""),
		hook("s1", "finish", "asked", "2 files"),
		prep("s1", "decide", ""), // followed, fewer than three files
		hook("s1", "prompt", "skipped", ""),
		hook("s1", "finish", "asked", "5 files"),
		hook("s1", "prompt", "skipped", ""),
		prep("s1", "decide", ""), // after the next prompt: not a follow-up
		hook("s1", "finish", "asked", "8 files"),
		prep("s2", "finding", ""), // another session
		prep("s1", "finding", ""), // followed, three to eight
		hook("s1", "prompt", "skipped", ""),
		hook("s1", "finish", "asked", "12 files"),
		Event{Kind: KindPrep, Session: "s1", Op: "decide", Verb: "E_USAGE", Failed: true}, // a failed decide records nothing
		hook("s1", "prompt", "skipped", ""),
		hook("s1", "finish", "asked", "0 scoped decisions"), // no file count
		hook("s1", "finish", "skipped", "2 files"),
		hook("s1", "finish", "skipped", "9 files"),
	), time.Time{})
	if got, want := r.Hooks.Stop, (StopCount{Asked: 5, Skipped: 2, Followed: 2}); got != want {
		t.Fatalf("stop: %+v, want %+v", got, want)
	}
	want := []StopBucket{
		{"<3", StopCount{Asked: 1, Skipped: 1, Followed: 1}},
		{"3-8", StopCount{Asked: 2, Followed: 1}},
		{">8", StopCount{Asked: 1, Skipped: 1}},
		{"unknown", StopCount{Asked: 1}},
	}
	if !reflect.DeepEqual(r.Hooks.StopByFiles, want) {
		t.Fatalf("by files: %+v\nwant %+v", r.Hooks.StopByFiles, want)
	}
	// Five prompts in s1, and the commands of s2, which has no prompt event.
	if r.Hooks.PromptsSent != 1 || r.Hooks.PromptsSkipped != 4 || r.Turns != 6 {
		t.Fatalf("prompts sent %d, skipped %d, turns %d", r.Hooks.PromptsSent, r.Hooks.PromptsSkipped, r.Turns)
	}
}

func TestReportHookOutcomes(t *testing.T) {
	r := BuildReport(stream(
		Event{Kind: KindHook, Session: "s1", Op: "session-start", Verb: "primed", Chars: 500},
		Event{Kind: KindHook, Session: "s1", Op: "prompt", Verb: "sent", Chars: 100},
		hook("s1", "before-tool", "denied read", ".prep/issues/x/issue.md"),
		hook("s1", "before-tool", "denied write", ".prep/issues/x/issue.md"),
		hook("s1", "before-tool", "denied write", ".prep/issues/y/issue.md"),
		Event{Kind: KindHook, Session: "s1", Op: "before-tool", Verb: "context", Chars: 40},
		hook("s1", "finish", HookError, "snapshot: git failed"),
		hook("s1", "prompt", HookError, "read: bad json"),
		hook("s1", "finish", HookError, "snapshot: git failed"),
	), time.Time{})
	h := r.Hooks
	if h.DeniedReads != 1 || h.DeniedWrites != 2 || h.EditNotices != 1 || h.Errors != 3 || h.InjectedChars != 640 {
		t.Fatalf("hooks: %+v", h)
	}
	if want := []NameCount{{"finish", 2}, {"prompt", 1}}; !reflect.DeepEqual(h.ErrorsByHook, want) {
		t.Fatalf("errors by hook: %+v", h.ErrorsByHook)
	}
	// A prompt whose hook failed still starts a turn.
	if r.Turns != 2 || r.Sessions != 1 {
		t.Fatalf("turns %d, sessions %d", r.Turns, r.Sessions)
	}
}

func TestReportRuleFollowing(t *testing.T) {
	r := BuildReport(stream(
		// s1: edits before any issue, then names A and edits in later turns.
		hook("s1", "prompt", "sent", ""),
		edit("s1"), // no issue, no prep command
		edit("s1"),
		hook("s1", "prompt", "skipped", ""),
		prep("s1", "guide", "A"),
		prep("s1", "show", "A"), // guide then show
		prep("s1", "knowledge find", ""),
		edit("s1"), // issue and knowledge named in the turn, belongs to A
		hook("s1", "prompt", "skipped", ""),
		edit("s1"), // issue named in an earlier turn only
		hook("s1", "prompt", "skipped", ""),
		prep("s1", "list", ""), // a turn without edits
		// s2: knowledge, then guide B (no show), then edits.
		hook("s2", "prompt", "sent", ""),
		prep("s2", "knowledge show", ""),
		prep("s2", "guide", "B"),
		edit("s2"),              // knowledge came before B was named
		prep("s2", "show", "A"), // not the guided issue
		Event{Kind: KindTool, Session: "s2", Op: "Bash", Area: AreaCode, Target: "go test"}, // not an edit
		Event{Kind: KindTool, Session: "s2", Op: "Write", Area: AreaKnowledge, Target: "x"}, // not code
		// s3: a failed guide still names its issue; apply_patch is an edit.
		Event{Kind: KindPrep, Session: "s3", Op: "guide", Issue: "C", Verb: "E_NOT_FOUND", Failed: true},
		Event{Kind: KindTool, Session: "s3", Op: "apply_patch", Area: AreaCode, Actor: "codex"},
	), time.Time{})
	ru := r.Rules
	if ru.Edits != 6 || ru.EditsNoIssue != 2 {
		t.Fatalf("edits %d, without issue %d", ru.Edits, ru.EditsNoIssue)
	}
	if want := (RuleTurns{Edited: 5, NoIssue: 1, IssueInTurn: 3, KnowledgeInTurn: 2, NoPrep: 2}); ru.Turns != want {
		t.Fatalf("turns: %+v\nwant  %+v", ru.Turns, want)
	}
	if want := (RuleSessions{Edited: 3, NoIssue: 1, Knowledge: 1, Guides: 2, GuidesThenShow: 1}); ru.Sessions != want {
		t.Fatalf("sessions: %+v\nwant     %+v", ru.Sessions, want)
	}
	if want := (RuleIssues{Edited: 3, Knowledge: 1, Guided: 2, GuidesThenShow: 1}); ru.Issues != want {
		t.Fatalf("issues: %+v\nwant   %+v", ru.Issues, want)
	}
	if ru.Guides != 2 || ru.GuidesThenShow != 1 {
		t.Fatalf("guides %d, then show %d", ru.Guides, ru.GuidesThenShow)
	}
	if r.Sessions != 3 || r.Turns != 6 {
		t.Fatalf("sessions %d, turns %d", r.Sessions, r.Turns)
	}
}

func TestReportCommandsAndHandoff(t *testing.T) {
	fail := func(op, issue, code string) Event {
		return Event{Kind: KindPrep, Op: op, Issue: issue, Verb: code, Failed: true}
	}
	r := BuildReport(stream(
		prep("", "guide", "A"),
		prep("", "guide", "B"),
		fail("guide", "", "E_NOT_FOUND"),
		fail("claim", "A", "E_GATE"),
		fail("claim", "B", "E_GATE"),
		fail("claim", "", "E_USAGE"),
		prep("", "brief", "A"),
		prep("", "brief", "B"),
		prep("", "handback", "A"),
		prep("", "complete", "A"),
		prep("", "handback", "B"), // never completed: a failed check
		fail("handback", "B", "E_GATE"),
		fail("handback", "B", "E_GATE"),
		fail("handback", "B", "E_USAGE"),
		Event{Kind: KindFocus, Issue: "A"},
	), time.Time{})
	want := []CommandCount{{"handback", 5, 3}, {"claim", 3, 3}, {"guide", 3, 1}, {"brief", 2, 0}, {"complete", 1, 0}, {"focus", 1, 0}}
	if !reflect.DeepEqual(r.Commands, want) {
		t.Fatalf("commands: %+v", r.Commands)
	}
	wantFail := []FailureCount{{"claim", "E_GATE", 2}, {"handback", "E_GATE", 2}, {"claim", "E_USAGE", 1}, {"guide", "E_NOT_FOUND", 1}, {"handback", "E_USAGE", 1}}
	if !reflect.DeepEqual(r.Failures, wantFail) {
		t.Fatalf("failures: %+v", r.Failures)
	}
	ho := r.Handoff
	if ho.Briefs != 2 || ho.Accepted != 2 || ho.Refused != 3 || ho.FailedChecks != 1 {
		t.Fatalf("handoff: %+v", ho)
	}
	if want := []NameCount{{"E_GATE", 2}, {"E_USAGE", 1}}; !reflect.DeepEqual(ho.RefusedBy, want) {
		t.Fatalf("refused by: %+v", ho.RefusedBy)
	}
	if r.Unattributed != 15 || r.Sessions != 0 {
		t.Fatalf("unattributed %d, sessions %d", r.Unattributed, r.Sessions)
	}
	// Without the commands the section is zero, not missing.
	if empty := BuildReport(nil, time.Time{}); empty.Handoff.Briefs != 0 || empty.Handoff.RefusedBy == nil || empty.Commands == nil {
		t.Fatalf("empty: %+v", empty)
	}
}

func TestReportCost(t *testing.T) {
	req := func(session, agent string, cost float64) Event {
		return Event{Kind: KindRequest, Session: session, AgentID: agent, Tokens: &Tokens{Input: 10, Output: 5, CacheRead: 100, CacheWrite: 1}, CostUSD: cost}
	}
	r := BuildReport(stream(
		req("s1", "", 0.25),
		req("s1", "", 0.5),
		Event{Kind: KindRequest, Session: "s1", Context: &Context{Tokens: 4000, Window: 200000}, CostUSD: 0.25}, // a turn's end: no model request
		Event{Kind: KindTool, Session: "s1", AgentID: "sub1", Op: "Read", Area: AreaCode},
		Event{Kind: KindTool, Session: "s1", AgentID: "sub2", Op: "Read", Area: AreaCode},
		prep("s2", "guide", "A"),
		Event{Kind: KindTool, Session: "s3", Op: "apply_patch", Area: AreaCode, Actor: "codex"},
	), time.Time{})
	c := r.Cost
	if c.Sessions != 3 || c.SessionsWithRequests != 1 || c.Subagents != 2 || c.SubagentsWithRequests != 0 {
		t.Fatalf("coverage: %+v", c)
	}
	if c.Requests != 2 || c.CostUSD != 1 || c.Tokens != (Tokens{Input: 20, Output: 10, CacheRead: 200, CacheWrite: 2}) || c.ContextUsed != 4000 {
		t.Fatalf("usage: %+v", c.Usage)
	}
	if want := []ActorCover{{"claude-code", 2, 1}, {"codex", 1, 0}}; !reflect.DeepEqual(c.ByActor, want) {
		t.Fatalf("by actor: %+v", c.ByActor)
	}
}

func TestReportSince(t *testing.T) {
	events := stream(
		hook("s1", "prompt", "sent", ""),
		prep("s1", "guide", "A"),
		edit("s1"),
		hook("s1", "finish", "asked", "1 files"),
		// The period starts here.
		hook("s1", "prompt", "skipped", ""),
		edit("s1"),
		prep("s1", "decide", ""), // the question before the period is not followed up in it
		prep("s2", "list", ""),
	)
	since := events[4].At
	r := BuildReport(events, since)
	if r.Events != 4 || r.Sessions != 2 || r.Turns != 2 || !r.Since.Equal(since) || !r.From.Equal(since) || !r.To.Equal(events[7].At) {
		t.Fatalf("period: %+v", r)
	}
	// The issue named before the period still counts as named.
	if ru := r.Rules; ru.Edits != 1 || ru.EditsNoIssue != 0 || ru.Turns.Edited != 1 || ru.Turns.NoIssue != 0 || ru.Turns.IssueInTurn != 0 || ru.Guides != 0 || ru.Issues.Edited != 1 {
		t.Fatalf("rules: %+v", r.Rules)
	}
	if r.Hooks.Stop != (StopCount{}) || r.Hooks.PromptsSkipped != 1 || r.Hooks.PromptsSent != 0 {
		t.Fatalf("hooks: %+v", r.Hooks)
	}
	if want := []CommandCount{{"decide", 1, 0}, {"list", 1, 0}}; !reflect.DeepEqual(r.Commands, want) {
		t.Fatalf("commands: %+v", r.Commands)
	}
}
