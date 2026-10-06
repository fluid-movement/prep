package mdstore

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestCompareAndSwapRejectsConcurrentEdit(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	tree := domain.NewTree(domain.Project{}, nil, nil, nil)
	c, err := tree.PlanNew(domain.NewIssueInput{Title: "A", Kind: domain.KindCode, Body: "A."}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}

	// Two writers load the same state.
	w1, w2 := Open(root), Open(root)
	for _, w := range []*Store{w1, w2} {
		if _, _, _, err := w.Load(); err != nil {
			t.Fatal(err)
		}
	}
	rel := IssueDir(c.IssueID) + "/history.md"
	if err := w1.write(rel, "first\n"); err != nil {
		t.Fatal(err)
	}
	err = w2.write(rel, "second\n")
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(root, rel))
	if string(b) != "first\n" {
		t.Fatalf("history.md = %q", b)
	}
	// No temp files are left behind.
	entries, _ := os.ReadDir(filepath.Join(root, IssueDir(c.IssueID)))
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".md" {
			t.Fatalf("leftover file %s", e.Name())
		}
	}
}

func TestRoundTripIsCanonical(t *testing.T) {
	raw := "---\ntitle: X\nkind: code\nparent: 01M45YYRG0AP2P2A41FRCHGR0Q\ndepends_on:\n  - 01M45YYSF8MH8F910XSA7Q366A\n---\n\nProse line.\n\n## Open questions\n\n- why?\n"
	i := &domain.Issue{}
	if err := parseIssue(raw, i); err != nil {
		t.Fatal(err)
	}
	if got := renderIssue(i); got != raw {
		t.Fatalf("round trip differs:\n%q\n%q", got, raw)
	}
	if i.Prose != "Prose line." || i.OpenQuestions != "- why?" {
		t.Fatalf("prose %q questions %q", i.Prose, i.OpenQuestions)
	}
}

func TestDecisionsTolerateBlankLines(t *testing.T) {
	raw := "## D1: Use CSV\n\ndate: 2026-10-05\n\noutcome: true\n\nBecause.\n\n## D2: Stream\ndate: 2026-10-06\n\nRows.\n"
	ds, problems := parseDecisions(raw)
	if len(problems) > 0 || len(ds) != 2 || ds[0].Date != "2026-10-05" || !ds[0].Outcome || ds[0].Body != "Because." || ds[1].Body != "Rows." {
		t.Fatalf("parsed %+v problems %v", ds, problems)
	}
	want := "## D1: Use CSV\ndate: 2026-10-05\noutcome: true\n\nBecause.\n\n## D2: Stream\ndate: 2026-10-06\n\nRows.\n"
	if got := canonicalDecisions(raw); got != want {
		t.Fatalf("canonical =\n%q\nwant\n%q", got, want)
	}
	if canonicalDecisions(want) != want {
		t.Fatal("canonical form is not stable")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	tree := domain.NewTree(domain.Project{Config: cfg}, nil, nil, nil)
	apply := func(cfg domain.Config) {
		t.Helper()
		c, err := tree.PlanConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Open(root).Apply(c); err != nil {
			t.Fatal(err)
		}
	}
	read := func() string {
		b, err := os.ReadFile(filepath.Join(root, Dir, "config.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	// Writing the configuration unchanged reproduces the file prep init wrote.
	apply(cfg)
	if got := read(); got != DefaultConfig {
		t.Fatalf("unchanged config rewritten as\n%s\nwant\n%s", got, DefaultConfig)
	}

	// Order, renames, new views and the commit mode survive a load.
	cfg.CommitMode = domain.CommitAll
	cfg.ViewOrder = []string{"Mine: urgent", "All", "Actionable"}
	cfg.Views = map[string]string{"Mine: urgent": "--priority critical,high  --tag ui", "All": "", "Actionable": "--actionable"}
	apply(cfg)
	got, err := Open(root).LoadConfig()
	if err != nil {
		t.Fatalf("%v\n%s", err, read())
	}
	if got.CommitMode != domain.CommitAll || strings.Join(got.ViewOrder, "|") != "Mine: urgent|All|Actionable" || got.Views["Mine: urgent"] != "--priority critical,high --tag ui" {
		t.Fatalf("loaded %+v from\n%s", got, read())
	}
	if !strings.HasPrefix(read(), "# prep project configuration") {
		t.Fatalf("header lost:\n%s", read())
	}
}
