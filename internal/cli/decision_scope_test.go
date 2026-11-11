package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestDecideRequiresDelivery(t *testing.T) {
	h := newHarness(t)
	args := []string{"decide", "--title", "Choice", "--description", "Reason"}
	code, out := h.run(append(args, "--json")...)
	for _, want := range []string{"E_USAGE", "prep decide <id>", "--scope", "--project-wide"} {
		if code == 0 || !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
	h.fails("not both", append(args, "--project-wide", "--scope", "src/**")...)
	h.fails("use --project-wide", append(args, "--scope", "**")...)
	h.fails("E_USAGE", append(args, "--scope", " ")...)
	h.ok(append(args, "--scope", "src/**")...)
	old := recordIDForTitle(h, "Choice")
	// Superseding does not inherit the old scope.
	h.fails("--project-wide", append(args, "--supersedes", old)...)
	h.ok(append(args, "--project-wide", "--supersedes", old)...)
	var shown struct{ Entries []domain.Entry }
	h.jsonOf(&shown, "knowledge", "list", "--type", "decision")
	if len(shown.Entries) != 2 || len(shown.Entries[1].Scope) != 1 || shown.Entries[1].Scope[0] != "**" {
		t.Fatalf("project-wide scope: %+v", shown)
	}
	issue := h.newIssue("--title", "Planning", "--kind", "decision")
	h.ok(append(args, issue)...)
	h.ok("finding", "--title", "Measured", "--description", "No link or scope required")
	h.ok("finding", "--title", "All paths", "--description", "Existing scope syntax", "--scope", "**")
	h.fails("flag provided but not defined", "finding", "--project-wide")
}

func TestDecideProposesSessionFocusOrSingleClaim(t *testing.T) {
	h := newHarness(t)
	first := h.newIssue("--title", "First", "--kind", "code", "--body", "Implement.")
	second := h.newIssue("--title", "Second", "--kind", "code", "--body", "Implement.")
	for _, id := range []string{first, second} {
		h.ok("define", id)
		h.ok("context", id, "--body", "Change the implementation.")
		h.ok("criterion", id, "--add", "It works")
		h.ok("ready", id)
	}
	t.Setenv("PREP_SESSION", "other")
	h.ok("claim", first)
	t.Setenv("PREP_SESSION", "current")
	assertProposal := func(want, unwanted string) {
		t.Helper()
		code, out := h.run("decide", "--title", "Choice", "--description", "Reason")
		if code == 0 || !strings.Contains(out, "prep decide "+want) || (unwanted != "" && strings.Contains(out, unwanted)) {
			t.Fatalf("proposal: %s", out)
		}
	}
	assertProposal(first, "") // Only one in-progress issue, in another session.
	t.Setenv("PREP_SESSION", "other")
	h.ok("claim", second)
	t.Setenv("PREP_SESSION", "current")
	assertProposal("<id>", first) // More than one, with no focus in this session.
	h.ok("focus", first)
	assertProposal(first, second)
	h.ok("decide", first, "--title", "Choice", "--description", "Retry succeeds")
}

func TestGuideAncestorDecisionsAtEveryStep(t *testing.T) {
	h := newHarness(t)
	root := h.newIssue("--title", "Root", "--kind", "code")
	parent := h.newIssue("--title", "Parent", "--kind", "code", "--parent", root)
	child := h.newIssue("--title", "Child", "--kind", "code", "--parent", parent, "--body", "Implement the child.")
	unrelated := h.newIssue("--title", "Other", "--kind", "code")
	for n, id := range []string{child, parent, root, unrelated} {
		h.ok("decide", id, "--title", fmt.Sprintf("Rule %d", n), "--description", "Reason")
	}
	h.ok("finding", parent, "--title", "Measurement", "--description", "Not a decision")
	assertGuide := func() {
		t.Helper()
		var g domain.Guide
		h.jsonOf(&g, "guide", child)
		if len(g.Decisions) != 3 || g.Decisions[0].From != "" || g.Decisions[1].From != parent || g.Decisions[2].From != root {
			t.Fatalf("decisions: %+v", g.Decisions)
		}
		out := h.ok("guide", child)
		for _, want := range []string{"Rule 0", "Rule 1 [ancestor: " + parent + "]", "Rule 2 [ancestor: " + root + "]"} {
			if !strings.Contains(out, want) {
				t.Fatalf("missing %q: %s", want, out)
			}
		}
	}
	assertGuide() // Open issues carry their content too.
	h.ok("define", child)
	assertGuide()
	h.ok("context", child, "--body", "Implement.")
	h.ok("criterion", child, "--add", "Works")
	h.ok("ready", child)
	assertGuide()
	h.ok("claim", child)
	assertGuide()
}

func TestPrimeProjectDecisions(t *testing.T) {
	h := newHarness(t)
	// Legacy unlinked and unscoped decisions still load, list and validate.
	legacy := filepath.Join(h.dir, ".prep/knowledge/decisions/01K6TGPKG00000000000000000.md")
	if err := os.MkdirAll(filepath.Dir(legacy), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("---\ntype: decision\ntitle: Legacy\ndescription: Old unscoped decision\ndate: 2025-10-05\n---\n\nStill valid.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	h.ok("fmt")
	h.ok("check")
	if out := h.ok("knowledge", "list"); !strings.Contains(out, "Legacy") {
		t.Fatal(out)
	}
	for n := 0; n < 12; n++ {
		h.ok("decide", "--project-wide", "--title", fmt.Sprintf("Global %02d", n), "--description", "Applies everywhere")
	}
	h.ok("decide", "--scope", "src/**", "--title", "Replacement", "--description", "Now narrow", "--supersedes", recordIDForTitle(h, "Global 11"))
	var b primeBrief
	h.jsonOf(&b, "prime")
	if len(b.ProjectDecisions) != 10 || b.ProjectDecisionsMore != 1 || b.ProjectDecisions[0].Title != "Global 10" || b.ProjectDecisions[9].Title != "Global 01" {
		t.Fatalf("project decisions: %+v, more %d", b.ProjectDecisions, b.ProjectDecisionsMore)
	}
	out := h.ok("prime")
	if !strings.Contains(out, b.ProjectDecisions[0].ID+" — Global 10") || !strings.Contains(out, "1 more project-wide decisions") || strings.Contains(out, "Legacy") || strings.Contains(out, "Global 11") || strings.Contains(out, "Replacement") {
		t.Fatal(out)
	}
	for _, path := range []string{"main.go", "nested/path/file.go"} {
		if !domain.ScopeMatch("**", path) {
			t.Fatalf("project-wide scope misses %s", path)
		}
	}
}
