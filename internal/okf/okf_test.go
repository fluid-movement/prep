package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	entries, diags, err := s.Load()
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
	if _, diags, _ := s.Load(); len(diags) != 0 {
		t.Fatalf("diagnostics after writing indexes: %v", diags)
	}
}

func TestEntryFrontmatterClosesOnExactLine(t *testing.T) {
	e, err := parseEntry("/a.md", "\xef\xbb\xbf---\ntype: component\ntitle: A\ndescription: |\n  ----\n  ---x\n---\n\nBody.\n")
	if err != nil || e.Title != "A" || e.Description != "----\n---x" || e.Body != "Body." {
		t.Fatalf("entry = %+v, %v", e, err)
	}
}

func TestIndexEscapesLinkText(t *testing.T) {
	if got := linkText.Replace(`A [draft] \ note`); got != `A \[draft\] \\ note` {
		t.Fatalf("escaped = %q", got)
	}
}
