package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestListHidesResolvedByDefault(t *testing.T) {
	h := newHarness(t)
	open := h.newIssue("--kind", "code", "--title", "Still open")
	dropped := h.newIssue("--kind", "code", "--title", "Resolved issue")
	h.ok("drop", dropped, "--reason", "Unneeded")
	out := h.ok("list")
	if !strings.Contains(out, open) || strings.Contains(out, dropped) || !strings.HasSuffix(out, "1 resolved issues hidden; --state all shows them\n") {
		t.Fatalf("default list: %s", out)
	}
	for _, args := range [][]string{{"list", "--state", "all"}, {"list", "--state=dropped"}, {"list", "--state", "not-open"}, {"list", "--json"}} {
		out := h.ok(args...)
		if !strings.Contains(out, dropped) || strings.Contains(out, "hidden") {
			t.Fatalf("%v: %s", args, out)
		}
	}
	// Saved views retain their query semantics even without a state filter.
	if err := os.WriteFile(filepath.Join(h.dir, ".prep", "config.yaml"), []byte("views:\n  everything: --kind code\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if out := h.ok("list", "--view", "everything"); !strings.Contains(out, dropped) || strings.Contains(out, "hidden") {
		t.Fatalf("view: %s", out)
	}
	if out := h.ok("list", "--tree"); strings.Contains(out, dropped) || !strings.Contains(out, "1 resolved issues hidden") {
		t.Fatalf("tree: %s", out)
	}
}

func TestKnowledgeListBoundAndNewestFirst(t *testing.T) {
	h := newHarness(t)
	for k := 0; k < 32; k++ {
		h.now = h.now.Add(24 * time.Hour)
		h.ok("decide", "--title", fmt.Sprintf("Choice %02d", k), "--description", "Choice")
	}
	out := h.ok("knowledge", "list", "--type", "decision")
	if strings.Count(out, "/decisions/") != 30 || strings.Contains(out, "Choice 00") || !strings.HasSuffix(out, "2 more; --max 32 or --all\n") {
		t.Fatalf("default: %s", out)
	}
	if strings.Index(out, "Choice 31") > strings.Index(out, "Choice 30") {
		t.Fatalf("newest first: %s", out)
	}
	for _, flags := range [][]string{{"--max", "32"}, {"--all"}, {"--json"}} {
		out := h.ok(append([]string{"knowledge", "list", "--type", "decision"}, flags...)...)
		if strings.Count(out, "/decisions/") != 32 || strings.Contains(out, " more;") {
			t.Fatalf("%v: %s", flags, out)
		}
	}
	// Superseded records sort behind active ones even when newer.
	h.ok("decide", "--title", "Replacement", "--description", "Choice", "--supersedes", recordIDForTitle(h, "Choice 31"))
	out = h.ok("knowledge", "list", "--type", "decision", "--all")
	if strings.Index(out, "Choice 31 [superseded]") < strings.Index(out, "Choice 00") {
		t.Fatalf("superseded ordering: %s", out)
	}
}

func recordIDForTitle(h *harness, title string) string {
	var out struct {
		Entries []struct {
			ID    string
			Title string
		}
	}
	h.jsonOf(&out, "knowledge", "list", title)
	if len(out.Entries) != 1 {
		h.t.Fatalf("entry %q: %+v", title, out)
	}
	return out.Entries[0].ID
}

func TestKnowledgeFindBoundsAndRecovery(t *testing.T) {
	h := newHarness(t)
	for k := 0; k < 7; k++ {
		h.entry(fmt.Sprintf("components/%d.md", k), "component", fmt.Sprintf("Needle %d", k), "Needle", fmt.Sprintf("needle result %d", k))
	}
	out := h.ok("knowledge", "find", "needle")
	if strings.Count(out, "== ") != 5 || !strings.HasSuffix(out, "2 more matches; --max 7\n") {
		t.Fatalf("entry bound: %s", out)
	}
	for _, flags := range [][]string{{"--max", "7"}, {"--max", "0"}, {"--json"}} {
		out := h.ok(append([]string{"knowledge", "find", "needle"}, flags...)...)
		if !strings.Contains(out, "needle result 6") || strings.Contains(out, "more matches;") {
			t.Fatalf("%v: %s", flags, out)
		}
	}
	// One oversized Unicode block must fit the total budget including the footer.
	body := "oversized " + strings.Repeat("é", 4000)
	h.entry("components/long.md", "component", "Oversized", "Oversized", body)
	out = h.ok("knowledge", "find", "oversized")
	if utf8.RuneCountInString(out) > 3000 || !utf8.ValidString(out) || !strings.Contains(out, "…\n") || !strings.HasSuffix(out, "1 more matches; --max 1\n") {
		t.Fatalf("character budget: %d characters, %s", utf8.RuneCountInString(out), out)
	}
	if out := h.ok("knowledge", "find", "oversized", "--max", "1"); !strings.Contains(out, body) || strings.Contains(out, "more matches;") {
		t.Fatal("explicit --max did not recover the whole block")
	}
}

func TestShowBoundsLargeChildList(t *testing.T) {
	h := newHarness(t)
	parent := h.newIssue("--kind", "code", "--title", "Parent")
	var resolved string
	for k := 0; k < 31; k++ {
		child := h.newIssue("--kind", "code", "--title", fmt.Sprintf("Child %d", k), "--parent", parent)
		if k == 0 {
			resolved = child
			h.ok("drop", child, "--reason", "Unneeded")
		}
		if k == 29 {
			out := h.ok("show", parent)
			if strings.Count(out, "child: ") != 30 || !strings.Contains(out, resolved) || strings.Contains(out, "children hidden") {
				t.Fatalf("30 children: %s", out)
			}
		}
	}
	out := h.ok("show", parent)
	footer := "1 resolved children hidden; prep list --under " + parent + " --state all\n"
	if strings.Count(out, "child: ") != 30 || strings.Contains(out, resolved) || !strings.Contains(out, footer) {
		t.Fatalf("31 children: %s", out)
	}
	if out := h.ok("list", "--under", parent, "--state", "all"); !strings.Contains(out, resolved) {
		t.Fatalf("recovery: %s", out)
	}
	var shown struct{ Children []string }
	h.jsonOf(&shown, "show", parent)
	if len(shown.Children) != 31 {
		t.Fatalf("JSON lost children: %+v", shown)
	}
}
