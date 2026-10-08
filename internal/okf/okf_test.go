package okf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
)

func TestDriftAgainstConfirmedCommit(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		out, err := gitx.Run(root, args...)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	write := func(rel, s string) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/export/csv.go", "package export\n")
	git("add", ".")
	git("commit", "-qm", "init")
	head := git("rev-parse", "HEAD")
	write(".prep/knowledge/components/export.md", "---\ntype: component\ntitle: Export\ndescription: CSV.\nscope:\n  - internal/export/**\nconfirmed_commit: "+head+"\n---\n\nBody with [link](../overview.md) and `[example](/nope.md)`.\n\n```\n[block](/nope2.md)\n```\n")

	s := &Store{Root: root}
	if _, err := s.WriteIndexes(false); err != nil {
		t.Fatal(err)
	}
	entries, diags, err := s.Load(true)
	if err != nil || len(diags) > 0 || len(entries) != 1 {
		t.Fatalf("load: %v %v %d", err, diags, len(entries))
	}
	if len(entries[0].Drifted) != 0 {
		t.Fatalf("unexpected drift %v", entries[0].Drifted)
	}
	if len(entries[0].Links) != 1 || entries[0].Links[0] != "/overview.md" {
		t.Fatalf("links = %v", entries[0].Links)
	}
	git("add", ".")
	git("commit", "-qm", "entry")

	drift := func() string {
		t.Helper()
		entries, _, err := s.Load(true)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(entries[0].Drifted, " ")
	}
	entryFile := ".prep/knowledge/components/export.md"
	entryText := func() string {
		b, err := os.ReadFile(filepath.Join(root, entryFile))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	// Code changed after the entry: uncommitted, then committed, it drifts.
	write("internal/export/csv.go", "package export\n\nfunc Stream() {}\n")
	if got := drift(); got != "internal/export/csv.go" {
		t.Fatalf("uncommitted code change: drift = %q", got)
	}
	git("commit", "-qam", "code")
	if got := drift(); got != "internal/export/csv.go" {
		t.Fatalf("committed code change: drift = %q", got)
	}

	// Updating the entry covers the scope's uncommitted changes, and once
	// both land in one commit the entry is current.
	write("internal/export/csv.go", "package export\n\nfunc Stream() {}\n\nfunc Rows() {}\n")
	write(entryFile, strings.Replace(entryText(), "Body with", "Streams rows. Body with", 1))
	if got := drift(); got != "internal/export/csv.go" {
		// The earlier committed change still counts until the entry is
		// committed; only the uncommitted one is covered now.
		t.Fatalf("entry being edited: drift = %q", got)
	}
	git("commit", "-qam", "code and entry")
	if got := drift(); got != "" {
		t.Fatalf("code and entry in one commit: drift = %q", got)
	}
	write("internal/export/csv.go", "package export\n")
	write(entryFile, entryText()+"\nMore.\n")
	if got := drift(); got != "" {
		t.Fatalf("entry edited with uncommitted code: drift = %q", got)
	}
	git("commit", "-qam", "both again")
	write("internal/export/csv.go", "package export\n\n// later\n")
	git("commit", "-qam", "code alone")
	if got := drift(); got != "internal/export/csv.go" {
		t.Fatalf("code committed after the entry: drift = %q", got)
	}
}

func TestIndexFiles(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entry := func(typ, title, desc string) string {
		return "---\ntype: " + typ + "\ntitle: " + title + "\ndescription: " + desc + "\n---\n\nBody.\n"
	}
	write(".prep/knowledge/overview.md", entry("overview", "Overview", "Entry point."))
	write(".prep/knowledge/components/cli.md", entry("component", "CLI", "Commands."))
	write(".prep/knowledge/components/tui/app.md", entry("component", "TUI app", "Screens."))
	write(".prep/knowledge/log.md", "# Log\n\n- 2026-10-06 started\n")
	write(".prep/knowledge/stale/index.md", "# stale\n")
	s := &Store{Root: root}

	entries, diags, err := s.Load(false)
	if err != nil || len(entries) != 3 {
		t.Fatalf("load: %v, %d entries (log.md must not be an entry)", err, len(entries))
	}
	if len(diags) != 4 { // three missing indexes and one unneeded
		t.Fatalf("index diagnostics = %v", diags)
	}
	touched, err := s.WriteIndexes(false)
	if err != nil || len(touched) != 4 {
		t.Fatalf("WriteIndexes: %v %v", touched, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".prep/knowledge/stale/index.md")); err == nil {
		t.Fatal("index in a directory without entries was kept")
	}
	b, _ := os.ReadFile(filepath.Join(root, ".prep/knowledge/index.md"))
	want := "---\nokf_version: \"0.2\"\n---\n\n# Knowledge base\n\n## Entries\n\n* [Overview](/overview.md) - Entry point.\n\n## Directories\n\n* [components](/components/index.md) - 2 entries\n"
	if string(b) != want {
		t.Fatalf("root index =\n%s\nwant\n%s", b, want)
	}
	b, _ = os.ReadFile(filepath.Join(root, ".prep/knowledge/components/index.md"))
	if !strings.Contains(string(b), "# components\n\n## Entries\n\n* [CLI](/components/cli.md) - Commands.\n\n## Directories\n\n* [tui](/components/tui/index.md) - 1 entry\n") {
		t.Fatalf("components index =\n%s", b)
	}
	if _, diags, _ := s.Load(false); len(diags) != 0 {
		t.Fatalf("diagnostics after writing indexes: %v", diags)
	}
}

func TestDriftWithoutConfirmedCommit(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		if _, err := gitx.Run(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	write := func(rel, s string) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/export/csv.go", "package export\n")
	write(".prep/knowledge/components/export.md", "---\ntype: component\ntitle: Export\ndescription: CSV.\nscope:\n  - internal/export\n---\n\nBody.\n")
	s := &Store{Root: root}
	drift := func() string {
		t.Helper()
		entries, _, err := s.Load(true)
		if err != nil || len(entries) != 1 {
			t.Fatalf("load: %v %d", err, len(entries))
		}
		return strings.Join(entries[0].Drifted, " ") + entries[0].DriftErr
	}
	if got := drift(); got != "" {
		t.Fatalf("uncommitted entry: drift = %q", got)
	}
	git("add", ".")
	git("commit", "-qm", "entry")
	write("internal/export/csv.go", "package export\n\nfunc Stream() {}\n")
	git("commit", "-qam", "code")
	if got := drift(); got != "internal/export/csv.go" {
		t.Fatalf("code committed after an unconfirmed entry: drift = %q", got)
	}
}

func TestEntryFrontmatterClosesOnExactLine(t *testing.T) {
	e, err := parseEntry("/a.md", "\xef\xbb\xbf---\ntype: component\ntitle: A\ndescription: |\n  ----\n  ---x\n---\n\nBody.\n")
	if err != nil || e.Title != "A" || e.Description != "----\n---x" || e.Body != "Body." {
		t.Fatalf("entry = %+v, %v", e, err)
	}
}

func TestNewEntryKeepsUnreadableFile(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".prep/knowledge/components/x.md")
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte("no frontmatter, but someone's notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, typ, title, desc := "Body.", "component", "X", "X."
	s := &Store{Root: root}
	_, err := s.Render(&domain.KnowledgeEdit{Path: "/components/x.md", New: true, Body: &body, Type: &typ, Title: &title, Description: &desc})
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}
