package domain

import (
	"strings"
	"testing"
)

func guideInputPaths(g Guide) string {
	var parts []string
	for _, p := range g.Inputs {
		parts = append(parts, p.Path)
	}
	return strings.Join(parts, " ")
}

func TestGuideCarriesRecordsAtEveryStep(t *testing.T) {
	crit := []Criterion{{Text: "It works"}}
	base := func(kind Kind, req string) []Baseline {
		return []Baseline{{Name: "20261017-120000", Kind: kind, Requirement: req}}
	}
	files := []string{"acceptance.md", "context.md"}
	cases := map[string]*Issue{
		"open":     {ID: idA, Kind: KindCode, Body: "Do it.\n\n## Open questions\n\nWhich?", Prose: "Do it."},
		"research": {ID: idA, Kind: KindResearch, Body: "Do it.", Prose: "Do it.", Criteria: crit, Context: "Scope.", Baselines: base(KindResearch, "Do it."), Files: files},
		"decision": {ID: idA, Kind: KindDecision, Body: "Do it.", Prose: "Do it.", Criteria: crit, Context: "Options.", Baselines: base(KindDecision, "Do it."), Files: files},
		"stale":    {ID: idA, Kind: KindCode, Body: "Changed.", Prose: "Changed.", Criteria: crit, Context: "Ctx.", Baselines: base(KindCode, "Old."), Files: files},
		"done":     {ID: idA, Kind: KindCode, Body: "Do it.", Prose: "Do it.", Criteria: crit, Context: "Ctx.", Baselines: base(KindCode, "Do it."), Files: files, Resolution: &Resolution{Outcome: OutcomeDone}},
	}
	for name, i := range cases {
		t.Run(name, func(t *testing.T) {
			tree := NewTree(Project{}, []*Issue{i}, nil, nil)
			g := tree.BuildGuide(idA, func(id string) string { return id }, nil)
			if g.Requirement != i.Body {
				t.Fatalf("requirement = %q", g.Requirement)
			}
			in := guideInputPaths(g)
			if strings.Contains(in, "issue.md") {
				t.Fatalf("inputs list issue.md: %s", in)
			}
			if name == "open" {
				return
			}
			if len(g.Criteria) != 1 || g.Context == "" {
				t.Fatalf("criteria/context missing: %+v", g)
			}
			if strings.Contains(in, "acceptance.md") || strings.Contains(in, "context.md") {
				t.Fatalf("inputs list printed records: %s", in)
			}
		})
	}
}

func TestGuideCarriesRecordsForParent(t *testing.T) {
	p := &Issue{ID: idA, Kind: KindCode, Body: "Parent.", Prose: "Parent.", Criteria: []Criterion{{Text: "All done"}}, Context: "Ctx."}
	c := &Issue{ID: idB, Kind: KindCode, Body: "Child.", Prose: "Child.", Parent: idA}
	tree := NewTree(Project{}, []*Issue{p, c}, nil, nil)
	g := tree.BuildGuide(idA, func(id string) string { return id }, nil)
	if !g.Parent || g.Requirement != "Parent." || len(g.Criteria) != 1 || g.Context != "Ctx." {
		t.Fatalf("parent guide: %+v", g)
	}
}
