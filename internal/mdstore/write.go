package mdstore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// write stores content atomically (temp file plus rename) after a
// compare-and-swap check against the hash seen at load time. Files never
// read in this run must not exist yet.
func (s *Store) write(rel, content string) error {
	if err := s.cas(rel); err != nil {
		return err
	}
	p := s.abs(rel)
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
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return err
	}
	s.hashes[rel] = hashOf([]byte(content))
	return nil
}

func (s *Store) remove(rel string) error {
	if err := s.cas(rel); err != nil {
		return err
	}
	if err := os.Remove(s.abs(rel)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	s.hashes[rel] = ""
	return nil
}

func (s *Store) cas(rel string) error {
	want, known := s.hashes[rel]
	b, err := os.ReadFile(s.abs(rel))
	cur := ""
	switch {
	case err == nil:
		cur = hashOf(b)
	case errors.Is(err, fs.ErrNotExist):
	default:
		return err
	}
	if !known {
		want = ""
	}
	if cur != want {
		return &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("%s changed since it was read; re-run the command", rel)}
	}
	return nil
}

// Apply renders a change into files and writes them. It returns the
// repository-relative paths it touched.
func (s *Store) Apply(c *domain.Change) ([]string, error) {
	dir := IssueDir(c.IssueID)
	var touched []string
	w := func(name, content string) error {
		rel := dir + "/" + name
		touched = append(touched, rel)
		return s.write(rel, content)
	}
	if c.NewIssue != nil {
		if _, err := os.Stat(s.abs(dir)); err == nil {
			return nil, &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("issue directory %s already exists", dir)}
		}
		i := *c.NewIssue
		i.Body = requirementBody(i.Prose)
		if err := w("issue.md", renderIssue(&i)); err != nil {
			return touched, err
		}
		for _, n := range schemaFiles {
			if err := w(n, ""); err != nil {
				return touched, err
			}
		}
	}
	if e := c.Edit; e != nil {
		rel := dir + "/issue.md"
		if err := s.cas(rel); err != nil {
			return touched, err
		}
		raw, _, err := s.read(rel)
		if err != nil {
			return touched, err
		}
		var i domain.Issue
		if err := parseIssue(raw, &i); err != nil {
			return touched, fmt.Errorf("%s: %w", rel, err)
		}
		i.Title, i.Kind, i.Parent, i.DependsOn, i.Tags, i.Priority = e.Title, e.Kind, e.Parent, e.DependsOn, e.Tags, e.Priority
		if e.Body != nil {
			i.Body = requirementBody(*e.Body)
		}
		if err := w("issue.md", renderIssue(&i)); err != nil {
			return touched, err
		}
	}
	if c.Baseline != nil {
		if err := w("baselines/"+c.Baseline.Name+".md", renderBaseline(c.Baseline)); err != nil {
			return touched, err
		}
	}
	if c.Ready != nil {
		if err := w("ready.md", renderReady(c.Ready)); err != nil {
			return touched, err
		}
	}
	if c.Claim != nil {
		if err := w("claim.md", renderClaim(c.Claim)); err != nil {
			return touched, err
		}
	}
	if c.RemoveClaim {
		touched = append(touched, dir+"/claim.md")
		if err := s.remove(dir + "/claim.md"); err != nil {
			return touched, err
		}
	}
	// current reads a record under compare-and-swap, so an edit never builds
	// on content that changed since load.
	current := func(name string) (string, error) {
		rel := dir + "/" + name
		if err := s.cas(rel); err != nil {
			return "", err
		}
		raw, _, err := s.read(rel)
		return raw, err
	}
	if c.Context != nil {
		if err := w("context.md", fileText(*c.Context)); err != nil {
			return touched, err
		}
	}
	if c.Findings != nil {
		if err := w("findings.md", fileText(*c.Findings)); err != nil {
			return touched, err
		}
	}
	if c.Decision != nil {
		raw, err := current("decisions.md")
		if err != nil {
			return touched, err
		}
		if err := w("decisions.md", fileText(normalize(raw)+"\n\n"+renderDecision(c.Decision))); err != nil {
			return touched, err
		}
	}
	if len(c.Acceptance) > 0 {
		raw, err := current("acceptance.md")
		if err != nil {
			return touched, err
		}
		if err := w("acceptance.md", fileText(applyAcceptance(raw, c.Acceptance))); err != nil {
			return touched, err
		}
	}
	if c.History != "" {
		raw, err := current("history.md")
		if err != nil {
			return touched, err
		}
		if err := w("history.md", fileText(normalize(raw)+"\n\n"+c.History)); err != nil {
			return touched, err
		}
	}
	if c.Resolution != nil {
		if err := w("resolution.md", renderResolution(c.Resolution)); err != nil {
			return touched, err
		}
	}
	return touched, nil
}

// Fmt rewrites every file in canonical format. With dryRun it only reports
// which files differ. Files that cannot be parsed are left alone.
func (s *Store) Fmt(dryRun bool) ([]string, error) {
	var changed []string
	try := func(rel string, canon func(string) (string, bool)) error {
		raw, ok, err := s.read(rel)
		if err != nil || !ok {
			return err
		}
		c, ok := canon(raw)
		if !ok || c == raw {
			return nil
		}
		changed = append(changed, rel)
		if dryRun {
			return nil
		}
		return s.write(rel, c)
	}
	always := func(f func(string) string) func(string) (string, bool) {
		return func(raw string) (string, bool) { return f(raw), true }
	}
	if err := try(Dir+"/project.md", func(raw string) (string, bool) { c := canonicalProject(raw); return c, c != "" }); err != nil {
		return changed, err
	}
	ids, err := s.issueIDs()
	if err != nil {
		return changed, err
	}
	for _, id := range ids {
		dir := IssueDir(id)
		steps := []struct {
			name  string
			canon func(string) (string, bool)
		}{
			{"issue.md", func(raw string) (string, bool) {
				i := &domain.Issue{}
				if parseIssue(raw, i) != nil {
					return "", false
				}
				return renderIssue(i), true
			}},
			{"acceptance.md", always(canonicalAcceptance)},
			{"context.md", always(fileText)},
			{"decisions.md", always(canonicalDecisions)},
			{"history.md", always(fileText)},
			{"findings.md", always(fileText)},
			{"ready.md", func(raw string) (string, bool) {
				r, err := parseReady(raw)
				if err != nil {
					return "", false
				}
				return renderReady(r), true
			}},
			{"claim.md", func(raw string) (string, bool) {
				c, err := parseClaim(raw)
				if err != nil {
					return "", false
				}
				return renderClaim(c), true
			}},
			{"resolution.md", func(raw string) (string, bool) {
				r, err := parseResolution(raw)
				if err != nil {
					return "", false
				}
				return renderResolution(r), true
			}},
		}
		for _, st := range steps {
			if err := try(dir+"/"+st.name, st.canon); err != nil {
				return changed, err
			}
		}
		bs, _ := os.ReadDir(s.abs(dir + "/baselines"))
		for _, b := range bs {
			name := strings.TrimSuffix(b.Name(), ".md")
			if b.IsDir() || !domain.ValidID(name) {
				continue
			}
			if err := try(dir+"/baselines/"+b.Name(), func(raw string) (string, bool) {
				bl, err := parseBaseline(name, raw)
				if err != nil {
					return "", false
				}
				return renderBaseline(&bl), true
			}); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}

// Fix applies safe automatic repairs: canonical format, missing empty schema
// files and duplicate dependencies.
func (s *Store) Fix() ([]string, error) {
	changed, err := s.Fmt(false)
	if err != nil {
		return changed, err
	}
	ids, err := s.issueIDs()
	if err != nil {
		return changed, err
	}
	for _, id := range ids {
		dir := IssueDir(id)
		if _, err := os.Stat(s.abs(dir + "/issue.md")); err != nil {
			continue
		}
		for _, n := range schemaFiles {
			rel := dir + "/" + n
			if _, err := os.Stat(s.abs(rel)); errors.Is(err, fs.ErrNotExist) {
				s.hashes[rel] = ""
				if err := s.write(rel, ""); err != nil {
					return changed, err
				}
				changed = append(changed, rel)
			}
		}
		rel := dir + "/issue.md"
		raw, _, err := s.read(rel)
		if err != nil {
			return changed, err
		}
		i := &domain.Issue{}
		if parseIssue(raw, i) != nil {
			continue
		}
		seen := map[string]bool{}
		var deps []string
		for _, d := range i.DependsOn {
			if !seen[d] {
				seen[d] = true
				deps = append(deps, d)
			}
		}
		if len(deps) != len(i.DependsOn) {
			i.DependsOn = deps
			if err := s.write(rel, renderIssue(i)); err != nil {
				return changed, err
			}
			changed = append(changed, rel)
		}
	}
	return changed, nil
}

func (s *Store) issueIDs() ([]string, error) {
	dirs, err := os.ReadDir(s.abs(Dir + "/issues"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, d := range dirs {
		if d.IsDir() && domain.ValidID(d.Name()) {
			ids = append(ids, d.Name())
		}
	}
	return ids, nil
}

// SetSchema rewrites the schema version in project.md, keeping its body.
func (s *Store) SetSchema(v int) error {
	rel := Dir + "/project.md"
	raw, ok, err := s.read(rel)
	if err != nil {
		return err
	}
	if !ok {
		return &domain.Error{Code: domain.ErrNoProject, Message: "project.md is missing"}
	}
	_, body, _, err := splitFrontmatter(raw)
	if err != nil {
		return err
	}
	return s.write(rel, renderProject(v, body))
}
