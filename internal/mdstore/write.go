package mdstore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// write stores content atomically (temp file plus rename) after a
// compare-and-swap check against the hash seen at load time. Files never
// read in this run must not exist yet.
func (s *Store) write(rel, content string) error {
	unlock, err := s.lock(true)
	if err != nil {
		return err
	}
	defer unlock()
	if err := s.recoverTransaction(); err != nil {
		return err
	}
	if err := s.cas(rel); err != nil {
		return err
	}
	before := s.hashes[rel]
	if _, ok := s.hashes[rel]; !ok {
		before = ""
	}
	if err := s.applyTransaction([]transactionEntry{{Path: rel, Before: before, After: hashOf([]byte(content))}}, map[string][]byte{rel: []byte(content)}); err != nil {
		return err
	}
	s.hashes[rel] = hashOf([]byte(content))
	return nil
}

func (s *Store) remove(rel string) error {
	unlock, err := s.lock(true)
	if err != nil {
		return err
	}
	defer unlock()
	if err := s.recoverTransaction(); err != nil {
		return err
	}
	if err := s.cas(rel); err != nil {
		return err
	}
	before := s.hashes[rel]
	if _, ok := s.hashes[rel]; !ok {
		before = ""
	}
	if err := s.applyTransaction([]transactionEntry{{Path: rel, Before: before, Remove: true}}, nil); err != nil {
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

// Apply renders a change into files and writes them as a locked, recoverable
// roll-forward transaction. It returns repository-relative touched paths.
func (s *Store) Apply(c *domain.Change) ([]string, error) {
	unlock, err := s.lock(true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	if err := s.recoverTransaction(); err != nil {
		return nil, err
	}
	loadedIssueIDs := slices.Clone(s.loadedIDs)
	if s.hasSnapshot {
		currentIDs, err := s.readIssueInventory()
		if err != nil {
			return nil, err
		}
		if !slices.Equal(loadedIssueIDs, currentIDs) {
			return nil, &domain.Error{Code: domain.ErrConflict, Message: "issue directory inventory changed since it was read; re-run the command"}
		}
	}
	loaded := cloneHashes(s.hashes)
	touched, contents, err := s.render(c)
	if err != nil {
		return nil, err
	}
	// Validate the complete load read-set as well as any additional files
	// consulted by render. This prevents a stale transition from using an old
	// verification/ready record while changing a different file.
	checks := cloneHashes(s.hashes)
	for rel, before := range loaded {
		checks[rel] = before
	}
	for rel, want := range checks {
		if err := s.casExpected(rel, want); err != nil {
			return nil, err
		}
	}
	entries := make([]transactionEntry, 0, len(touched))
	data := make(map[string][]byte, len(touched))
	for k, rel := range touched {
		before, ok := loaded[rel]
		if !ok {
			before, ok = s.hashes[rel]
		}
		if !ok {
			before = ""
		}
		if contents[k] == nil {
			entries = append(entries, transactionEntry{Path: rel, Before: before, Remove: true})
		} else {
			body := []byte(*contents[k])
			entries = append(entries, transactionEntry{Path: rel, Before: before, After: hashOf(body)})
			data[rel] = body
		}
	}
	if err := s.applyTransaction(entries, data); err != nil {
		return touched, err
	}
	for _, e := range entries {
		s.hashes[e.Path] = e.After
	}
	if c.NewIssue != nil && s.hasSnapshot {
		s.loadedIDs = append(s.loadedIDs, c.IssueID)
		slices.Sort(s.loadedIDs)
	}
	return touched, nil
}

func cloneHashes(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (s *Store) casExpected(rel, want string) error {
	b, err := os.ReadFile(s.abs(rel))
	cur := ""
	if err == nil {
		cur = hashOf(b)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if cur != want {
		return &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("%s changed since it was read; re-run the command", rel)}
	}
	return nil
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
	if c.Links != nil || c.Policy != nil {
		rel := Dir + "/project.md"
		if c.PolicyHistory != nil {
			if err := s.cas(rel); err != nil {
				return nil, nil, err
			}
		}
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
		if c.Links != nil {
			p.Links = *c.Links
		}
		if c.Policy != nil {
			p.Policy = c.Policy
		}
		if c.PolicyHistory != nil {
			p.PolicyHistory = *c.PolicyHistory
		}
		out := renderProject(p, body)
		touched = append(touched, rel)
		contents = append(contents, &out)
	}
	if r := c.LinkRename; r != nil {
		rels, err := s.linkRefFiles()
		if err != nil {
			return nil, nil, err
		}
		for _, rel := range rels {
			raw, _, err := s.read(rel)
			if err != nil {
				return nil, nil, err
			}
			if out := domain.RenameLinkInText(raw, r[0], r[1]); out != raw {
				touched = append(touched, rel)
				contents = append(contents, &out)
			}
		}
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
		for _, n := range schemaFiles {
			if err := w(n, ""); err != nil {
				return nil, nil, err
			}
		}
	}
	if e := c.Edit; e != nil {
		rel := dir + "/issue.md"
		if err := s.cas(rel); err != nil {
			return nil, nil, err
		}
		raw, _, err := s.read(rel)
		if err != nil {
			return nil, nil, err
		}
		var i domain.Issue
		if err := parseIssue(raw, &i); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", rel, err)
		}
		i.Title, i.Kind, i.Parent, i.DependsOn, i.DependsOnApproval, i.Tags, i.Priority = e.Title, e.Kind, e.Parent, e.DependsOn, e.DependsOnApproval, e.Tags, e.Priority
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
	if c.Verification != nil {
		if err := w("verification.md", renderVerification(*c.Verification)); err != nil {
			return nil, nil, err
		}
	}
	if c.KnowledgeReviews != nil {
		if err := s.cas(dir + "/knowledge-reviews.md"); err != nil {
			return nil, nil, err
		}
		if err := w("knowledge-reviews.md", renderKnowledgeReviews(*c.KnowledgeReviews)); err != nil {
			return nil, nil, err
		}
	}
	if c.DependencyOverrides != nil {
		if err := s.cas(dir + "/dependency-overrides.md"); err != nil {
			return nil, nil, err
		}
		if err := w("dependency-overrides.md", renderDependencyOverrides(*c.DependencyOverrides)); err != nil {
			return nil, nil, err
		}
	}
	if c.Approvals != nil || c.DoDRevision != nil || c.OptOutRevisions != nil {
		var state approvalsFM
		if err := s.cas(dir + "/approvals.md"); err != nil {
			return nil, nil, err
		}
		raw, exists, err := s.read(dir + "/approvals.md")
		if err != nil {
			return nil, nil, err
		}
		if exists {
			state, err = parseApprovalState(raw)
			if err != nil {
				return nil, nil, err
			}
		}
		if c.Approvals != nil {
			state.Approvals = *c.Approvals
		}
		if c.DoDRevision != nil {
			state.DoDRevision = *c.DoDRevision
		}
		if c.OptOutRevisions != nil {
			state.OptOutRevisions = *c.OptOutRevisions
		}
		if err := w("approvals.md", renderApprovalState(state)); err != nil {
			return nil, nil, err
		}
	}
	if c.ResolutionHistory != nil {
		if err := w("resolution-history.md", renderResolutionHistory(*c.ResolutionHistory)); err != nil {
			return nil, nil, err
		}
	}
	if c.ResolutionReviews != nil {
		if err := w("resolution-reviews.md", renderResolutionReviews(*c.ResolutionReviews)); err != nil {
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
		if err := s.cas(rel); err != nil {
			return "", err
		}
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
		if err := w("history.md", fileText(normalize(raw)+"\n\n"+c.History)); err != nil {
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
	return touched, contents, nil
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
			{"approvals.md", func(raw string) (string, bool) {
				state, err := parseApprovalState(raw)
				if err != nil {
					return "", false
				}
				return renderApprovalState(state), true
			}},
			{"verification.md", func(raw string) (string, bool) {
				records, err := parseVerification(raw)
				if err != nil {
					return "", false
				}
				return renderVerification(records), true
			}},
			{"knowledge-reviews.md", func(raw string) (string, bool) {
				records, err := parseKnowledgeReviews(raw)
				if err != nil {
					return "", false
				}
				return renderKnowledgeReviews(records), true
			}},
			{"dependency-overrides.md", func(raw string) (string, bool) {
				records, err := parseDependencyOverrides(raw)
				if err != nil {
					return "", false
				}
				return renderDependencyOverrides(records), true
			}},
			{"resolution.md", func(raw string) (string, bool) {
				r, err := parseResolution(raw)
				if err != nil {
					return "", false
				}
				return renderResolution(r), true
			}},
			{"resolution-history.md", func(raw string) (string, bool) {
				rs, err := parseResolutionHistory(raw)
				if err != nil {
					return "", false
				}
				return renderResolutionHistory(rs), true
			}},
			{"resolution-reviews.md", func(raw string) (string, bool) {
				fm, body, ok, err := splitFrontmatter(raw)
				if err != nil || !ok || strings.TrimSpace(body) != "" {
					return "", false
				}
				var state resolutionReviewsFM
				if err := decodeStrict(fm, &state); err != nil {
					return "", false
				}
				if err := validateResolutionReviewChain(state.Reviews); err != nil {
					return "", false
				}
				return renderResolutionReviews(state.Reviews), true
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

// fixLegacyConfig converts a prep 0.1.0 configuration: both config files
// in the current format, config.yaml.dist shared (from the personal
// config.yaml when it is missing) and config.yaml ignored. Taking
// config.yaml out of git's index is the CLI's part.
func (s *Store) fixLegacyConfig() ([]string, error) {
	var changed []string
	ownRel, distRel, ignRel := Dir+"/"+ConfigOwn, Dir+"/"+ConfigDist, Dir+"/.gitignore"
	own, hasOwn, err := s.read(ownRel)
	if err != nil {
		return nil, err
	}
	dist, hasDist, err := s.read(distRel)
	if err != nil {
		return nil, err
	}
	if !hasOwn && !hasDist {
		return nil, nil
	}
	if conv, old := convertLegacyConfig(own); hasOwn && old {
		if err := s.write(ownRel, conv); err != nil {
			return changed, err
		}
		own, changed = conv, append(changed, ownRel)
	}
	if conv, old := convertLegacyConfig(dist); hasDist && old {
		if err := s.write(distRel, conv); err != nil {
			return changed, err
		}
		changed = append(changed, distRel)
	}
	if hasOwn && !hasDist {
		if err := s.write(distRel, own); err != nil {
			return changed, err
		}
		changed = append(changed, distRel)
	}
	if !s.IgnoresOwnConfig() {
		if err := s.addIgnore(ConfigOwn); err != nil {
			return changed, err
		}
		changed = append(changed, ignRel)
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

// Fix applies safe automatic repairs: canonical format, missing empty schema
// files, duplicate dependencies and repeated section headings whose copies
// hold content at most once.
func (s *Store) Fix() ([]string, error) {
	changed, err := s.Fmt(false)
	if err != nil {
		return changed, err
	}
	conf, err := s.fixLegacyConfig()
	changed = append(changed, conf...)
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

// linkRefFiles lists the files that can refer to a link: every issue's
// requirement and text records, and every knowledge entry.
func (s *Store) linkRefFiles() ([]string, error) {
	ids, err := s.issueIDs()
	if err != nil {
		return nil, err
	}
	var rels []string
	for _, id := range ids {
		for _, n := range []string{"issue.md", "context.md", "findings.md", "decisions.md", "acceptance.md", "history.md"} {
			rel := IssueDir(id) + "/" + n
			if _, err := os.Stat(s.abs(rel)); err == nil {
				rels = append(rels, rel)
			}
		}
	}
	root := s.abs(Dir + "/knowledge")
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		rel, err := filepath.Rel(s.Root, p)
		if err == nil {
			rels = append(rels, filepath.ToSlash(rel))
		}
		return err
	})
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	return rels, err
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

// SetSchema rewrites the schema version in project.md, keeping the rest.
func (s *Store) SetSchema(v int) error {
	return s.editProject(func(p *domain.Project) { p.Schema = v })
}

// SetIdentity writes the project's ID and name into project.md, keeping the
// rest; an empty value leaves that field as it is.
func (s *Store) SetIdentity(id, name string) error {
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
	// A frontmatter from an older schema may hold other keys; what decodes
	// is kept and the rest is dropped, as SetSchema always did.
	var f projectFM
	fm, _, _, _ := splitFrontmatter(raw)
	_ = decodeStrict(fm, &f)
	p := f.project()
	edit(&p)
	return s.write(rel, renderProject(p, body))
}
