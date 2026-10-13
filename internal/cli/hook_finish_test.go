package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookFinishReviewsOnlyTurnChanges(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	h.ok("knowledge", "new", "/code.md", "--type", "component", "--title", "Code", "--description", "Test code", "--scope", "main.go", "--body", "Current behavior.")
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = h.dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, b)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(h.dir, path), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.go", "original")
	git("init", "-q")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.invalid")
	git("add", ".")
	git("commit", "-qm", "baseline")
	// Old dirty work must not cause a finish reminder.
	write("main.go", "old dirty work")
	for _, harness := range []string{"claude-code", "codex"} {
		e := hookEvent{Cwd: h.dir, Session: harness + "-session"}
		runHook(t, "prompt", harness, e)
		if out := runHook(t, "finish", harness, e); out != "" {
			t.Fatalf("old dirty work: %s", out)
		}
		write("main.go", "this turn "+harness)
		out := runHook(t, "finish", harness, e)
		if !strings.Contains(out, "Before finishing") || !strings.Contains(out, "/code.md") || !strings.Contains(out, "issue covering") {
			t.Fatalf("missing review: %s", out)
		}
		if harness == "codex" && !strings.Contains(out, `"decision": "block"`) {
			t.Fatal("Codex Stop did not continue")
		}
		if harness == "claude-code" && !strings.Contains(out, `"hookEventName": "Stop"`) {
			t.Fatal("Claude Stop lacks context")
		}
		e.StopActive = true
		write("main.go", "during review")
		if runHook(t, "finish", harness, e) != "" {
			t.Fatal("stop loop")
		}
		e.StopActive = false
		runHook(t, "prompt", harness, e)
		write("main.go", "documented "+harness)
		h.ok("new", "--kind", "code", "--title", "Turn record "+harness)
		h.ok("knowledge", "update", "/code.md", "--body", "Documents "+harness)
		if out := runHook(t, "finish", harness, e); out != "" {
			t.Fatalf("already reconciled: %s", out)
		}
	}
	// Finish also sees changes committed during the turn.
	e := hookEvent{Cwd: h.dir, Session: "committed-turn"}
	runHook(t, "prompt", "codex", e)
	write("other.go", "new code")
	git("add", "other.go")
	git("commit", "-qm", "code during turn")
	if out := runHook(t, "finish", "codex", e); !strings.Contains(out, "other.go") {
		t.Fatalf("missed committed change: %s", out)
	}
	if runHook(t, "finish", "codex", e) != "" {
		t.Fatal("repeat finish on same changes")
	}
}

func TestHookFinishWithoutBaselineFailsOpen(t *testing.T) {
	h := newHarness(t)
	if runHook(t, "finish", "codex", hookEvent{Cwd: h.dir, Session: "unknown"}) != "" {
		t.Fatal("missing baseline did not fail open")
	}
}
