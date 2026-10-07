// Package gitx wraps the few git operations prep needs. Git is versioning,
// not the system of record: every caller tolerates git being unavailable.
package gitx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes git in dir and returns trimmed stdout.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out.String()), nil
}

// IsRepo reports whether dir is inside a git work tree.
func IsRepo(dir string) bool {
	out, err := Run(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// CommitExists reports whether ref resolves to a commit.
func CommitExists(dir, ref string) bool {
	_, err := Run(dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	return err == nil
}

// pathspec turns a scope pattern into a git pathspec.
func pathspec(scope string) string {
	if strings.ContainsAny(scope, "*?[") {
		return ":(glob)" + scope
	}
	return scope
}

// ChangedSince lists files under the scopes that differ between commit and
// the working tree.
func ChangedSince(dir, commit string, scopes []string) ([]string, error) {
	args := []string{"diff", "--name-only", commit, "--"}
	for _, s := range scopes {
		args = append(args, pathspec(s))
	}
	out, err := Run(dir, args...)
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// ChangedBetween lists files under the scopes that differ between two
// commits, ignoring the working tree.
func ChangedBetween(dir, from, to string, scopes []string) ([]string, error) {
	args := []string{"diff", "--name-only", from, to, "--"}
	for _, s := range scopes {
		args = append(args, pathspec(s))
	}
	out, err := Run(dir, args...)
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// LastCommit returns the newest commit that changed path, or "".
func LastCommit(dir, path string) string {
	out, _ := Run(dir, "log", "-1", "--format=%H", "--", path)
	return out
}

// AddedIn returns the commit that added path, or "" while it is not
// committed.
func AddedIn(dir, path string) string {
	out, _ := Run(dir, "log", "--diff-filter=A", "--format=%H", "--", path)
	lines := strings.Split(out, "\n")
	return lines[len(lines)-1] // the oldest, should the file have been re-added
}

// IsAncestor reports whether commit a is an ancestor of (or equal to) b.
func IsAncestor(dir, a, b string) bool {
	_, err := Run(dir, "merge-base", "--is-ancestor", a, b)
	return err == nil
}

// IsTracked reports whether path is in git's index.
func IsTracked(dir, path string) bool {
	_, err := Run(dir, "ls-files", "--error-unmatch", "--", path)
	return err == nil
}

// Untrack removes path from git's index, keeping the file.
func Untrack(dir, path string) error {
	_, err := Run(dir, "rm", "--cached", "--quiet", "--", path)
	return err
}

// Dirty reports whether path has changes not yet committed (staged or not).
func Dirty(dir, path string) bool {
	out, _ := Run(dir, "status", "--porcelain", "--", path)
	return out != ""
}

// FilesInCommit lists the files a commit changed.
func FilesInCommit(dir, commit string) ([]string, error) {
	out, err := Run(dir, "show", "--name-only", "--format=", commit)
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// Tracked drops the paths git ignores, such as the user's own
// .prep/config.yaml, so staging never fails on them.
func Tracked(dir string, paths []string) []string {
	if len(paths) == 0 {
		return paths
	}
	// check-ignore prints the ignored paths and exits 1 when there are none.
	out, _ := Run(dir, append([]string{"check-ignore", "--"}, paths...)...)
	ignored := map[string]bool{}
	for _, p := range strings.Split(out, "\n") {
		ignored[p] = true
	}
	var keep []string
	for _, p := range paths {
		if !ignored[p] {
			keep = append(keep, p)
		}
	}
	return keep
}

// Stage adds paths (including deletions) to the index.
func Stage(dir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := Run(dir, append([]string{"add", "-A", "--"}, paths...)...)
	return err
}

// Head returns the full hash of the current commit.
func Head(dir string) (string, error) {
	out, err := Run(dir, "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}
