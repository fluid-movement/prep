package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func TestProjectDoDIsEditableGuidance(t *testing.T) {
	h := newHarness(t)
	projectPath := filepath.Join(h.dir, ".prep", "project.md")
	original, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectPath, append(original, []byte("\n## Notes\n\nKeep this prose.\n")...), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := mdstore.ReadProject(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	h.ok("dod", "--add", "Review the output", "--add", "Inspect empty input")
	h.ok("dod", "--remove", "Review the output")
	after, err := mdstore.ReadProject(h.dir)
	if err != nil {
		t.Fatal(err)
	}
	want := append(append([]string(nil), before.DoD...), "Inspect empty input")
	if !reflect.DeepEqual(after.DoD, want) {
		t.Fatalf("DoD = %v, want %v", after.DoD, want)
	}
	raw, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "## Notes\n\nKeep this prose.") || after.ID != before.ID {
		t.Fatalf("project facts or prose lost: %s", raw)
	}
	h.fails("without an issue ID", "dod", "some-issue", "--add", "Old issue form")
	h.fails("no project Definition of Done item", "dod", "--remove", "Missing item")

	if out := h.ok("prime"); !strings.Contains(out, "Definition of Done (guidance):") || !strings.Contains(out, "Inspect empty input") {
		t.Fatalf("briefing lacks guidance: %s", out)
	}
	var brief primeBrief
	h.jsonOf(&brief, "prime")
	if !reflect.DeepEqual(brief.DoD, want) {
		t.Fatalf("briefing DoD = %v", brief.DoD)
	}

	id := h.newIssue("--title", "Finish with guidance", "--kind", "code", "--body", "Exercise the completion path.")
	h.ok("criterion", id, "--add", "Behavior is verified")
	h.ok("context", id, "--body", "Inspect the CLI behavior.")
	h.ok("define", id)
	h.ok("ready", id)
	h.ok("claim", id)
	var guide domain.Guide
	h.jsonOf(&guide, "guide", id)
	if !reflect.DeepEqual(guide.DoD, want) {
		t.Fatalf("brief lacks guidance: %+v", guide)
	}
	h.ok("criterion", id, "--check", "1")
	h.jsonOf(&guide, "guide", id)
	if guide.Step != "check" || !reflect.DeepEqual(guide.DoD, want) {
		t.Fatalf("check guide lacks guidance: %+v", guide)
	}
	if out := h.ok("guide", id); !strings.Contains(out, "Inspect empty input") {
		t.Fatalf("text guide lacks guidance: %s", out)
	}
	if out := h.ok("show", id); strings.Contains(out, "Definition of Done") {
		t.Fatalf("show contains project guidance: %s", out)
	}
	// No additional per-issue records are needed to finish, and the project
	// guidance has no checked state or completion gate.
	h.ok("complete", id)
	h.expectState(id, "done", false)
	for _, file := range []string{"ready.md", "resolution.md"} {
		raw := h.read(id, file)
		for _, field := range []string{"contract:", "verification:", "approvals:", "dod:", "knowledge_reviews:"} {
			if strings.Contains(raw, field) {
				t.Fatalf("%s contains obsolete field %s: %s", file, field, raw)
			}
		}
	}
	h.ok("check")
}

func TestHelpOmitsRemovedCommands(t *testing.T) {
	h := newHarness(t)
	help := h.ok("help")
	for _, text := range []string{"policy ", "evidence ", "exception ", "dod approve", "knowledge review", "--opt-out"} {
		if strings.Contains(help, text) {
			t.Fatalf("help contains %q: %s", text, help)
		}
	}
}

func TestGuideIsImplementerBrief(t *testing.T) {
	h := newHarness(t)
	h.ok("dod", "--add", "Run the full suite")
	id := h.newIssue("--title", "Brief", "--kind", "code", "--body", "Unique requirement prose.")
	h.ok("criterion", id, "--add", "Unique criterion")
	h.ok("context", id, "--body", "Unique context text.")
	h.ok("define", id)
	h.ok("decide", id, "--title", "Unique decision", "--description", "d", "--body", "b")
	h.ok("ready", id)
	for _, step := range []string{"ready", "claimed"} {
		if step == "claimed" {
			h.ok("claim", id)
		}
		out := h.ok("guide", id)
		for _, want := range []string{"Unique requirement prose.", "1. [ ] Unique criterion", "Unique context text.", "Unique decision", "Run the full suite", "ready for check", "log the conflict", "Checker", "prep complete " + id + " [--commit <hash>]"} {
			if !strings.Contains(out, want) {
				t.Fatalf("%s guide lacks %q:\n%s", step, want, out)
			}
		}
		implementer := out[:strings.Index(out, "Transitions:")]
		if strings.Contains(strings.ReplaceAll(implementer, "Never run prep complete", ""), "prep complete") {
			t.Fatalf("%s implementer is told to complete:\n%s", step, implementer)
		}
		var g domain.Guide
		h.jsonOf(&g, "guide", id)
		if g.Requirement == "" || len(g.Criteria) != 1 || g.Context == "" || len(g.Decisions) != 1 || len(g.DoD) == 0 || len(g.Checker) == 0 {
			t.Fatalf("%s json guide = %+v", step, g)
		}
	}
}
