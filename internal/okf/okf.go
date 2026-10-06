// Package okf is the knowledge store adapter: an Open Knowledge Format
// (OKF v0.2) bundle of markdown entries in .prep/knowledge.
package okf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"gopkg.in/yaml.v3"
)

// Dir is the bundle location relative to the repository root.
const Dir = ".prep/knowledge"

// Store reads the knowledge bundle.
type Store struct {
	Root string // repository root
}

// Load reads every entry. With drift, it compares scoped paths against each
// entry's confirmed_commit using git.
func (s *Store) Load(drift bool) ([]*domain.Entry, []domain.Diagnostic, error) {
	base := filepath.Join(s.Root, filepath.FromSlash(Dir))
	var entries []*domain.Entry
	var diags []domain.Diagnostic
	useGit := drift && gitx.IsRepo(s.Root)
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") || d.Name() == "index.md" || d.Name() == "log.md" {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		bpath := "/" + filepath.ToSlash(rel)
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		e, perr := parseEntry(bpath, string(raw))
		if perr != nil {
			diags = append(diags, domain.Diagnostic{Code: domain.CodeKnowledgeFrontmatter, Severity: domain.SevError, Class: domain.ClassManual,
				File: Dir + bpath, Message: perr.Error(), Fix: "entries start with YAML frontmatter containing type, title and description"})
			return nil
		}
		if useGit && e.ConfirmedCommit != "" && len(e.Scope) > 0 {
			if !gitx.CommitExists(s.Root, e.ConfirmedCommit) {
				e.DriftErr = fmt.Sprintf("confirmed_commit %s is not in this repository", e.ConfirmedCommit)
			} else if changed, err := gitx.ChangedSince(s.Root, e.ConfirmedCommit, e.Scope); err != nil {
				e.DriftErr = err.Error()
			} else {
				e.Drifted = changed
			}
		}
		entries = append(entries, e)
		return nil
	})
	sort.Slice(entries, func(a, b int) bool { return entries[a].Path < entries[b].Path })
	if err != nil {
		return entries, diags, err
	}
	_, stale, ierr := s.indexState(entries)
	if ierr != nil {
		return entries, diags, ierr
	}
	for _, k := range stale {
		diags = append(diags, domain.Diagnostic{Code: domain.CodeKnowledgeIndex, Severity: domain.SevWarning, Class: domain.ClassFixable,
			File: Dir + k, Message: "index file is missing, outdated or no longer needed", Fix: "run prep fmt or prep fix to regenerate the OKF index files"})
	}
	return entries, diags, nil
}

var linkRe = regexp.MustCompile(`\]\(([^)\s]+)\)`)

var fencedRe = regexp.MustCompile("(?ms)^\\s*(```|~~~).*?^\\s*(```|~~~)\\s*$")
var codeSpanRe = regexp.MustCompile("`[^`\n]*`")

// StripCode removes fenced blocks and inline code so examples are not
// mistaken for links.
func StripCode(s string) string {
	return codeSpanRe.ReplaceAllString(fencedRe.ReplaceAllString(s, ""), "")
}

func parseEntry(bpath, raw string) (*domain.Entry, error) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(raw, "---\n") {
		return nil, fmt.Errorf("missing frontmatter")
	}
	end := strings.Index(raw[4:], "\n---")
	if end < 0 {
		return nil, fmt.Errorf("frontmatter is not closed with ---")
	}
	fm, body := raw[4:4+end+1], raw[4+end+4:]
	var m map[string]any
	if err := yaml.Unmarshal([]byte(fm), &m); err != nil {
		return nil, fmt.Errorf("frontmatter: %v", strings.TrimPrefix(err.Error(), "yaml: "))
	}
	e := &domain.Entry{Path: bpath, Size: len(raw)}
	str := func(k string) string {
		if v, ok := m[k]; ok && v != nil {
			return strings.TrimSpace(fmt.Sprint(v))
		}
		return ""
	}
	e.Type, e.Title, e.Description, e.Status, e.ConfirmedCommit = str("type"), str("title"), str("description"), str("status"), str("confirmed_commit")
	switch v := m["scope"].(type) {
	case string:
		e.Scope = []string{v}
	case []any:
		for _, x := range v {
			e.Scope = append(e.Scope, fmt.Sprint(x))
		}
	}
	seen := map[string]bool{}
	for _, l := range linkRe.FindAllStringSubmatch(StripCode(body), -1) {
		t := l[1]
		if k := strings.IndexByte(t, '#'); k >= 0 {
			t = t[:k]
		}
		if t == "" || strings.Contains(t, "://") || strings.HasPrefix(t, "mailto:") || !strings.HasSuffix(t, ".md") {
			continue
		}
		if !strings.HasPrefix(t, "/") {
			t = path.Join(path.Dir(bpath), t)
		}
		t = path.Clean(t)
		if !seen[t] {
			seen[t] = true
			e.Links = append(e.Links, t)
		}
	}
	return e, nil
}
