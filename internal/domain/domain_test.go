package domain

import "testing"

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
	p := &Issue{ID: "20261005-120000", Title: "P", Kind: KindCode, DoDAdd: []string{"changelog"}}
	c := &Issue{ID: "20261005-120001", Title: "C", Kind: KindCode, Parent: p.ID, DoDOptOuts: []OptOut{{Item: "tests pass", Reason: "docs only"}}, DoDAdd: []string{"screenshots"}}
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
