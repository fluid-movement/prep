package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnowledgeRecords(t *testing.T) {
	h := newHarness(t)
	issue := h.newIssue("--title", "Question", "--kind", "decision", "--body", "Choose.")
	h.ok("define", issue)
	create := func(args ...string) string {
		t.Helper()
		var out struct {
			Entry string `json:"entry"`
		}
		h.jsonOf(&out, args...)
		if out.Entry == "" {
			t.Fatal("missing entry path")
		}
		return out.Entry
	}
	first := create("decide", issue, "--title", "First choice", "--description", "Distinctive decision", "--body", "A reason.", "--scope", "internal/**", "--scope", "cmd/**")
	recordID := strings.TrimSuffix(filepath.Base(first), ".md")
	before, err := os.ReadFile(filepath.Join(h.dir, ".prep/knowledge", first))
	if err != nil {
		t.Fatal(err)
	}
	second := create("decide", "--title", "Replacement", "--description", "A better choice", "--supersedes", recordID[len(recordID)-8:])
	after, _ := os.ReadFile(filepath.Join(h.dir, ".prep/knowledge", first))
	if string(before) != string(after) {
		t.Fatal("superseded record was edited")
	}
	for _, args := range [][]string{{"knowledge", "show", first}, {"knowledge", "find", "Distinctive"}, {"knowledge", "list"}, {"show", issue}} {
		if out := h.ok(args...); !strings.Contains(out, "superseded") {
			t.Fatalf("%v misses superseded status: %s", args, out)
		}
	}
	var listed struct {
		Entries []struct {
			ID, Issue  string
			Superseded bool
		}
	}
	h.jsonOf(&listed, "knowledge", "list", "--type", "decision")
	if len(listed.Entries) != 2 || !listed.Entries[0].Superseded || listed.Entries[0].Issue != issue {
		t.Fatalf("records: %+v", listed)
	}
	var shown struct {
		Entries []struct {
			Scope []string
			Date  string
		}
	}
	h.jsonOf(&shown, "knowledge", "show", first)
	if len(shown.Entries) != 1 || len(shown.Entries[0].Scope) != 2 || shown.Entries[0].Date != "2026-10-05" {
		t.Fatalf("metadata: %+v", shown)
	}
	if out := h.ok("guide", issue); !strings.Contains(out, "First choice") {
		t.Fatalf("guide omits record title: %s", out)
	}
	h.fails("G_OUTCOME", "complete", issue, "--json")
	// A standalone outcome cannot satisfy an issue's completion gate.
	create("decide", "--title", "Standalone outcome", "--description", "Unlinked", "--outcome")
	h.fails("G_OUTCOME", "complete", issue, "--json")
	create("decide", issue, "--title", "Final choice", "--description", "Agreed", "--supersedes", strings.TrimSuffix(filepath.Base(second), ".md"), "--outcome")
	h.ok("complete", issue)
	research := h.newIssue("--title", "Research", "--kind", "research", "--body", "Investigate.")
	h.ok("define", research)
	create("finding", "--title", "Standalone finding", "--description", "Unlinked")
	h.fails("G_FINDINGS", "complete", research, "--json")
	create("finding", research, "--title", "Answer", "--description", "Linked finding", "--body", "The result.")
	h.ok("complete", research)
	if out := h.ok("knowledge", "find", "Unlinked"); !strings.Contains(out, "Standalone finding") {
		t.Fatalf("title/description-only record not searchable: %s", out)
	}
	if out := h.ok("knowledge", "list", "--scope", "other/path.go"); !strings.Contains(out, "Standalone finding") {
		t.Fatalf("project-wide record missing from scoped query: %s", out)
	}
	for _, id := range []string{issue, research} {
		for _, name := range []string{"decisions.md", "findings.md"} {
			if _, err := os.Stat(filepath.Join(h.dir, ".prep/issues", id, name)); !os.IsNotExist(err) {
				t.Fatalf("unexpected legacy record %s", name)
			}
		}
	}
	h.ok("check")
	h.ok("fmt", "--check")
	// The generated index can be discarded and rebuilt after a union merge.
	index := filepath.Join(h.dir, ".prep/knowledge/decisions/index.md")
	b, _ := os.ReadFile(index)
	if err := os.WriteFile(index, append(b, b...), 0644); err != nil {
		t.Fatal(err)
	}
	h.fails("not in canonical", "fmt", "--check")
	h.ok("fmt")
	h.ok("fmt", "--check")
}

func TestRemovedKnowledgeCommandsAndCompletionFlags(t *testing.T) {
	h := newHarness(t)
	for _, sub := range []string{"new", "update", "confirm", "bootstrap"} {
		h.fails("subcommand", "knowledge", sub)
	}
	h.fails("unknown command", "findings")
	id := h.newIssue("--title", "Task", "--kind", "code")
	for _, flag := range []string{"--docs", "--no-impact"} {
		h.fails("flag provided but not defined", "complete", id, flag, "x")
	}
	h.fails("flag provided but not defined", "init", "--no-bootstrap")
	h.fails("flag provided but not defined", "check", "--no-drift")
	h.fails("one-line", "decide", "--title", "Missing description")
	h.fails("one-line", "finding", "--title", "Multiline", "--description", "one\ntwo")
	h.fails("no knowledge record", "decide", "--title", "Bad supersedes", "--description", "Bad reference", "--supersedes", "D1")
	var out map[string]any
	if err := json.Unmarshal([]byte(h.ok("prime", "--json")), &out); err != nil {
		t.Fatal(err)
	}
	if _, ok := out["bootstrap"]; ok {
		t.Fatal("bootstrap alert remains")
	}
	// init has no prerequisite knowledge or bootstrap issues.
	fresh := &harness{t: t, dir: t.TempDir(), now: h.now}
	fresh.ok("init")
	fresh.jsonOf(&out, "list")
	if issues, ok := out["issues"].([]any); ok && len(issues) > 0 {
		t.Fatal("init created issues")
	}
}
