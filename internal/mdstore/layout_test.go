package mdstore

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

// issueFileNames lists the top-level entries of an issue directory.
func issueFileNames(t *testing.T, root, id string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, IssueDir(id)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestNewIssueWritesOnlyIssueFile(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	if got := issueFileNames(t, root, id); strings.Join(got, ",") != "issue.md" {
		t.Fatalf("new issue files = %v, want only issue.md", got)
	}
	_, issues, diags, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("missing record files reported: %+v", diags)
	}
	if len(issues) != 1 || issues[0].Context != "" || issues[0].History != "" {
		t.Fatalf("missing files did not read as empty: %+v", issues)
	}
}

func TestFixDoesNotCreateRecordFiles(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	if _, err := s.Fix(); err != nil {
		t.Fatal(err)
	}
	if got := issueFileNames(t, root, id); strings.Join(got, ",") != "issue.md" {
		t.Fatalf("fix created record files: %v", got)
	}
}

func TestEmptyWriteRemovesRecordFile(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	text := "Touches `internal/x.go`."
	if _, err := s.Apply(&domain.Change{IssueID: id, Context: &text}); err != nil {
		t.Fatal(err)
	}
	if !contains(issueFileNames(t, root, id), "context.md") {
		t.Fatal("context.md not written")
	}
	empty := ""
	if _, err := s.Apply(&domain.Change{IssueID: id, Context: &empty}); err != nil {
		t.Fatal(err)
	}
	if got := issueFileNames(t, root, id); strings.Join(got, ",") != "issue.md" {
		t.Fatalf("empty context left files %v, want only issue.md", got)
	}
}

func TestFmtRemovesWhitespaceOnlyRecordFile(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	path := filepath.Join(root, IssueDir(id), "history.md")
	if err := os.WriteFile(path, []byte("\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Fmt(false); err != nil {
		t.Fatal(err)
	}
	if contains(issueFileNames(t, root, id), "history.md") {
		t.Fatal("fmt left a whitespace-only history.md")
	}
}

func TestInitWritesHistoryMergeOnce(t *testing.T) {
	const line = ".prep/**/history.md merge=union"
	t.Run("new file", func(t *testing.T) {
		root := t.TempDir()
		if _, err := Open(root).Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
			t.Fatal(err)
		}
		if got := readAttributes(t, root); got != line+"\n" {
			t.Fatalf(".gitattributes = %q", got)
		}
	})
	t.Run("existing file", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte("*.png binary"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(root).Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
			t.Fatal(err)
		}
		if got := readAttributes(t, root); got != "*.png binary\n"+line+"\n" {
			t.Fatalf(".gitattributes = %q", got)
		}
	})
	t.Run("already present", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte("*.png binary\n"+line+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(root).Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
			t.Fatal(err)
		}
		if got := readAttributes(t, root); got != "*.png binary\n"+line+"\n" {
			t.Fatalf(".gitattributes duplicated the line: %q", got)
		}
	})
}

func readAttributes(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// TestHistoryAppendsMergeWithoutConflict runs real git merges: two branches
// each append a different line to the same history.md.
func TestHistoryAppendsMergeWithoutConflict(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		full := append([]string{"-C", root, "-c", "user.name=prep-test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)
		out, err := exec.Command("git", full...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Work log", Kind: domain.KindCode, Body: "Merge."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	history := filepath.Join(root, IssueDir(c.IssueID), "history.md")
	if err := os.WriteFile(history, []byte("- base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	base := git("rev-parse", "HEAD")

	appendLine := func(branch, line string) {
		t.Helper()
		git("checkout", "-q", "-b", branch, base)
		b, err := os.ReadFile(history)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(history, append(b, []byte(line+"\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
		git("commit", "-q", "-am", branch)
	}
	appendLine("branch-a", "- line from a")
	appendLine("branch-b", "- line from b")
	git("checkout", "-q", "branch-a")
	git("merge", "-q", "--no-edit", "branch-b")

	b, err := os.ReadFile(history)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"- base", "- line from a", "- line from b"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("merged history lacks %q:\n%s", want, b)
		}
	}
	if strings.Contains(string(b), "<<<<<<<") {
		t.Fatalf("merge left conflict markers:\n%s", b)
	}
}

// Emptying a file that was never written touches nothing.
func TestEmptyWriteOfMissingFileTouchesNothing(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	empty := ""
	touched, err := s.Apply(&domain.Change{IssueID: id, Context: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if len(touched) != 0 {
		t.Fatalf("touched = %v, want nothing", touched)
	}
}
