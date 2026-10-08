package gitx

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// repo makes a git repository with one commit of the given files.
func repo(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@example.com"}, {"config", "user.name", "t"}} {
		if _, err := Run(dir, args...); err != nil {
			t.Skip("git unavailable:", err)
		}
	}
	for _, f := range files {
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Run(dir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(dir, "commit", "-q", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPathsKeepNonASCIINames(t *testing.T) {
	names := []string{"src/größe.go", "src/plain.go", "src/with space.go"}
	dir := repo(t, names...)
	base, _ := Run(dir, "rev-parse", "HEAD")
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("b\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := ChangedSince(dir, base, []string{"src"})
	if err != nil || !slices.Equal(got, names) {
		t.Fatalf("ChangedSince = %q, %v", got, err)
	}
	files, err := FilesInCommit(dir, base)
	if err != nil || !slices.Equal(files, names) {
		t.Fatalf("FilesInCommit = %q, %v", files, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("src/größe.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(dir, "rm", "-q", "--cached", "src/größe.go"); err != nil {
		t.Fatal(err)
	}
	if keep := Tracked(dir, names); !slices.Equal(keep, names[1:]) {
		t.Fatalf("Tracked = %q", keep)
	}
}
