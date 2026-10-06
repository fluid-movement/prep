package mdstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestDuplicateHeadings(t *testing.T) {
	text := strings.Join([]string{
		"Prose.",
		"## Open questions",
		"## Notes",
		"First.",
		"```",
		"## Notes",
		"```",
		"## open Questions ",
		"- Which one?",
		"## Notes",
		"Second.",
		"## Done",
	}, "\n")
	got := duplicateHeadings(text)
	if len(got) != 2 {
		t.Fatalf("duplicates = %+v", got)
	}
	// Case and surrounding space do not matter; a fenced heading is no heading.
	if got[0].heading != "Open questions" || got[0].count != 2 || !got[0].mergeable {
		t.Fatalf("open questions: %+v", got[0])
	}
	if got[1].heading != "Notes" || got[1].count != 2 || got[1].mergeable {
		t.Fatalf("notes: %+v", got[1])
	}
	if d := duplicateHeadings("## A\n\nText.\n\n## B\n"); d != nil {
		t.Fatalf("no duplicates expected: %+v", d)
	}
}

func TestMergeDuplicateSections(t *testing.T) {
	cases := []struct{ in, want string }{
		// The copy with content stays where it is.
		{"Prose.\n\n## Open questions\n\n## Notes\n\nN.\n\n## Open questions\n\n- Q?\n",
			"Prose.\n\n## Notes\n\nN.\n\n## Open questions\n\n- Q?\n"},
		// All empty: the first stays.
		{"A.\n\n## Open questions\n\n## Open questions\n", "A.\n\n## Open questions\n"},
		// Content in two copies is left for a person.
		{"## Notes\n\nOne.\n\n## Notes\n\nTwo.\n", "## Notes\n\nOne.\n\n## Notes\n\nTwo.\n"},
	}
	for _, c := range cases {
		got, changed := mergeDuplicateSections(c.in)
		if normalize(got) != normalize(c.want) || changed != (c.in != c.want) {
			t.Errorf("merge(%q) = %q, %v; want %q", c.in, got, changed, c.want)
		}
	}
}

func TestFixMergesDuplicateSections(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init(); err != nil {
		t.Fatal(err)
	}
	tree := domain.NewTree(domain.Project{}, nil, nil, nil)
	c, err := tree.PlanNew(domain.NewIssueInput{Title: "A", Kind: domain.KindCode, Body: "A."}, time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, IssueDir(c.IssueID))
	put := func(name, text string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("issue.md", "---\ntitle: A\nkind: code\n---\n\nA.\n\n## Open questions\n\n## Open questions\n\n- Q?\n")
	put("context.md", "## Notes\n\nOne.\n\n## Notes\n\nTwo.\n")
	put("acceptance.md", "- [ ] works\n\n## Definition of Done\n\n## Definition of Done\n\n- docs\n")

	codes := func() map[string]domain.Severity {
		_, _, diags, err := Open(root).Load()
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]domain.Severity{}
		for _, d := range diags {
			if d.Code == domain.CodeDuplicateHeading {
				out[d.File] = d.Severity
			}
		}
		return out
	}
	before := codes()
	if before["issue.md"] != domain.SevError || before["context.md"] != domain.SevWarning || before["acceptance.md"] != domain.SevWarning {
		t.Fatalf("before fix: %v", before)
	}

	fixer := Open(root)
	if _, _, _, err := fixer.Load(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixer.Fix(); err != nil {
		t.Fatal(err)
	}
	// The mergeable ones are merged; the one with content twice stays.
	if after := codes(); len(after) != 1 || after["context.md"] != domain.SevWarning {
		t.Fatalf("after fix: %v", after)
	}
	_, issues, _, err := Open(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	i := issues[0]
	if i.OpenQuestions != "- Q?" || len(i.DoDAdd) != 1 || i.DoDAdd[0] != "docs" {
		t.Fatalf("after fix: open questions %q, DoD %v", i.OpenQuestions, i.DoDAdd)
	}
}
