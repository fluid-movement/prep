package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookFinishReviewsOnlyTurnChanges(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	h.entry("/code.md", "component", "Code", "Test code", "Current behavior.", "main.go")
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
	// A turn changes three files, the fewest finish asks about.
	change := func(body string) {
		t.Helper()
		for _, p := range []string{"main.go", "a.go", "b.go"} {
			write(p, body)
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
		change("this turn " + harness)
		out := runHook(t, "finish", harness, e)
		if !strings.Contains(out, "did this turn make or receive a decision") || !strings.Contains(out, "including one the user stated") || !strings.Contains(out, "prep decide [<issue>]") || !strings.Contains(out, "choice between options") || !strings.Contains(out, "one line") || !strings.Contains(out, "grants no additional permission") || !strings.Contains(out, "no-record-write") {
			t.Fatalf("missing review: %s", out)
		}
		for _, unwanted := range []string{"/code.md", "issue covering", "prep guide", "record actual progress", "Affected knowledge", "create one"} {
			if strings.Contains(out, unwanted) {
				t.Fatalf("unexpected tracking or descriptive knowledge text %q: %s", unwanted, out)
			}
		}
		if runHook(t, "finish", harness, e) != "" {
			t.Fatal("repeat finish on same changes")
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
		change("documented " + harness)
		h.ok("new", "--kind", "code", "--title", "Turn record "+harness)
		h.entry("/code.md", "component", "Code", "Test code", "Documents "+harness, "main.go")
		if out := runHook(t, "finish", harness, e); !strings.Contains(out, "prep decide") {
			t.Fatalf("record updates suppressed code reminder: %s", out)
		}
		runHook(t, "prompt", harness, e)
		h.ok("new", "--kind", "code", "--title", "Records only "+harness)
		h.entry("/code.md", "component", "Code", "Test code", "Records only "+harness, "main.go")
		h.ok("decide", "--title", "Records only", "--description", "No code changes", "--scope", "main.go")
		if out := runHook(t, "finish", harness, e); out != "" {
			t.Fatalf("records-only turn: %s", out)
		}
	}
	// Finish also sees changes committed during the turn.
	e := hookEvent{Cwd: h.dir, Session: "committed-turn"}
	runHook(t, "prompt", "codex", e)
	for _, p := range []string{"other.go", "other2.go", "other3.go"} {
		write(p, "new code")
	}
	git("add", "other.go", "other2.go", "other3.go")
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

func TestHookFinishScopedDecisions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	create := func(command, title, scope string, extra ...string) string {
		t.Helper()
		var out struct{ Entry string }
		args := []string{command, "--title", title, "--description", "Test record", "--scope", scope}
		h.jsonOf(&out, append(args, extra...)...)
		return strings.TrimSuffix(filepath.Base(out.Entry), ".md")
	}
	old := create("decide", "Old choice", "*.go")
	ids := map[string]string{}
	for i := 0; i < 6; i++ {
		title := fmt.Sprintf("Choice %d", i)
		var extra []string
		if i == 0 {
			extra = []string{"--supersedes", old}
		}
		ids[title] = create("decide", title, "*.go", extra...)
	}
	create("decide", "Unrelated", "docs/**")
	create("finding", "A finding", "*.go")
	h.entry("/component.md", "component", "A component", "Descriptive", "Body", "*.go")
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = h.dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git: %v %s", err, b)
	}
	e := hookEvent{Cwd: h.dir, Session: "scoped-decisions"}
	runHook(t, "prompt", "codex", e)
	for _, p := range []string{"main.go", "a.go", "b.go"} {
		if err := os.WriteFile(filepath.Join(h.dir, p), []byte("new code"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := runHook(t, "finish", "codex", e)
	for i := 0; i < 5; i++ {
		title := fmt.Sprintf("Choice %d", i)
		if !strings.Contains(out, title+" ("+ids[title]+")") {
			t.Fatalf("missing decision %s: %s", title, out)
		}
	}
	for _, unwanted := range []string{"Old choice", old, "Choice 5", "Unrelated", "A finding", "A component"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("unexpected decision %q: %s", unwanted, out)
		}
	}
	if !strings.Contains(out, "prep decide --supersedes") {
		t.Fatalf("missing supersede guidance: %s", out)
	}
}

func TestHookFinishThresholdCarriesOver(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = h.dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git: %v %s", err, b)
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(h.dir, path), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The finish events of one session, as "verb target".
	finishes := func(session string) []string {
		t.Helper()
		var got []string
		for _, ev := range hookOutcomes(t, h.dir) {
			if ev.Op == "finish" && ev.Session == session {
				got = append(got, ev.Verb+" "+ev.Target)
			}
		}
		return got
	}
	same := func(session string, want ...string) {
		t.Helper()
		if got := finishes(session); strings.Join(got, "|") != strings.Join(want, "|") {
			t.Fatalf("finish events: %v, want %v", got, want)
		}
	}

	// One file, then a second: both turns are skipped with the running count.
	e := hookEvent{Cwd: h.dir, Session: "carry"}
	runHook(t, "prompt", "claude-code", e)
	write("one.go", "1")
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("one file asked: %s", out)
	}
	same("carry", "skipped 1 files")
	runHook(t, "prompt", "claude-code", e)
	write("two.go", "2")
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("two files asked: %s", out)
	}
	same("carry", "skipped 1 files", "skipped 2 files")
	// A turn that changes nothing new is still below the threshold.
	runHook(t, "prompt", "claude-code", e)
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("two files asked: %s", out)
	}
	// The third file brings the ask, which lists the files of the skipped turns.
	runHook(t, "prompt", "claude-code", e)
	write("three.go", "3")
	out := runHook(t, "finish", "claude-code", e)
	if !strings.Contains(out, "Changed files: one.go, three.go, two.go.") {
		t.Fatalf("ask does not list the carried files: %s", out)
	}
	same("carry", "skipped 1 files", "skipped 2 files", "skipped 2 files", "asked 0 scoped decisions")
	// The ask resets the count: the next single file is skipped again.
	runHook(t, "prompt", "claude-code", e)
	write("four.go", "4")
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("count not reset by the ask: %s", out)
	}
	same("carry", "skipped 1 files", "skipped 2 files", "skipped 2 files", "asked 0 scoped decisions", "skipped 1 files")

	// A session start resets the count, a prompt does not.
	runHook(t, "session-start", "claude-code", e)
	runHook(t, "prompt", "claude-code", e)
	write("five.go", "5")
	write("six.go", "6")
	if out := runHook(t, "finish", "claude-code", e); out != "" {
		t.Fatalf("session start kept the pending snapshot: %s", out)
	}
	if got := finishes("carry"); got[len(got)-1] != "skipped 2 files" {
		t.Fatalf("finish events: %v", got)
	}

	// A turn that reverts the carried change has no changes and resets too.
	e = hookEvent{Cwd: h.dir, Session: "revert"}
	runHook(t, "prompt", "codex", e)
	write("one.go", "changed")
	if out := runHook(t, "finish", "codex", e); out != "" {
		t.Fatalf("one file asked: %s", out)
	}
	runHook(t, "prompt", "codex", e)
	write("one.go", "1")
	if out := runHook(t, "finish", "codex", e); out != "" {
		t.Fatalf("reverted change asked: %s", out)
	}
	same("revert", "skipped 1 files")
	runHook(t, "prompt", "codex", e)
	write("two.go", "changed")
	write("three.go", "changed")
	if out := runHook(t, "finish", "codex", e); out != "" {
		t.Fatalf("turn without changes kept the pending snapshot: %s", out)
	}
	same("revert", "skipped 1 files", "skipped 2 files")
	write("four.go", "changed")
	if out := runHook(t, "finish", "codex", e); !strings.Contains(out, "four.go, three.go, two.go") {
		t.Fatalf("three files not asked: %s", out)
	}
}
