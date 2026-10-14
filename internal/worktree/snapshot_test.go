package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFingerprintTracksScopedWorkAndCheckConfiguration(t *testing.T) {
	root := t.TempDir()
	write := func(path, body string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("src/work.go", "before")
	write("go.mod", "check config")
	write(".prep/knowledge/overview.md", "knowledge")
	fingerprint := func(scopes ...string) string {
		t.Helper()
		got, err := Fingerprint(root, scopes)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	initial := fingerprint("src", "go.mod")
	if got := fingerprint("go.mod", "src", "src/work.go"); got != initial {
		t.Fatal("scope order or overlap changed fingerprint")
	}
	write("unrelated.txt", "irrelevant")
	write(".prep/issues/task/history.md", "recorded report")
	write(".git/index", "staging")
	if got := fingerprint("src", "go.mod"); got != initial {
		t.Fatal("unrelated bookkeeping invalidated evidence")
	}
	write("go.mod", "new check config")
	if got := fingerprint("src", "go.mod"); got == initial {
		t.Fatal("check configuration change was missed")
	}
	initial = fingerprint("src", "go.mod")
	write("src/untracked.go", "new work")
	if got := fingerprint("src", "go.mod"); got == initial {
		t.Fatal("untracked work was missed")
	}
	initial = fingerprint(".")
	write(".prep/issues/task/history.md", "more bookkeeping")
	write(".git/index", "new staging")
	if got := fingerprint("."); got != initial {
		t.Fatal("whole-project scope included bookkeeping")
	}
	write(".prep/knowledge/overview.md", "updated knowledge")
	if got := fingerprint("."); got == initial {
		t.Fatal("knowledge changes were omitted")
	}
	missing := fingerprint("removed.go")
	write("removed.go", "restored")
	if got := fingerprint("removed.go"); got == missing {
		t.Fatal("missing-path change was missed")
	}
}

func TestFingerprintRejectsUnsafeScopes(t *testing.T) {
	root := t.TempDir()
	for _, scopes := range [][]string{nil, {""}, {"../outside"}, {root}, {".git"}, {".prep/issues"}, {".prep/project.md"}} {
		if _, err := Fingerprint(root, scopes); err == nil {
			t.Errorf("accepted unsafe scopes %q", scopes)
		}
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"link", "link/file", "."} {
		if _, err := Fingerprint(root, []string{scope}); err == nil {
			t.Errorf("accepted symlink scope %s", scope)
		}
	}
}
