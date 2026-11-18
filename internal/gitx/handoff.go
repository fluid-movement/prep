package gitx

import (
	"fmt"
	"strings"
)

// RequirePushed verifies the live upstream, not a possibly stale tracking ref.
func RequirePushed(dir string) error {
	remote, err := Run(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil {
		return fmt.Errorf("HEAD has no upstream; push the branch and set its upstream first: %w", err)
	}
	branch, err := Run(dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return err
	}
	remoteName, err := Run(dir, "config", "--get", "branch."+branch+".remote")
	if err != nil {
		return err
	}
	ref, err := Run(dir, "config", "--get", "branch."+branch+".merge")
	if err != nil {
		return err
	}
	if _, err := Run(dir, "fetch", "--no-tags", remoteName, ref); err != nil {
		return fmt.Errorf("cannot verify pushed HEAD against %s: %w", remote, err)
	}
	if _, err := Run(dir, "merge-base", "--is-ancestor", "HEAD", "FETCH_HEAD"); err != nil {
		return fmt.Errorf("HEAD is not pushed to %s; push first", remote)
	}
	return nil
}

// CloudBranches fetches origin and lists only its unmerged claude/* branches.
func CloudBranches(dir string) ([]string, error) {
	heads, err := Run(dir, "ls-remote", "--heads", "origin", "refs/heads/claude/*")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(heads, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "refs/heads/")
		ref := "refs/remotes/origin/" + name
		if _, err := Run(dir, "fetch", "--no-tags", "origin", "+"+fields[1]+":"+ref); err != nil {
			return nil, err
		}
		if _, err := Run(dir, "merge-base", "--is-ancestor", ref, "HEAD"); err != nil {
			branches = append(branches, ref)
		}
	}
	return branches, nil
}

// BranchFiles lists changes since the branch diverged from HEAD.
func BranchFiles(dir, ref string) ([]string, error) {
	out, err := Run(dir, "diff", "--name-only", "-z", "HEAD..."+ref, "--", ".prep/issues/", ".prep/knowledge/findings/")
	return paths(out), err
}

func Merge(dir, ref string) error {
	_, err := Run(dir, "merge", "--no-edit", ref)
	return err
}
