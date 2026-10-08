package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestScopeMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"internal/export", "internal/export/csv.go", true},
		{"internal/export/", "internal/export", true},
		{"internal/export", "internal/exporter/x.go", false},
		{"internal/**", "internal/a/b/c.go", true},
		{"internal/**/*.go", "internal/a/b/c.go", true},
		{"internal/**/*.go", "internal/c.go", true},
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", false},
		{"cmd/*", "cmd/prep/main.go", true},
	}
	for _, c := range cases {
		if got := ScopeMatch(c.pattern, c.path); got != c.want {
			t.Errorf("ScopeMatch(%q, %q) = %v", c.pattern, c.path, got)
		}
	}
}

func TestParseFilter(t *testing.T) {
	f, err := ParseFilter([]string{"--state", "ready", "--state=in-progress", "--kind", "code,research", "--stale=false", "--leaf"})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.States) != 2 || f.States[1] != StateInProgress || len(f.Kinds) != 2 || f.Stale == nil || *f.Stale || f.Parent == nil || *f.Parent {
		t.Fatalf("filter = %+v", f)
	}
	for _, bad := range [][]string{{"--state", "nope"}, {"--kind"}, {"--colour", "red"}, {"ready"}} {
		if _, err := ParseFilter(bad); err == nil {
			t.Errorf("ParseFilter(%v) accepted", bad)
		}
	}
}

func TestEffectiveDoDCascades(t *testing.T) {
	p := &Issue{ID: "01M45YYRG0AP2P2A41FRCHGR0Q", Title: "P", Kind: KindCode, DoDAdd: []string{"changelog"}}
	c := &Issue{ID: "01M45YYSF8MH8F910XSA7Q366A", Title: "C", Kind: KindCode, Parent: p.ID, DoDOptOuts: []OptOut{{Item: "tests pass", Reason: "docs only"}}, DoDAdd: []string{"screenshots"}}
	tr := NewTree(Project{DoD: []string{"tests pass", "lint clean"}}, []*Issue{p, c}, nil, nil)
	items, opt := tr.EffectiveDoD(c.ID)
	want := []string{"lint clean", "changelog", "screenshots"}
	if len(items) != len(want) || len(opt) != 1 {
		t.Fatalf("dod = %v, opt-outs %v", items, opt)
	}
	for k := range want {
		if items[k] != want[k] {
			t.Fatalf("dod = %v", items)
		}
	}
}

func TestPlanConfigValidates(t *testing.T) {
	tree := NewTree(Project{}, nil, nil, nil)
	bad := []Config{
		{ViewOrder: []string{""}, Views: map[string]string{"": ""}},
		{ViewOrder: []string{"A", "A"}, Views: map[string]string{"A": ""}},
		{ViewOrder: []string{"A"}, Views: map[string]string{"A": "--state nope"}},
		{ViewOrder: []string{"A"}, Views: map[string]string{"A": "", "B": ""}},
	}
	for _, cfg := range bad {
		if _, err := tree.PlanConfig(cfg); err == nil {
			t.Errorf("PlanConfig(%+v) accepted", cfg)
		}
	}
	c, err := tree.PlanConfig(Config{ViewOrder: []string{"Hot"}, Views: map[string]string{"Hot": " --priority  high "}})
	if err != nil {
		t.Fatal(err)
	}
	if got := tree.Apply(c).Project.Config; got.Views["Hot"] != "--priority high" {
		t.Fatalf("applied config = %+v", got)
	}
}

func TestKnowledgeViewQueries(t *testing.T) {
	entries := []*Entry{
		{Path: "/components/cli.md", Type: "component", Title: "CLI", Status: "stable", Scope: []string{"internal/cli"}},
		{Path: "/components/tui.md", Type: "component", Title: "TUI", Status: "draft", Scope: []string{"internal/tui/*.go"}},
		{Path: "/decisions/stack.md", Type: "decision", Title: "Stack choice", Status: "stable"},
	}
	done := &Issue{ID: "01KDYYYSM08C9CDRA20DXH61HN", Title: "A", Kind: KindCode, ContextLinks: []string{"/components/cli.md", "/missing.md"},
		Resolution: &Resolution{Outcome: "done", Documentation: &Documentation{Entries: []string{"/components/tui.md", "/components/cli.md"}}}}
	later := &Issue{ID: "01KDZ2CN80NAXMQZ9SREPYAMQQ", Title: "B", Kind: KindCode,
		Resolution: &Resolution{Outcome: "done", Documentation: &Documentation{Entries: []string{"/components/tui.md"}}}}
	tree := NewTree(Project{}, []*Issue{done, later}, entries, nil)

	if got := tree.EntryIssues("/components/tui.md"); strings.Join(got, ",") != later.ID+","+done.ID {
		t.Fatalf("EntryIssues = %v, want newest first", got)
	}
	if got := tree.IssueEntries(done.ID); strings.Join(got, ",") != "/components/cli.md,/components/tui.md" {
		t.Fatalf("IssueEntries = %v", got)
	}
	for args, want := range map[string]string{
		"--type component":          "/components/cli.md,/components/tui.md",
		"--status draft":            "/components/tui.md",
		"--scope internal":          "/components/cli.md,/components/tui.md",
		"--scope internal/cli/x.go": "/components/cli.md",
		"stack":                     "/decisions/stack.md",
		"--type decision,component --status stable": "/components/cli.md,/decisions/stack.md",
	} {
		f, err := ParseKnowledgeFilter(strings.Fields(args))
		if err != nil {
			t.Fatalf("%s: %v", args, err)
		}
		if got := strings.Join(tree.QueryKnowledge(f), ","); got != want {
			t.Errorf("%s: %s, want %s", args, got, want)
		}
	}
	if _, err := ParseKnowledgeFilter([]string{"--type", "nope"}); err == nil {
		t.Error("unknown type accepted")
	}
}

func TestKnowledgeReservedNames(t *testing.T) {
	tree := NewTree(Project{}, nil, nil, nil)
	body, typ := "Body.", "component"
	for _, p := range []string{"/index.md", "/components/index.md", "/log.md"} {
		_, err := tree.PlanKnowledge(KnowledgeEdit{Path: p, New: true, Body: &body, Type: &typ, Title: &typ, Description: &typ})
		var de *Error
		if !errors.As(err, &de) || de.Code != ErrUsage {
			t.Fatalf("%s: expected a usage error, got %v", p, err)
		}
	}
}

const (
	idA = "01M45YYRG0AP2P2A41FRCHGR0Q"
	idB = "01M45YYSF8MH8F910XSA7Q366A"
	idC = "01M45YYVDRJAN3ZJDXNK2DSAPX"
)

func TestResolveIgnoresCase(t *testing.T) {
	tree := NewTree(Project{}, []*Issue{{ID: idA, Title: "A", Kind: KindCode}, {ID: idB, Title: "B", Kind: KindCode}}, nil, nil)
	for _, ref := range []string{idA, strings.ToLower(idA), "chgr0q", "CHGR0Q"} {
		if got, err := tree.Resolve(ref); err != nil || got != idA {
			t.Errorf("Resolve(%q) = %q, %v", ref, got, err)
		}
	}
}

func TestPlanNewDependencies(t *testing.T) {
	dropped := &Issue{ID: idA, Title: "A", Kind: KindCode, Resolution: &Resolution{Outcome: OutcomeDropped, Reason: "no"}}
	open := &Issue{ID: idB, Title: "B", Kind: KindCode}
	tree := NewTree(Project{}, []*Issue{dropped, open}, nil, nil)
	c, err := tree.PlanNew(NewIssueInput{Title: "N", Kind: KindCode, DependsOn: []string{idB, idB}}, time.Now())
	if err != nil || len(c.NewIssue.DependsOn) != 1 {
		t.Fatalf("plan = %+v, %v", c, err)
	}
	_, err = tree.PlanNew(NewIssueInput{Title: "N", Kind: KindCode, DependsOn: []string{idA}}, time.Now())
	var de *Error
	if !errors.As(err, &de) || len(de.Unmet) != 1 || de.Unmet[0].Code != GateDeps {
		t.Fatalf("depending on a dropped issue: %v", err)
	}
}

func TestAckAfterKindChangeNeedsContext(t *testing.T) {
	i := &Issue{ID: idA, Title: "A", Kind: KindCode, Prose: "Do it.", Criteria: []Criterion{{Text: "x"}},
		Baselines: []Baseline{{Name: "20261005-120000", Kind: KindManual, Requirement: "Do it."}}, Ready: &Ready{Baseline: "20261005-120000"}}
	tree := NewTree(Project{}, []*Issue{i}, nil, nil)
	if !tree.Stale(idA) {
		t.Fatal("a kind change should make the issue stale")
	}
	unmet, _ := tree.Gates(idA, OpAck, nil)
	if len(unmet) != 1 || unmet[0].Code != GateContext {
		t.Fatalf("unmet = %+v", unmet)
	}
	i.Context = "Where."
	if unmet, _ := NewTree(Project{}, []*Issue{i}, nil, nil).Gates(idA, OpAck, nil); len(unmet) != 0 {
		t.Fatalf("with context: unmet = %+v", unmet)
	}
}

func TestDroppedParentWithOpenChild(t *testing.T) {
	p := &Issue{ID: idA, Title: "P", Kind: KindManual, Resolution: &Resolution{Outcome: OutcomeDropped, Reason: "no"}}
	c := &Issue{ID: idB, Title: "C", Kind: KindCode, Parent: idA}
	found := false
	for _, d := range Validate(NewTree(Project{}, []*Issue{p, c}, nil, nil)) {
		found = found || d.Code == CodeResolvedChildOpen && d.Issue == idA
	}
	if !found {
		t.Fatal("no warning for a dropped parent with an open child")
	}
}

// codes lists the gate codes of unmet conditions.
func codes(unmet []Unmet) []string {
	var out []string
	for _, u := range unmet {
		out = append(out, u.Code)
	}
	return out
}

func TestGatesForDefineAndComplete(t *testing.T) {
	open := &Issue{ID: idA, Title: "A", Kind: KindCode}
	tree := NewTree(Project{}, []*Issue{open}, nil, nil)
	if got := codes(must(tree.Gates(idA, OpDefine, nil))); !slices.Equal(got, []string{GateRequirement}) {
		t.Fatalf("define without prose: %v", got)
	}

	busy := &Issue{ID: idB, Title: "B", Kind: KindCode, Prose: "Do it.", Body: "Do it.", Context: "Where.", Criteria: []Criterion{{Text: "x", Checked: true}},
		Baselines: []Baseline{{Name: "20261005-120000", Kind: KindCode, Requirement: "Do it."}}, Ready: &Ready{Baseline: "20261005-120000"}, Claim: &Claim{By: "agent/1"}}
	entries := []*Entry{{Path: "/overview.md", Type: "overview", Title: "O", Description: "O."}}
	tree = NewTree(Project{}, []*Issue{busy}, entries, nil)
	cases := []struct {
		in   Input
		want []string
	}{
		{Input{Actor: "agent/1"}, []string{GateDocs}},
		{Input{Actor: "agent/1", Docs: []string{"/overview.md"}, NoImpact: "none"}, []string{GateDocs}},
		{Input{Actor: "agent/1", Docs: []string{"/gone.md"}}, []string{GateDocEntry}},
		{Input{Actor: "agent/1", Docs: []string{"overview"}}, nil},
		{Input{Actor: "agent/1", NoImpact: "internal only"}, nil},
	}
	for _, c := range cases {
		unmet, _ := tree.Gates(idB, OpComplete, &c.in)
		if got := codes(unmet); !slices.Equal(got, c.want) {
			t.Errorf("complete %+v: unmet %v, want %v", c.in, got, c.want)
		}
	}
}

func must(unmet []Unmet, _ []string) []Unmet { return unmet }
