// Package worktree fingerprints explicitly scoped work without requiring Git.
package worktree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Fingerprint includes file names, modes and contents, including untracked
// files and missing paths. Git metadata and prep bookkeeping are excluded;
// knowledge content is work, not bookkeeping. Symlinks are rejected instead
// of silently omitting their targets or following paths outside the project.
func Fingerprint(root string, scopes []string) (string, error) {
	if len(scopes) == 0 {
		return "", fmt.Errorf("at least one work scope is required")
	}
	paths := map[string]bool{}
	for _, scope := range scopes {
		if scope == "" || filepath.IsAbs(scope) {
			return "", fmt.Errorf("scope must be project-relative: %q", scope)
		}
		clean := filepath.Clean(scope)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("scope escapes project: %q", scope)
		}
		if excluded(filepath.ToSlash(clean)) {
			return "", fmt.Errorf("scope selects bookkeeping: %q", scope)
		}
		// Reject symlinks in intermediate components, not only final paths.
		current := root
		for _, part := range strings.Split(clean, string(filepath.Separator)) {
			current = filepath.Join(current, part)
			st, err := os.Lstat(current)
			if os.IsNotExist(err) {
				break
			}
			if err != nil {
				return "", err
			}
			if st.Mode()&os.ModeSymlink != 0 {
				return "", fmt.Errorf("symlink in scope %q", scope)
			}
		}
		err := filepath.WalkDir(filepath.Join(root, clean), func(path string, d fs.DirEntry, err error) error {
			rel, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			rel = filepath.ToSlash(rel)
			if excluded(rel) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if os.IsNotExist(err) {
				paths[rel] = true
				return nil
			}
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink in scope: %s", rel)
			}
			paths[rel] = true
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	names := make([]string, 0, len(paths))
	for path := range paths {
		names = append(names, path)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, path := range names {
		full := filepath.Join(root, filepath.FromSlash(path))
		st, err := os.Lstat(full)
		if os.IsNotExist(err) {
			fmt.Fprintf(h, "missing %q\n", path)
			continue
		}
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "path %q mode %d\n", path, st.Mode())
		if st.IsDir() {
			continue
		}
		if !st.Mode().IsRegular() {
			return "", fmt.Errorf("non-regular scoped file: %s", path)
		}
		f, err := os.Open(full)
		if err != nil {
			return "", err
		}
		fh := sha256.New()
		_, err = io.Copy(fh, f)
		closeErr := f.Close()
		if err != nil {
			return "", err
		}
		if closeErr != nil {
			return "", closeErr
		}
		fmt.Fprintf(h, "content %x\n", fh.Sum(nil))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func excluded(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == ".git" {
			return true
		}
	}
	return strings.HasPrefix(path, ".prep/") && path != ".prep/knowledge" && !strings.HasPrefix(path, ".prep/knowledge/")
}
