package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/mdstore"
)

func runHook(t *testing.T, mode, harness string, payload any) string {
	t.Helper()
	b, _ := json.Marshal(payload)
	var out, errb bytes.Buffer
	if code := Main([]string{"hook", mode, "--harness", harness}, bytes.NewReader(b), &out, &errb); code != 0 || errb.Len() != 0 {
		t.Fatalf("hook exit %d: %s", code, errb.String())
	}
	return out.String()
}

type unreadableHookInput struct{ t *testing.T }

func (r unreadableHookInput) Read([]byte) (int, error) {
	r.t.Fatal("off hook read input")
	return 0, nil
}

func TestHookContextLevelsAndFailOpen(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PREP_PROJECT", "unknown-project")
	h := newHarness(t)
	id := h.newIssue("--title", "Fix duplicate issue creation", "--kind", "code")
	sub := filepath.Join(h.dir, "src")
	os.MkdirAll(sub, 0o755)
	e := hookEvent{Cwd: sub, Prompt: "Implement " + id}
	for _, harness := range []string{"claude-code", "codex"} {
		out := runHook(t, "prompt", harness, e)
		if !strings.Contains(out, "existing work") || strings.Contains(out, "step:") {
			t.Fatalf("small: %s", out)
		}
		if runHook(t, "prompt", harness, hookEvent{Cwd: t.TempDir()}) != "" {
			t.Fatal("outside hook printed")
		}
	}
	configDir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "prep")
	os.MkdirAll(configDir, 0o755)
	writeConfig := func(s string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeConfig("hooks:\n  prompt: large\n")
	out := runHook(t, "prompt", "codex", e)
	if !strings.Contains(out, "step: define") || !strings.Contains(out, id) {
		t.Fatalf("large: %s", out)
	}
	e.Prompt = "Implement " + id[len(id)-6:]
	if strings.Contains(runHook(t, "prompt", "codex", e), "step:") {
		t.Fatal("suffix guessed")
	}
	e.Prompt = strings.Repeat(id+" ", 100)
	if out := runHook(t, "prompt", "codex", e); strings.Count(out, "step: define") != 1 || len(out) > 14000 {
		t.Fatalf("unbounded: %d", len(out))
	}
	writeConfig("hooks:\n  prompt: off\n")
	var off bytes.Buffer
	if code := Main([]string{"hook", "prompt", "--harness", "codex"}, unreadableHookInput{t}, &off, &off); code != 0 || off.Len() != 0 {
		t.Fatalf("off: %d %s", code, off.String())
	}
	for _, config := range []string{"hooks:\n  prompt: typo\n", "hooks: [invalid"} {
		writeConfig(config)
		if runHook(t, "prompt", "codex", e) != "" {
			t.Fatal("bad configuration did not fail open")
		}
	}
	for _, input := range []string{"broken", `{"tool_input":"bad"}`, strings.Repeat("x", 1<<20+1)} {
		var buf bytes.Buffer
		if code := Main([]string{"hook", "before-tool", "--harness", "codex"}, strings.NewReader(input), &buf, &buf); code != 0 || buf.Len() != 0 {
			t.Fatal("bad input did not fail open")
		}
	}
}

func TestHookSmallDoesNotLoadTracker(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, mdstore.Dir, "issues", "invalid"), 0o755)
	os.WriteFile(filepath.Join(root, mdstore.Dir, "project.md"), []byte("invalid yaml"), 0o644)
	if out := runHook(t, "prompt", "codex", hookEvent{Cwd: root}); !strings.Contains(out, "existing work") {
		t.Fatalf("small loaded tracker: %s", out)
	}
}

func TestHookRecordRedirectsAndPermissions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	knowledge := filepath.Join(h.dir, mdstore.Dir, "knowledge", "overview.md")
	issue := filepath.Join(mdstore.Dir, "issues", "some-id", "issue.md")
	cases := []struct {
		tool  string
		input map[string]any
		deny  bool
		hint  string
	}{
		{"Read", map[string]any{"file_path": knowledge}, true, "prep knowledge"},
		{"Grep", map[string]any{"path": filepath.Join(mdstore.Dir, "knowledge"), "pattern": "hook"}, true, "prep knowledge"},
		{"Glob", map[string]any{"pattern": filepath.Join(mdstore.Dir, "issues", "*", "issue.md")}, true, "prep show"},
		{"Write", map[string]any{"file_path": issue}, true, "write commands"},
		{"Bash", map[string]any{"command": "prep next && cat " + issue}, true, "prep show"},
		{"Bash", map[string]any{"command": "cat " + issue + "; rm " + issue}, true, "write commands"},
		{"Bash", map[string]any{"command": "sed -i '' s/a/b/ " + issue}, true, "write commands"},
		{"Bash", map[string]any{"command": "git diff -- " + issue}, false, ""},
		{"Bash", map[string]any{"command": "go test ./...", "timeout": 10000}, false, ""},
		{"Bash", map[string]any{"command": "prep knowledge show /overview.md"}, false, ""},
		{"Read", map[string]any{"file_path": "README.md"}, false, ""},
		{"Edit", map[string]any{"file_path": filepath.Join(mdstore.Dir, "local", "activity.jsonl")}, false, ""},
		{"apply_patch", map[string]any{"command": "*** Begin Patch\n*** Update File: internal/cli/hook.go\n@@\n+text about " + issue + "\n*** End Patch"}, false, ""},
		{"apply_patch", map[string]any{"command": "*** Begin Patch\n*** Update File: " + issue + "\n@@\n+text\n*** End Patch"}, true, "write commands"},
	}
	for _, harness := range []string{"claude-code", "codex"} {
		for _, c := range cases {
			out := runHook(t, "before-tool", harness, hookEvent{Cwd: h.dir, Session: "s1", Tool: c.tool, Input: c.input})
			if c.deny {
				if !strings.Contains(out, `"permissionDecision": "deny"`) || !strings.Contains(out, c.hint) {
					t.Fatalf("%s %s: %s", harness, c.tool, out)
				}
			} else if out != "" {
				t.Fatalf("ordinary tool permissions changed: %s", out)
			}
		}
	}
}

func TestHookSessionAndActivity(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	for _, harness := range []string{"claude-code", "codex"} {
		e := hookEvent{Cwd: h.dir, Session: "session-1"}
		out := runHook(t, "session-start", harness, e)
		if !strings.Contains(out, "SessionStart") || !strings.Contains(out, "prep guide") {
			t.Fatalf("session: %s", out)
		}
		e.Tool = "Bash"
		e.Input = map[string]any{"command": "go test ./..."}
		e.Response = json.RawMessage(`{"exit_code":1}`)
		if runHook(t, "after-tool", harness, e) != "" {
			t.Fatal("activity printed")
		}
		e.Input = map[string]any{"command": "prep next"}
		runHook(t, "after-tool", harness, e)
		e.Input = map[string]any{"command": "prep next && go test ./..."}
		runHook(t, "after-tool", harness, e)
	}
	var feed struct {
		Events []struct {
			Actor, Session, Op, Kind string
			Failed                   bool
		}
	}
	h.jsonOf(&feed, "activity", "--max", "20")
	var tools int
	for _, e := range feed.Events {
		if e.Kind == "tool" {
			tools++
		}
	}
	if tools != 4 {
		t.Fatalf("activity duplicates or missing: %+v", feed)
	}
	for _, e := range feed.Events {
		if e.Kind != "tool" {
			continue
		}
		if e.Session != "session-1" || !e.Failed || e.Kind != "tool" {
			t.Fatalf("attribution: %+v", e)
		}
	}
}
