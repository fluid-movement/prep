package mdstore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// write stores content atomically (temporary file plus rename) under the
// checkout lock.
func (s *Store) write(rel, content string) error {
	unlock, err := s.lock(true)
	if err != nil {
		return err
	}
	defer unlock()
	return s.writeFile(rel, []byte(content))
}

func (s *Store) remove(rel string) error {
	unlock, err := s.lock(true)
	if err != nil {
		return err
	}
	defer unlock()
	return s.removeFile(rel)
}

// writeFile replaces rel through a synced temporary file in the same
// directory and a rename, so a reader sees the old or the new content.
func (s *Store) writeFile(rel string, content []byte) error {
	path := s.abs(rel)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, tempPrefix+"*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	fail := func(err error) error {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if _, err := f.Write(content); err != nil {
		return fail(err)
	}
	if err := f.Sync(); err != nil {
		return fail(err)
	}
	if err := f.Close(); err != nil {
		return fail(err)
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return fail(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fail(err)
	}
	return syncDirectory(dir)
}

func (s *Store) removeFile(rel string) error {
	path := s.abs(rel)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return syncDirectory(filepath.Dir(path))
}

// Apply renders a change into files and writes them one at a time, the
// state-defining file last, so an interrupted write leaves the previous state
// plus harmless extra content. It takes the checkout lock unless the caller
// holds it (Hold). It returns repository-relative touched paths in write order.
func (s *Store) Apply(c *domain.Change) ([]string, error) {
	unlock, err := s.lock(true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	touched, contents, err := s.render(c)
	if err != nil {
		return nil, err
	}
	var done []string
	for k, rel := range touched {
		if contents[k] == nil || *contents[k] == "" {
			// A record file exists only while it has content; an empty
			// body removes it, and a missing file reads as empty. Removing a
			// file that never existed touches nothing, so it is not reported
			// (staging a path git has never seen fails).
			if _, statErr := os.Stat(s.abs(rel)); errors.Is(statErr, fs.ErrNotExist) {
				continue
			}
			err = s.removeFile(rel)
		} else {
			err = s.writeFile(rel, []byte(*contents[k]))
		}
		if err != nil {
			return done, err
		}
		done = append(done, rel)
	}
	return done, nil
}

// writeRank orders a change's files: supporting records first, then the
// files that define the issue's state (baseline, ready, claim), then the
// resolution.
func writeRank(rel string) int {
	switch {
	case strings.HasSuffix(rel, "/resolution.md"):
		return 2
	case strings.HasSuffix(rel, "/ready.md"), strings.HasSuffix(rel, "/claim.md"), strings.Contains(rel, "/baselines/"):
		return 1
	}
	return 0
}

// orderForWrite sorts touched files (and their contents) by writeRank,
// keeping the render order within a rank. Removals come after every write:
// complete writes resolution.md before it removes claim.md, so an
// interruption leaves the issue done with a stale claim, never back in ready;
// release and reopen remove their state file last.
func orderForWrite(touched []string, contents []*string) {
	idx := make([]int, len(touched))
	for i := range idx {
		idx[i] = i
	}
	rank := func(k int) int {
		if contents[k] == nil || *contents[k] == "" {
			return 3
		}
		return writeRank(touched[k])
	}
	sort.SliceStable(idx, func(a, b int) bool { return rank(idx[a]) < rank(idx[b]) })
	t := make([]string, len(touched))
	c := make([]*string, len(contents))
	for i, k := range idx {
		t[i], c[i] = touched[k], contents[k]
	}
	copy(touched, t)
	copy(contents, c)
}

// render turns a change into the files it writes, in order, with their
// contents; nil content removes the file.
func (s *Store) render(c *domain.Change) (touched []string, contents []*string, err error) {
	dir := IssueDir(c.IssueID)
	w := func(name, content string) error {
		touched = append(touched, dir+"/"+name)
		contents = append(contents, &content)
		return nil
	}
	if c.Config != nil {
		// The first save creates the user's own config from the dist
		// file, keeping its header.
		rel := Dir + "/" + ConfigOwn
		current, ok, err := s.read(rel)
		if err == nil && !ok {
			current, _, err = s.read(Dir + "/" + ConfigDist)
		}
		if err != nil {
			return nil, nil, err
		}
		out, err := renderConfig(current, *c.Config)
		if err != nil {
			return nil, nil, err
		}
		touched = append(touched, rel)
		contents = append(contents, &out)
	}
	if c.DoD != nil {
		rel := Dir + "/project.md"
		raw, ok, err := s.read(rel)
		if err == nil && !ok {
			err = &domain.Error{Code: domain.ErrNoProject, Message: "project.md is missing"}
		}
		if err != nil {
			return nil, nil, err
		}
		fm, body, _, err := splitFrontmatter(raw)
		if err != nil {
			return nil, nil, err
		}
		var f projectFM
		if err := decodeStrict(fm, &f); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", rel, err)
		}
		p := f.project()
		if c.DoD != nil {
			body = replaceProjectDoD(body, *c.DoD)
		}
		out := renderProject(p, body)
		touched = append(touched, rel)
		contents = append(contents, &out)
	}
	if c.NewIssue != nil {
		if _, err := os.Stat(s.abs(dir)); err == nil {
			return nil, nil, &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("issue directory %s already exists", dir)}
		}
		i := *c.NewIssue
		i.Body = requirementBody(i.Prose)
		if err := w("issue.md", renderIssue(&i)); err != nil {
			return nil, nil, err
		}
	}
	if e := c.Edit; e != nil {
		rel := dir + "/issue.md"
		raw, _, err := s.read(rel)
		if err != nil {
			return nil, nil, err
		}
		var i domain.Issue
		if err := parseIssue(raw, &i); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", rel, err)
		}
		i.Title, i.Kind, i.Parent, i.DependsOn, i.Tags, i.Priority = e.Title, e.Kind, e.Parent, e.DependsOn, e.Tags, e.Priority
		if e.Body != nil {
			i.Body = requirementBody(*e.Body)
		}
		if err := w("issue.md", renderIssue(&i)); err != nil {
			return nil, nil, err
		}
	}
	if c.Baseline != nil {
		if err := w("baselines/"+c.Baseline.Name+".md", renderBaseline(c.Baseline)); err != nil {
			return nil, nil, err
		}
	}
	if c.Ready != nil {
		if err := w("ready.md", renderReady(c.Ready)); err != nil {
			return nil, nil, err
		}
	}
	if c.Claim != nil {
		if err := w("claim.md", renderClaim(c.Claim)); err != nil {
			return nil, nil, err
		}
	}
	if c.RemoveClaim {
		touched = append(touched, dir+"/claim.md")
		contents = append(contents, nil)
	}
	// current reads a record under compare-and-swap, so an edit never builds
	// on content that changed since load.
	current := func(name string) (string, error) {
		rel := dir + "/" + name
		raw, _, err := s.read(rel)
		return raw, err
	}
	if c.Context != nil {
		if err := w("context.md", fileText(*c.Context)); err != nil {
			return nil, nil, err
		}
	}
	if c.Findings != nil {
		if err := w("findings.md", fileText(*c.Findings)); err != nil {
			return nil, nil, err
		}
	}
	if c.Decision != nil {
		raw, err := current("decisions.md")
		if err != nil {
			return nil, nil, err
		}
		if err := w("decisions.md", fileText(normalize(raw)+"\n\n"+renderDecision(c.Decision))); err != nil {
			return nil, nil, err
		}
	}
	if len(c.Acceptance) > 0 {
		raw, err := current("acceptance.md")
		if err != nil {
			return nil, nil, err
		}
		if err := w("acceptance.md", fileText(applyAcceptance(raw, c.Acceptance))); err != nil {
			return nil, nil, err
		}
	}
	if c.History != "" {
		raw, err := current("history.md")
		if err != nil {
			return nil, nil, err
		}
		entry := c.History
		if c.Op == domain.OpReopen {
			// Reopen keeps the removed resolution's text in the work log.
			old, err := current("resolution.md")
			if err != nil {
				return nil, nil, err
			}
			entry += "\n\n" + archivedResolution(old)
		}
		if err := w("history.md", fileText(normalize(raw)+"\n\n"+entry)); err != nil {
			return nil, nil, err
		}
	}
	if c.Resolution != nil {
		if err := w("resolution.md", renderResolution(c.Resolution)); err != nil {
			return nil, nil, err
		}
	}
	if c.RemoveResolution {
		touched = append(touched, dir+"/resolution.md")
		contents = append(contents, nil)
	}
	orderForWrite(touched, contents)
	return touched, contents, nil
}

// Fmt rewrites every file in canonical format. With dryRun it only reports
// which files differ. Files that cannot be parsed are left alone.
func (s *Store) Fmt(dryRun bool) ([]string, error) {
	if !dryRun {
		release, err := s.Hold()
		if err != nil {
			return nil, err
		}
		defer release()
	}
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
		if c == "" {
			return s.remove(rel)
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
			if b.IsDir() || !domain.ValidStamp(name) {
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

// addIgnore appends a line to .prep/.gitignore.
func (s *Store) addIgnore(line string) error {
	rel := Dir + "/.gitignore"
	ign, _, err := s.read(rel)
	if err != nil {
		return err
	}
	if ign != "" && !strings.HasSuffix(ign, "\n") {
		ign += "\n"
	}
	return s.write(rel, ign+line+"\n")
}

// Fix applies safe automatic repairs: canonical format, duplicate dependencies and repeated section headings whose copies
// hold content at most once.
func (s *Store) Fix() ([]string, error) {
	release, err := s.Hold()
	if err != nil {
		return nil, err
	}
	defer release()
	changed, err := s.Fmt(false)
	if err != nil {
		return changed, err
	}
	if s.localUnignored() {
		if err := s.addIgnore(localIgnore); err != nil {
			return changed, err
		}
		changed = append(changed, Dir+"/.gitignore")
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
		body, merged := mergeDuplicateSections(i.Body)
		if len(deps) != len(i.DependsOn) || merged {
			i.DependsOn, i.Body = deps, normalize(body)
			if err := s.write(rel, renderIssue(i)); err != nil {
				return changed, err
			}
			changed = append(changed, rel)
		}
		for _, f := range []struct {
			name  string
			canon func(string) string
		}{{"acceptance.md", canonicalAcceptance}, {"context.md", fileText}, {"findings.md", fileText}} {
			rel, canon := dir+"/"+f.name, f.canon
			raw, ok, err := s.read(rel)
			if err != nil {
				return changed, err
			}
			if !ok {
				continue
			}
			if text, merged := mergeDuplicateSections(normalize(raw)); merged {
				if err := s.write(rel, canon(text)); err != nil {
					return changed, err
				}
				changed = append(changed, rel)
			}
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

// SetIdentity writes the project's ID and name into project.md, keeping the
// rest; an empty value leaves that field as it is.
func (s *Store) SetIdentity(id, name string) error {
	release, err := s.Hold()
	if err != nil {
		return err
	}
	defer release()
	return s.editProject(func(p *domain.Project) {
		if id != "" {
			p.ID = id
		}
		if name != "" {
			p.Name = name
		}
	})
}

// editProject rewrites project.md's frontmatter through edit, keeping its
// body.
func (s *Store) editProject(edit func(*domain.Project)) error {
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
	var f projectFM
	fm, _, _, _ := splitFrontmatter(raw)
	if err := decodeStrict(fm, &f); err != nil {
		return err
	}
	p := f.project()
	edit(&p)
	return s.write(rel, renderProject(p, body))
}

// archivedResolution quotes a removed resolution.md in a fenced block whose
// fence is longer than any backtick run in the record.
func archivedResolution(raw string) string {
	fence := "```"
	for strings.Contains(raw, fence) {
		fence += "`"
	}
	return fence + "\n" + normalize(raw) + "\n" + fence
}
