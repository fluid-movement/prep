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

	// Order, renames and new views survive a load.
	cfg.ViewOrder = []string{"Mine: urgent", "All", "Actionable"}
	cfg.Views = map[string]string{"Mine: urgent": "--priority critical,high  --tag ui", "All": "", "Actionable": "--actionable"}
	apply(cfg)
	got, err := Open(root).LoadConfig()
	if err != nil {
		t.Fatalf("%v\n%s", err, read())
	}
	if strings.Join(got.ViewOrder, "|") != "Mine: urgent|All|Actionable" || got.Views["Mine: urgent"] != "--priority critical,high --tag ui" {
		t.Fatalf("loaded %+v from\n%s", got, read())
	}
	if !strings.HasPrefix(read(), "# prep configuration") {
		t.Fatalf("header lost:\n%s", read())
	}
}

func TestConfigFallsBackToDist(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	written, err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(written, " "), Dir+"/config.yaml ") || !strings.Contains(strings.Join(written, " "), Dir+"/config.yaml.dist") {
		t.Fatalf("init wrote %v", written)
	}
	file := func(name string) string { return filepath.Join(root, Dir, name) }
	if b, err := os.ReadFile(file(".gitignore")); err != nil || string(b) != "config.yaml\n" {
		t.Fatalf(".gitignore = %q, %v", b, err)
	}
	if _, err := os.Stat(file("config.yaml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("init wrote config.yaml: %v", err)
	}
	views := func() string {
		t.Helper()
		cfg, err := Open(root).LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(cfg.ViewOrder, "|")
	}
	if got := views(); got != "Unresolved|All" {
		t.Fatalf("without an own config: %s", got)
	}

	// An own config replaces the dist file as a whole.
	if err := os.WriteFile(file("config.yaml"), []byte("views:\n  Mine: --tag me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := views(); got != "Mine" {
		t.Fatalf("with an own config: %s", got)
	}

	// prep check reports each invalid file under its own name.
	diags := func() []string {
		t.Helper()
		_, _, d, err := Open(root).Load()
		if err != nil {
			t.Fatal(err)
		}
		var files []string
		for _, x := range d {
			if x.Code == domain.CodeConfigInvalid {
				files = append(files, x.File)
			}
		}
		return files
	}
	if got := diags(); len(got) != 0 {
		t.Fatalf("valid configs reported: %v", got)
	}
	if err := os.WriteFile(file("config.yaml.dist"), []byte("views:\n  Bad: --state nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(diags(), " "); got != "config.yaml.dist" {
		t.Fatalf("invalid dist reported as %q", got)
	}
	if _, _, d, _ := Open(root).Load(); strings.HasPrefix(d[0].Message, "config.yaml") {
		t.Fatalf("message repeats the file name: %q", d[0].Message)
	}
	if err := os.WriteFile(file("config.yaml"), []byte("views:\n  Bad: --kind nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(diags(), " "); got != "config.yaml.dist config.yaml" {
		t.Fatalf("invalid configs reported as %q", got)
	}
}

func TestConfigKeepsThemes(t *testing.T) {
	root := t.TempDir()
	if _, err := Open(root).Init(); err != nil {
		t.Fatal(err)
	}
	own := filepath.Join(root, Dir, "config.yaml")
	src := "views:\n  All: \"\"\ntheme: mine\nthemes:\n  mine:\n    base: nord\n    accent: \"#123456\"\n"
	if err := os.WriteFile(own, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Open(root).LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme != "mine" || cfg.Themes["mine"].Base != "nord" || cfg.Themes["mine"].Colors["accent"] != "#123456" {
		t.Fatalf("loaded %+v", cfg)
	}

	// A settings save (here: a new view) keeps the theme and custom themes.
	cfg.ViewOrder = append(cfg.ViewOrder, "Hot")
	cfg.Views["Hot"] = "--priority high"
	c, err := domain.NewTree(domain.Project{Config: cfg}, nil, nil, nil).PlanConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root).Apply(c); err != nil {
		t.Fatal(err)
	}
	got, err := Open(root).LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != "mine" || got.Themes["mine"].Colors["accent"] != "#123456" || got.Views["Hot"] != "--priority high" {
		b, _ := os.ReadFile(own)
		t.Fatalf("after save %+v\n%s", got, b)
	}

	// An invalid theme is a P003 error naming the file, and is never written.
	if err := os.WriteFile(own, []byte("theme: nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, d, err := Open(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(d) == 0 || d[0].Code != domain.CodeConfigInvalid || d[0].File != "config.yaml" || !strings.Contains(d[0].Message, `unknown theme "nope"`) {
		t.Fatalf("diagnostics %+v", d)
	}
	cfg.Theme = "nope"
	if c, err := domain.NewTree(domain.Project{Config: cfg}, nil, nil, nil).PlanConfig(cfg); err == nil {
		if _, err := Open(root).Apply(c); err == nil {
			t.Fatal("an unknown theme was written")
		}
	}
}
