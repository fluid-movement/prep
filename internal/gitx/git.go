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

// FilesInCommit lists the files a commit changed.
func FilesInCommit(dir, commit string) ([]string, error) {
	out, err := Run(dir, "show", "--name-only", "--format=", commit)
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// Stage adds paths (including deletions) to the index.
func Stage(dir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := Run(dir, append([]string{"add", "-A", "--"}, paths...)...)
	return err
}

// Commit commits exactly the given paths, leaving anything else staged alone.
func Commit(dir, message string, paths []string) error {
	if err := Stage(dir, paths); err != nil {
		return err
	}
	if out, _ := Run(dir, append([]string{"diff", "--cached", "--name-only", "--"}, paths...)...); out == "" {
		return nil
	}
	_, err := Run(dir, append([]string{"commit", "--quiet", "-m", message, "--only", "--"}, paths...)...)
	return err
}

// Head returns the full hash of the current commit.
func Head(dir string) (string, error) {
	out, err := Run(dir, "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}
