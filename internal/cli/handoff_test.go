package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creack/pty"
	"github.com/fluid-movement/prep/internal/gitx"
)

func handoffGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := gitx.Run(dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func handoffReady(h *harness, kind string) string {
	h.t.Helper()
	id := h.newIssue("--title", "Cloud task", "--kind", kind, "--body", "Do this task.")
	h.ok("define", id)
	h.ok("criterion", id, "--add", "it works")
	h.ok("context", id, "--body", "Change `internal/x.go`.")
	h.ok("ready", id)
	return id
}

func handoffRemote(t *testing.T, h *harness) {
	t.Helper()
	remote := t.TempDir()
	handoffGit(t, remote, "init", "--bare")
	handoffGit(t, h.dir, "init", "-b", "main")
	handoffGit(t, h.dir, "config", "user.email", "test@example.org")
	handoffGit(t, h.dir, "config", "user.name", "Test")
	handoffGit(t, h.dir, "remote", "add", "origin", remote)
	handoffPush(t, h)
}

func handoffPush(t *testing.T, h *harness) {
	t.Helper()
	handoffGit(t, h.dir, "add", ".")
	handoffGit(t, h.dir, "commit", "-m", "fixture")
	handoffGit(t, h.dir, "push", "-u", "origin", "HEAD")
}

func fakeCloud(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	calls := filepath.Join(bin, "calls")
	t.Setenv("CLOUD_CALLS", calls)
	script := `#!/bin/sh
test -t 0 && test -t 1 || exit 9
test "$1" = --cloud || exit 8
printf '%s\n' "$2" >> "$CLOUD_CALLS"
printf 'Created cloud session: session_test\nView: https://example.org/session_test\n'
`
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return calls
}

func handoffTerminal(t *testing.T, h *harness, answer string, args ...string) (int, string) {
	t.Helper()
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()
	if answer != "" {
		if _, err := master.Write([]byte(answer + "\n")); err != nil {
			t.Fatal(err)
		}
	}
	var out, errb bytes.Buffer
	code := Main(append([]string{"--root", h.dir}, args...), slave, &out, &errb)
	return code, out.String() + errb.String()
}

func TestHandoffStart(t *testing.T) {
	h := newHarness(t)
	codeID := handoffReady(h, "code")
	researchID := handoffReady(h, "research")
	handoffRemote(t, h)
	calls := fakeCloud(t)
	code, out := handoffTerminal(t, h, "", "handoff", "start", codeID, researchID, "--yes")
	if code != 0 || !strings.Contains(out, "About to start 2 cloud sessions") {
		t.Fatalf("start: %d %s", code, out)
	}
	for _, id := range []string{codeID, researchID} {
		if !strings.Contains(out, id+" session_test https://example.org/session_test") {
			t.Fatalf("missing session: %s", out)
		}
		// show omits the work log while an issue is still ready.
		history := h.read(id, "history.md")
		if !strings.Contains(history, "Cloud session: session_test https://example.org/session_test") {
			t.Fatalf("missing log: %s", history)
		}
		h.expectState(id, "ready", false)
	}
	text, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"prep guide " + codeID, "prep guide " + researchID, "commit the work AND prep records", "push your branch", "Do not complete", "Cloud handback ", "## Summary", "## Criteria", "## Findings"} {
		if !strings.Contains(string(text), want) {
			t.Fatalf("prompt lacks %q: %s", want, text)
		}
	}
	// No push: the logs are local, and the remote still points at the base.
	if got := handoffGit(t, h.dir, "status", "--porcelain"); !strings.Contains(got, "history.md") {
		t.Fatalf("logs not staged: %s", got)
	}
}

func TestHandoffStartPreflight(t *testing.T) {
	for _, scenario := range []string{"open", "stale", "blocked", "claimed", "unpushed", "no upstream", "manual", "duplicate"} {
		t.Run(scenario, func(t *testing.T) {
			h := newHarness(t)
			good := handoffReady(h, "code")
			id := handoffReady(h, "code")
			want := ""
			switch scenario {
			case "open":
				id = h.newIssue("--title", "Open", "--kind", "code", "--body", "Not ready.")
				want = "it is open"
			case "stale":
				h.ok("edit", id, "--body", "Changed requirement.")
				want = "requirement changed"
			case "blocked":
				h.ok("edit", id, "--depends-on", good)
				want = "unresolved dependencies"
			case "claimed":
				h.ok("claim", id)
				want = "unclaimed"
			case "manual":
				id = handoffReady(h, "manual")
				want = "code or research"
			case "duplicate":
				id = good
				want = "duplicate issue"
			}
			handoffRemote(t, h)
			switch scenario {
			case "unpushed":
				h.ok("log", id, "local change")
				handoffGit(t, h.dir, "commit", "-am", "not pushed")
				want = "HEAD is not pushed"
			case "no upstream":
				handoffGit(t, h.dir, "branch", "--unset-upstream")
				want = "no upstream"
			}
			calls := fakeCloud(t)
			code, out := handoffTerminal(t, h, "", "handoff", "start", good, id, "--yes")
			if code == 0 || !strings.Contains(out, want) {
				t.Fatalf("preflight: %d %s (want %s)", code, out, want)
			}
			if _, err := os.Stat(calls); !os.IsNotExist(err) {
				t.Fatalf("claude ran before preflight finished: %v", err)
			}
		})
	}
}

func TestHandoffStartTerminalAndConfirmation(t *testing.T) {
	h := newHarness(t)
	id := handoffReady(h, "code")
	handoffRemote(t, h)
	calls := fakeCloud(t)
	h.fails("claude --cloud 'Run prep guide "+id, "handoff", "start", id, "--yes")
	code, out := handoffTerminal(t, h, "n", "handoff", "start", id)
	if code == 0 || !strings.Contains(out, "no sessions started") {
		t.Fatalf("confirmation: %d %s", code, out)
	}
	if _, err := os.Stat(calls); !os.IsNotExist(err) {
		t.Fatal("claude ran without confirmation")
	}
	code, out = handoffTerminal(t, h, "yes", "handoff", "start", id)
	if code != 0 || !strings.Contains(out, "session_test") {
		t.Fatalf("confirmed: %d %s", code, out)
	}
}

func handoffBranch(t *testing.T, h *harness, id, branch, report string) {
	t.Helper()
	handoffGit(t, h.dir, "switch", "-c", branch, "main")
	h.ok("claim", id)
	if report != "" {
		h.ok("finding", id, "--title", "Cloud handback "+id, "--description", "Cloud implementation report", "--body", "## Summary\n"+report)
	}
	handoffPush(t, h)
	handoffGit(t, h.dir, "switch", "main")
}

func TestHandoffFetch(t *testing.T) {
	h := newHarness(t)
	a, b := handoffReady(h, "code"), handoffReady(h, "code")
	handoffRemote(t, h)
	handoffBranch(t, h, a, "claude/a", "Implemented a.")
	handoffBranch(t, h, b, "claude/b", "")
	handoffGit(t, h.dir, "switch", "-c", "other")
	h.ok("log", a, "not a cloud branch")
	handoffPush(t, h)
	handoffGit(t, h.dir, "switch", "main")
	out := h.ok("handoff", "fetch")
	for _, want := range []string{"claude/a " + a + " (report present)", "claude/b " + b + " (report missing)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("fetch lacks %s: %s", want, out)
		}
	}
	if strings.Contains(out, "other") {
		t.Fatalf("listed non-cloud branch: %s", out)
	}
	var result struct {
		Branches []cloudBranch `json:"branches"`
	}
	h.jsonOf(&result, "handoff", "fetch", "--merge")
	if len(result.Branches) != 2 || !result.Branches[0].Merged || !result.Branches[1].Merged {
		t.Fatalf("merges: %+v", result)
	}
	h.ok("check")
	h.ok("fmt", "--check")
	if out := h.ok("handoff", "fetch"); out != "" {
		t.Fatalf("merged branches listed: %s", out)
	}
	if main := handoffGit(t, h.dir, "rev-parse", "origin/main"); main == handoffGit(t, h.dir, "rev-parse", "HEAD") {
		t.Fatal("fetch --merge pushed main")
	}
	// A merge commit has exactly two parents, never an octopus.
	if parents := strings.Fields(handoffGit(t, h.dir, "show", "-s", "--format=%P")); len(parents) != 2 {
		t.Fatalf("parents: %v", parents)
	}
}

func TestHandoffFetchStopsAtMergeFailure(t *testing.T) {
	h := newHarness(t)
	id := handoffReady(h, "code")
	handoffRemote(t, h)
	handoffBranch(t, h, id, "claude/a", "first")
	handoffBranch(t, h, id, "claude/b", "conflict")
	third := handoffReady(h, "code")
	handoffPush(t, h)
	handoffBranch(t, h, third, "claude/c", "must not merge")
	h.fails("merge failed, stopped", "handoff", "fetch", "--merge")
	if _, err := gitx.Run(h.dir, "merge-base", "--is-ancestor", "origin/claude/c", "HEAD"); err == nil {
		t.Fatal("merged after failure")
	}
}

func TestHandoffFetchStopsAtCheckFailure(t *testing.T) {
	h := newHarness(t)
	id, next := handoffReady(h, "code"), handoffReady(h, "code")
	handoffRemote(t, h)
	handoffGit(t, h.dir, "switch", "-c", "claude/a")
	// Removing a dependency on a branch leaves invalid records to exercise
	// the post-merge check through the real command path.
	h.ok("edit", id, "--depends-on", next)
	handoffGit(t, h.dir, "rm", "-r", ".prep/issues/"+next)
	handoffPush(t, h)
	handoffGit(t, h.dir, "switch", "main")
	handoffBranch(t, h, next, "claude/b", "must not merge")
	h.fails("prep check failed, stopped", "handoff", "fetch", "--merge")
	if _, err := gitx.Run(h.dir, "merge-base", "--is-ancestor", "origin/claude/b", "HEAD"); err == nil {
		t.Fatal("merged after validation failure")
	}
}

func TestHandoffJSONAndUsage(t *testing.T) {
	h := newHarness(t)
	id := handoffReady(h, "code")
	handoffRemote(t, h)
	fakeCloud(t)
	code, out := handoffTerminal(t, h, "", "handoff", "start", id, "--yes", "--json")
	// stderr has the session count; stdout alone is one JSON value.
	payload, _, _ := strings.Cut(out, "About to start")
	var result struct {
		OK bool `json:"ok"`
	}
	if code != 0 || json.Unmarshal([]byte(payload), &result) != nil || !result.OK {
		t.Fatalf("JSON: %d %s", code, out)
	}
	h.fails("usage:", "handoff")
	h.fails("usage:", "handoff", "start")
	h.fails("usage:", "handoff", "fetch", id)
	h.fails("unknown handoff", "handoff", "bogus")
}
