// Package gitx wraps the few git operations prep needs. Git is versioning,
// not the system of record: every caller tolerates git being unavailable.
package gitx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes git in dir and returns trimmed stdout. Optional locks are
// off, so a read such as git status never takes index.lock from under a
// commit the person runs at the same time.
func Run(dir string, args ...string) (string, error) {
	return run(dir, "", args...)
}

// run is Run with text on standard input.
func run(dir, stdin string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	cmd.Stdin = strings.NewReader(stdin)
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

// paths splits the NUL-separated output of a -z command: paths come
// verbatim, where the line form quotes non-ASCII and special characters.
func paths(out string) []string {
	var list []string
	for p := range strings.SplitSeq(out, "\x00") {
		if p != "" {
			list = append(list, p)
		}
	}
	return list
}

// AddedIn returns the commit that added path, or "" while it is not
// committed.
func AddedIn(dir, path string) string {
	out, _ := Run(dir, "log", "--diff-filter=A", "--format=%H", "--", path)
	lines := strings.Split(out, "\n")
	return lines[len(lines)-1] // the oldest, should the file have been re-added
}

// FilesInCommit lists the files a commit changed.
func FilesInCommit(dir, commit string) ([]string, error) {
	out, err := Run(dir, "show", "--name-only", "-z", "--format=", commit)
	return paths(out), err
}

// Tracked drops the paths git ignores, such as the user's own
// .prep/config.yaml, so staging never fails on them.
func Tracked(dir string, list []string) []string {
	if len(list) == 0 {
		return list
	}
	// check-ignore prints the ignored paths and exits 1 when there are none.
	out, _ := run(dir, strings.Join(list, "\x00")+"\x00", "check-ignore", "-z", "--stdin")
	ignored := map[string]bool{}
	for _, p := range paths(out) {
		ignored[p] = true
	}
	var keep []string
	for _, p := range list {
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

// Remote returns the URL of dir's origin remote, or "" when it has none.
func Remote(dir string) string {
	out, err := Run(dir, "config", "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
