package okf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// OKFVersion is the Open Knowledge Format version the bundle follows.
const OKFVersion = "0.2"

// IndexFiles computes the OKF index.md of every bundle directory that holds
// entries, directly or below it, keyed by bundle path ("/index.md",
// "/components/index.md"). Without entries there are no index files.
func IndexFiles(entries []*domain.Entry) map[string]string {
	if len(entries) == 0 {
		return map[string]string{}
	}
	direct := map[string][]*domain.Entry{} // dir -> entries in it
	subdirs := map[string]map[string]bool{}
	count := map[string]int{} // dir -> entries at or below it
	for _, e := range entries {
		dir := path.Dir(e.Path)
		direct[dir] = append(direct[dir], e)
		for d := dir; ; d = path.Dir(d) {
			count[d]++
			if d == "/" {
				break
			}
			p := path.Dir(d)
			if subdirs[p] == nil {
				subdirs[p] = map[string]bool{}
			}
			subdirs[p][d] = true
		}
	}
	out := map[string]string{}
	for dir := range count {
		var b strings.Builder
		if dir == "/" {
			fmt.Fprintf(&b, "---\nokf_version: %q\n---\n\n# Knowledge base\n", OKFVersion)
		} else {
			fmt.Fprintf(&b, "# %s\n", strings.TrimPrefix(dir, "/"))
		}
		if es := direct[dir]; len(es) > 0 {
			sort.Slice(es, func(a, c int) bool { return es[a].Path < es[c].Path })
			b.WriteString("\n## Entries\n\n")
			for _, e := range es {
				title := e.Title
				if title == "" {
					title = path.Base(e.Path)
				}
				line := fmt.Sprintf("* [%s](%s)", title, e.Path)
				if d := strings.Join(strings.Fields(e.Description), " "); d != "" {
					line += " - " + d
				}
				b.WriteString(line + "\n")
			}
		}
		if len(subdirs[dir]) > 0 {
			var ds []string
			for d := range subdirs[dir] {
				ds = append(ds, d)
			}
			sort.Strings(ds)
			b.WriteString("\n## Directories\n\n")
			for _, d := range ds {
				noun := "entries"
				if count[d] == 1 {
					noun = "entry"
				}
				fmt.Fprintf(&b, "* [%s](%s/index.md) - %d %s\n", path.Base(d), d, count[d], noun)
			}
		}
		key := path.Join(dir, "index.md")
		out[key] = b.String()
	}
	return out
}

// indexState compares the index files on disk with the expected ones and
// returns the bundle paths that are missing, outdated or no longer needed.
func (s *Store) indexState(entries []*domain.Entry) (want map[string]string, stale []string, err error) {
	want = IndexFiles(entries)
	base := filepath.Join(s.Root, filepath.FromSlash(Dir))
	have := map[string]bool{}
	err = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return filepath.SkipAll
			}
			return err
		}
		if !d.IsDir() && d.Name() == "index.md" {
			rel, _ := filepath.Rel(base, p)
			have["/"+filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	for k, content := range want {
		b, rerr := os.ReadFile(s.file(k))
		if rerr != nil || string(b) != content {
			stale = append(stale, k)
		}
	}
	for k := range have {
		if _, ok := want[k]; !ok {
			stale = append(stale, k)
		}
	}
	sort.Strings(stale)
	return want, stale, nil
}

// WriteIndexes brings the index files in line with the entries: writes
// missing or outdated ones and removes those in directories without
// entries. With dryRun it only reports what would change.
func (s *Store) WriteIndexes(dryRun bool) ([]string, error) {
	entries, _, err := s.Load(false)
	if err != nil {
		return nil, err
	}
	want, stale, err := s.indexState(entries)
	if err != nil {
		return nil, err
	}
	var touched []string
	for _, k := range stale {
		touched = append(touched, Dir+k)
		if dryRun {
			continue
		}
		content, keep := want[k]
		if !keep {
			if err := os.Remove(s.file(k)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return touched, err
			}
			continue
		}
		if err := writeAtomic(s.file(k), content); err != nil {
			return touched, err
		}
	}
	return touched, nil
}

func writeAtomic(p, content string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".prep-tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p)
}
