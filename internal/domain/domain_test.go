package domain

import (
	"strings"
	"testing"
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
		{CommitMode: "sometimes"},
		{CommitMode: CommitOff, ViewOrder: []string{""}, Views: map[string]string{"": ""}},
		{CommitMode: CommitOff, ViewOrder: []string{"A", "A"}, Views: map[string]string{"A": ""}},
		{CommitMode: CommitOff, ViewOrder: []string{"A"}, Views: map[string]string{"A": "--state nope"}},
		{CommitMode: CommitOff, ViewOrder: []string{"A"}, Views: map[string]string{"A": "", "B": ""}},
	}
	for _, cfg := range bad {
		if _, err := tree.PlanConfig(cfg); err == nil {
			t.Errorf("PlanConfig(%+v) accepted", cfg)
		}
	}
	c, err := tree.PlanConfig(Config{CommitMode: CommitAll, ViewOrder: []string{"Hot"}, Views: map[string]string{"Hot": " --priority  high "}})
	if err != nil {
		t.Fatal(err)
	}
	if got := tree.Apply(c).Project.Config; got.CommitMode != CommitAll || got.Views["Hot"] != "--priority high" {
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
