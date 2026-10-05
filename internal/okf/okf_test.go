package okf

import (
	"os"
	"path/filepath"
	"testing"

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
	write("internal/export/csv.go", "package export\n\nfunc Stream() {}\n")
	entries, _, _ = s.Load(true)
	if len(entries[0].Drifted) != 1 || entries[0].Drifted[0] != "internal/export/csv.go" {
		t.Fatalf("drift = %v", entries[0].Drifted)
	}
}
