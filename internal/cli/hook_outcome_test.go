package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func hookOutcomes(t *testing.T, root string) []activity.Event {
	t.Helper()
	events, err := activity.Read(filepath.Join(root, mdstore.Dir), 0)
	if err != nil {
		t.Fatal(err)
	}
	var out []activity.Event
	for _, e := range events {
		if e.Kind == activity.KindHook {
			out = append(out, e)
		}
	}
	return out
}

func TestHookOutcomesAreRecorded(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	id := h.newIssue("--title", "Record hook outcomes", "--kind", "code")
	// Finish compares turn snapshots from git state, so the project needs a repository.
	if b, err := exec.Command("git", "-C", h.dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git: %v %s", err, b)
	}
	const session = "s-outcomes"
	prompt := hookEvent{Cwd: h.dir, Session: session, Prompt: "Implement " + id}
	runHook(t, "prompt", "claude-code", prompt)
	runHook(t, "prompt", "claude-code", prompt)
	runHook(t, "before-tool", "claude-code", hookEvent{Cwd: h.dir, Session: session, Tool: "Read", Input: map[string]any{"file_path": filepath.Join(h.dir, mdstore.Dir, "knowledge", "overview.md")}})
	runHook(t, "before-tool", "claude-code", hookEvent{Cwd: h.dir, Session: session, Tool: "Write", Input: map[string]any{"file_path": filepath.Join(mdstore.Dir, "issues", id, "issue.md")}})
	// Finish asks from three changed files on.
	for _, p := range []string{"main.go", "a.go", "b.go"} {
		if err := os.WriteFile(filepath.Join(h.dir, p), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runHook(t, "finish", "claude-code", hookEvent{Cwd: h.dir, Session: session})

	got := hookOutcomes(t, h.dir)
	type want struct{ op, verb, area, target string }
	wants := []want{
		{"prompt", "sent", "", ""},
		{"prompt", "skipped", "", ""},
		{"before-tool", "denied read", activity.AreaKnowledge, filepath.Join(h.dir, mdstore.Dir, "knowledge", "overview.md")},
		{"before-tool", "denied write", activity.AreaIssue, filepath.Join(mdstore.Dir, "issues", id, "issue.md")},
		{"finish", "asked", "", "0 scoped decisions"},
	}
	if len(got) != len(wants) {
		t.Fatalf("hook events: %+v", got)
	}
	for k, w := range wants {
		e := got[k]
		if e.Op != w.op || e.Verb != w.verb || e.Area != w.area || e.Target != w.target || e.Actor != "claude-code" || e.Session != session {
			t.Errorf("event %d: %+v, want %+v", k, e, w)
		}
	}
	// Hook events never move focus, and a finish with no code change asks nothing.
	if len(activity.Foci(got)) != 0 {
		t.Fatalf("hook events moved focus: %+v", activity.Foci(got))
	}
	runHook(t, "finish", "claude-code", hookEvent{Cwd: h.dir, Session: session})
	if n := len(hookOutcomes(t, h.dir)); n != len(wants) {
		t.Fatalf("finish without changes recorded an event: %d events", n)
	}
}

func TestHookOutcomesFailOpen(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	id := h.newIssue("--title", "Hooks stay fail-open", "--kind", "code")
	// The local directory is a file, so neither the stream nor the hook state can be written.
	local := filepath.Join(h.dir, mdstore.Dir, "local")
	if err := os.RemoveAll(local); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := hookEvent{Cwd: h.dir, Session: "s-blocked", Prompt: "Implement " + id}
	if out := runHook(t, "prompt", "claude-code", session); !strings.Contains(out, "existing work") {
		t.Fatalf("prompt not sent: %s", out)
	}
	out := runHook(t, "before-tool", "claude-code", hookEvent{Cwd: h.dir, Session: "s-blocked", Tool: "Read", Input: map[string]any{"file_path": filepath.Join(h.dir, mdstore.Dir, "knowledge", "overview.md")}})
	if !strings.Contains(out, `"permissionDecision": "deny"`) {
		t.Fatalf("redirect lost: %s", out)
	}
	runHook(t, "finish", "claude-code", hookEvent{Cwd: h.dir, Session: "s-blocked"})
}

// A hook that fails after the project is opened says so in the stream, and
// still prints nothing and exits 0 (runHook fails the test otherwise).
func TestHookFailuresAreRecorded(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	h := newHarness(t)
	if b, err := exec.Command("git", "-C", h.dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git: %v %s", err, b)
	}
	const session = "s-broken"
	e := hookEvent{Cwd: h.dir, Session: session}
	runHook(t, "prompt", "claude-code", e)

	// The turn snapshot is unreadable as JSON.
	a := &app{actor: "claude-code", store: mdstore.Open(h.dir)}
	if err := os.WriteFile(hookTurnPath(a, e), []byte("{torn"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("failed finish printed: %s", out)
	}
	// The user config does not parse.
	if err := os.MkdirAll(filepath.Join(config, "prep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config, "prep", "config.yaml"), []byte("hooks: [invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := runHook(t, "prompt", "claude-code", e); out != "" {
		t.Fatalf("failed prompt printed: %s", out)
	}

	var failed []activity.Event
	for _, ev := range hookOutcomes(t, h.dir) {
		if ev.Verb == activity.HookError {
			failed = append(failed, ev)
		}
	}
	if len(failed) != 2 {
		t.Fatalf("error events: %+v", failed)
	}
	for k, w := range []struct{ op, reason string }{{"finish", "turn state: "}, {"prompt", "user config: "}} {
		ev := failed[k]
		if ev.Op != w.op || !strings.HasPrefix(ev.Target, w.reason) || len(ev.Target) == len(w.reason) || ev.Session != session || ev.Actor != "claude-code" {
			t.Errorf("event %d: %+v, want %s with reason %q", k, ev, w.op, w.reason)
		}
	}
}

// Without a repository the turn cannot be snapshotted: the prompt hook
// records that and still sends its reminder.
// A project outside a git repository has no turn snapshot; that is not a hook
// failure and must not leave an error event on every prompt.
func TestHookWithoutGitRecordsNoSnapshotError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	h := newHarness(t)
	if out := runHook(t, "prompt", "claude-code", hookEvent{Cwd: h.dir, Session: "s-no-git"}); !strings.Contains(out, "existing work") {
		t.Fatalf("prompt not sent: %s", out)
	}
	got := hookOutcomes(t, h.dir)
	if len(got) != 1 || got[0].Verb != "sent" {
		t.Fatalf("hook events: %+v", got)
	}
}

func TestHookEventsCarryAgentAndInjectedSize(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	if b, err := exec.Command("git", "-C", h.dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git: %v %s", err, b)
	}
	const session = "s-sizes"
	context := func(out string) string {
		t.Helper()
		var v struct {
			Hook struct {
				Context string `json:"additionalContext"`
				Reason  string `json:"permissionDecisionReason"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil {
			t.Fatalf("hook output %q: %v", out, err)
		}
		return v.Hook.Context + v.Hook.Reason
	}
	// The Claude Code payload names the subagent a tool call comes from.
	payload := func(tool, path, agent string) map[string]any {
		p := map[string]any{"cwd": h.dir, "session_id": session, "tool_name": tool, "tool_input": map[string]any{"file_path": path}}
		if agent != "" {
			p["agent_id"], p["agent_type"] = agent, "general-purpose"
		}
		return p
	}
	primed := context(runHook(t, "session-start", "claude-code", hookEvent{Cwd: h.dir, Session: session}))
	sent := context(runHook(t, "prompt", "claude-code", hookEvent{Cwd: h.dir, Session: session, Prompt: "Go on"}))
	runHook(t, "prompt", "claude-code", hookEvent{Cwd: h.dir, Session: session, Prompt: "Go on"})
	record := filepath.Join(h.dir, mdstore.Dir, "knowledge", "overview.md")
	denied := context(runHook(t, "before-tool", "claude-code", payload("Read", record, "agent-7")))
	runHook(t, "after-tool", "claude-code", payload("Read", filepath.Join(h.dir, "main.go"), "agent-7"))
	runHook(t, "after-tool", "claude-code", payload("Read", filepath.Join(h.dir, "main.go"), ""))
	// Finish asks from three changed files on.
	for _, p := range []string{"main.go", "a.go", "b.go"} {
		if err := os.WriteFile(filepath.Join(h.dir, p), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	asked := context(runHook(t, "finish", "claude-code", hookEvent{Cwd: h.dir, Session: session}))

	events, err := activity.Read(filepath.Join(h.dir, mdstore.Dir), 0)
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		kind, verb, agent string
		chars             int
	}
	var got []row
	for _, e := range events {
		if e.Kind == activity.KindHook || e.Kind == activity.KindTool {
			chars := e.Chars
			if e.Kind == activity.KindTool {
				chars = 0 // a tool event's chars is the size of its response
			}
			got = append(got, row{e.Kind, e.Verb, e.AgentID, chars})
		}
	}
	size := func(s string) int {
		if s == "" {
			t.Fatal("a hook added no context")
		}
		return len([]rune(s))
	}
	want := []row{
		{activity.KindHook, "primed", "", size(primed)},
		{activity.KindHook, "sent", "", size(sent)},
		{activity.KindHook, "skipped", "", 0},
		{activity.KindHook, "denied read", "agent-7", size(denied)},
		{activity.KindTool, "read", "agent-7", 0},
		{activity.KindTool, "read", "", 0},
		{activity.KindHook, "asked", "", size(asked)},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("events:\n%+v\nwant\n%+v", got, want)
	}
}
