package cli

import (
	"os"
	"os/exec"
	"path/filepath"
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
	if err := os.WriteFile(filepath.Join(h.dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
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
