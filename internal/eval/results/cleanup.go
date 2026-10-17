package results

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// removeWorkspace only deletes retained run environments, never the per-run
// result directory also stored in RunResult.Dir for runs that were not kept.
func removeWorkspace(dir string) error {
	dir = filepath.Clean(dir)
	parent := filepath.Dir(dir)
	if !filepath.IsAbs(dir) || !strings.HasPrefix(filepath.Base(dir), "prep-eval-") ||
		!strings.HasPrefix(filepath.Base(parent), "prep-eval-work-") {
		return nil
	}
	info, err := os.Lstat(parent)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() { // Do not follow a replaced workspace parent symlink.
		return nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	_ = os.Remove(parent) // Remove the parent only once all its environments are gone.
	return nil
}

// CleanWorkspaces removes temporary eval workspace parents older than before.
// A zero cutoff includes all workspace parents, regardless of age.
func CleanWorkspaces(before time.Time) ([]string, error) {
	root := os.TempDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "prep-eval-work-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return removed, err
		}
		if !before.IsZero() && !info.ModTime().Before(before) {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		if err := os.RemoveAll(dir); err != nil {
			return removed, err
		}
		removed = append(removed, dir)
	}
	return removed, nil
}
