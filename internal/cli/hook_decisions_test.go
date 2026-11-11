package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func hookDecide(t *testing.T, h *harness, title string, extra ...string) string {
	t.Helper()
	var out struct{ Entry string }
	h.jsonOf(&out, append([]string{"decide", "--title", title, "--description", "Test record"}, extra...)...)
	return strings.TrimSuffix(filepath.Base(out.Entry), ".md")
}

func hookEdit(dir, session, tool, path string) hookEvent {
	return hookEvent{Cwd: dir, Session: session, Tool: tool, Input: map[string]any{"file_path": path}}
}

func TestHookBeforeToolNamesScopedDecisions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	old := hookDecide(t, h, "Old choice", "--scope", "src/*.go")
	cart := hookDecide(t, h, "Cart choice", "--scope", "src/cart.go", "--supersedes", old)
	hookDecide(t, h, "Docs choice", "--scope", "docs/**")
	h.ok("finding", "--title", "A finding", "--description", "Test record", "--scope", "src/cart.go")
	h.entry("/component.md", "component", "A component", "Descriptive", "Body", "src/cart.go")

	for _, harness := range []string{"claude-code", "codex"} {
		session := "s-" + harness
		out := runHook(t, "before-tool", harness, hookEdit(h.dir, session, "Edit", filepath.Join(h.dir, "src", "cart.go")))
		if !strings.Contains(out, `Decisions scoped to src/cart.go:\n- Cart choice (`+cart+`): Test record\nRead one`) || !strings.Contains(out, `"hookEventName": "PreToolUse"`) || !strings.Contains(out, `"additionalContext"`) {
			t.Fatalf("%s: decision not named: %s", harness, out)
		}
		if !strings.Contains(out, "prep knowledge show <id>") || !strings.Contains(out, "prep decide --supersedes <id>") {
			t.Fatalf("%s: missing guidance: %s", harness, out)
		}
		// Context only: no permission decision, so the edit proceeds.
		for _, unwanted := range []string{"permissionDecision", `"decision"`, "Old choice", old, "Docs choice", "A finding", "A component"} {
			if strings.Contains(out, unwanted) {
				t.Fatalf("%s: unexpected %q: %s", harness, unwanted, out)
			}
		}
		// Once per session, also for another tool and a relative path.
		if out := runHook(t, "before-tool", harness, hookEdit(h.dir, session, "Write", "src/cart.go")); out != "" {
			t.Fatalf("%s: named again: %s", harness, out)
		}
		// A subagent has its own context.
		sub := hookEdit(h.dir, session, "Edit", "src/cart.go")
		sub.AgentID = "agent-1"
		if out := runHook(t, "before-tool", harness, sub); !strings.Contains(out, cart) {
			t.Fatalf("%s: subagent not told: %s", harness, out)
		}
		if out := runHook(t, "before-tool", harness, sub); out != "" {
			t.Fatalf("%s: subagent told twice: %s", harness, out)
		}
		// Session start (also after compaction) starts the context over.
		runHook(t, "session-start", harness, hookEvent{Cwd: h.dir, Session: session})
		if out := runHook(t, "before-tool", harness, hookEdit(h.dir, session, "Edit", "src/cart.go")); !strings.Contains(out, cart) {
			t.Fatalf("%s: not named again after session start: %s", harness, out)
		}
	}

	// apply_patch names its files in the patch.
	patch := hookEvent{Cwd: h.dir, Session: "s-patch", Tool: "apply_patch", Input: map[string]any{"command": "*** Begin Patch\n*** Update File: src/cart.go\n@@\n-a\n+b\n*** End Patch\n"}}
	if out := runHook(t, "before-tool", "codex", patch); !strings.Contains(out, `Decisions scoped to src/cart.go:\n- Cart choice (`+cart+`)`) {
		t.Fatalf("apply_patch: %s", out)
	}

	// A long description is cut at 150 characters.
	long := strings.Repeat("word ", 60)
	var created struct{ Entry string }
	h.jsonOf(&created, "decide", "--title", "Pay choice", "--description", long, "--scope", "src/pay.go")
	out := runHook(t, "before-tool", "claude-code", hookEdit(h.dir, "s-long", "Edit", "src/pay.go"))
	if want := "): " + strings.TrimSpace(long[:149]) + `…\nRead one`; !strings.Contains(out, want) {
		t.Fatalf("description not cut: %s", out)
	}

	// Nothing for other tools, unmatched paths and paths outside the project.
	outside := filepath.Join(t.TempDir(), "src", "cart.go")
	for name, e := range map[string]hookEvent{
		"read":      hookEdit(h.dir, "s-none", "Read", "src/cart.go"),
		"bash":      {Cwd: h.dir, Session: "s-none", Tool: "Bash", Input: map[string]any{"command": "gofmt -w src/cart.go"}},
		"unmatched": hookEdit(h.dir, "s-none", "Edit", "src/tax.go"),
		"outside":   hookEdit(h.dir, "s-none", "Edit", outside),
		"dotdot":    hookEdit(h.dir, "s-none", "Edit", "../src/cart.go"),
	} {
		if out := runHook(t, "before-tool", "claude-code", e); out != "" {
			t.Fatalf("%s: unexpected output: %s", name, out)
		}
	}
	// An edit under .prep is denied as before and names no decisions.
	out = runHook(t, "before-tool", "claude-code", hookEdit(h.dir, "s-none", "Write", filepath.Join(mdstore.Dir, "knowledge", "decisions", cart+".md")))
	if !strings.Contains(out, `"permissionDecision": "deny"`) || strings.Contains(out, "Decisions scoped") {
		t.Fatalf(".prep edit: %s", out)
	}

	// The outcome is a hook event carrying the path, the size and the subagent.
	var context []activity.Event
	for _, e := range hookOutcomes(t, h.dir) {
		if e.Verb == "context" {
			context = append(context, e)
		}
	}
	if len(context) != 8 {
		t.Fatalf("context events: %+v", context)
	}
	if e := context[0]; e.Op != "before-tool" || e.Area != activity.AreaCode || e.Target != "src/cart.go" || e.Chars == 0 || e.Session != "s-claude-code" || e.Actor != "claude-code" || e.AgentID != "" {
		t.Fatalf("context event: %+v", e)
	}
	if context[1].AgentID != "agent-1" {
		t.Fatalf("subagent event: %+v", context[1])
	}
}

func TestHookBeforeToolNamesFiveMostSpecificThenNewest(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	wide := hookDecide(t, h, "Wide choice", "--scope", "**")
	var same []string
	for i := 0; i < 6; i++ {
		same = append(same, hookDecide(t, h, fmt.Sprintf("Go choice %d", i), "--scope", "src/*.go"))
	}
	exact := hookDecide(t, h, "Exact choice", "--scope", "src/cart.go")
	sort.Sort(sort.Reverse(sort.StringSlice(same)))

	e := hookEdit(h.dir, "s-many", "Edit", "src/cart.go")
	out := runHook(t, "before-tool", "claude-code", e)
	// The exact scope first, then the newest of the equally specific ones.
	last := strings.Index(out, exact)
	if last < 0 {
		t.Fatalf("most specific decision missing: %s", out)
	}
	for _, id := range same[:4] {
		at := strings.Index(out, id)
		if at < last {
			t.Fatalf("order of %s: %s", id, out)
		}
		last = at
	}
	for _, id := range append(same[4:], wide) {
		if strings.Contains(out, id) {
			t.Fatalf("more than five named (%s): %s", id, out)
		}
	}
	if !strings.Contains(out, `\n- 2 more match: prep knowledge list --type decision --scope src/cart.go\n`) {
		t.Fatalf("remaining count missing: %s", out)
	}
	// The rest follow at the next edit, then nothing. A project-wide decision
	// (scope **) is never named here.
	out = runHook(t, "before-tool", "claude-code", e)
	for _, id := range same[4:] {
		if !strings.Contains(out, id) {
			t.Fatalf("remaining decision %s not named: %s", id, out)
		}
	}
	if strings.Contains(out, exact) || strings.Contains(out, wide) || strings.Contains(out, "more match") {
		t.Fatalf("second call: %s", out)
	}
	if out := runHook(t, "before-tool", "claude-code", e); out != "" {
		t.Fatalf("third call: %s", out)
	}
}

func TestHookBeforeToolDecisionsFailOpen(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := newHarness(t)
	id := hookDecide(t, h, "Cart choice", "--scope", "src/cart.go")
	// The local directory is a file, so the shown list cannot be kept: the decision is still named.
	local := filepath.Join(h.dir, mdstore.Dir, "local")
	if err := os.RemoveAll(local); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if out := runHook(t, "before-tool", "claude-code", hookEdit(h.dir, "s-blocked", "Edit", "src/cart.go")); !strings.Contains(out, id) {
			t.Fatalf("decision not named: %s", out)
		}
	}
}
