// Command build-fixture makes one isolated, deterministic knowledge-policy
// trial workspace from the existing webapp eval fixture.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fluid-movement/prep/internal/eval/fixture"
)

func main() {
	condition := flag.String("condition", "", "code-only, current, lean, or stale")
	out := flag.String("out", "", "new output directory")
	binary := flag.String("binary", "", "prep binary built from the pinned source")
	specPath := flag.String("spec", "eval/fixtures/webapp.yaml", "existing eval fixture spec")
	flag.Parse()
	if *out == "" || *binary == "" {
		fatal("--out and --binary are required")
	}
	f, err := os.Open(*specPath)
	if err != nil {
		fatal("open fixture: %v", err)
	}
	spec, err := fixture.LoadSpec(f)
	_ = f.Close()
	if err != nil {
		fatal("load fixture: %v", err)
	}
	profile := fixture.KnowledgeSpec{
		Path:        "/components/profile-fields.md",
		Type:        "component",
		Title:       "Profile fields",
		Description: "Profile field storage and display behavior.",
		Scope:       []string{"db/schema.sql", "src/lib/profile.ts"},
	}
	overview := fixture.KnowledgeSpec{
		Path:        "/overview.md",
		Type:        "overview",
		Title:       "Project knowledge",
		Description: "Entry point and basic navigation for this fixture project.",
		Body:        "This project tracks player behavior. Check db/schema.sql and src/lib/profile.ts; use prep knowledge find/show for project guidance and verify it against current code.",
	}
	switch *condition {
	case "code-only":
		spec.Knowledge = []fixture.KnowledgeSpec{overview}
	case "current":
		// Keep the existing full webapp entries and add the shared scaffold.
		spec.Knowledge = append(spec.Knowledge, overview)
	case "lean":
		profile.Body = "Storage and display limits must stay aligned; the tracked criterion defines the target. Inspect db/schema.sql and src/lib/profile.ts. Code is the authority if this note disagrees."
		spec.Knowledge = []fixture.KnowledgeSpec{overview, profile}
	case "stale":
		profile.Body = "The headline is stored as VARCHAR(200); profile rendering does not truncate it."
		spec.Knowledge = []fixture.KnowledgeSpec{overview, profile}
	default:
		fatal("unknown condition %q", *condition)
	}
	project, err := fixture.Build(spec, *out, fixture.Options{
		Binary: *binary,
		Start:  time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC),
		Seed:   711,
		Actor:  "fixture",
	})
	if err != nil {
		fatal("build workspace: %v", err)
	}
	fmt.Printf("project=%s\nheadline_issue=%s\n", project.Dir, project.IDs["headline"])
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
