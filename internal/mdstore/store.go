// Package mdstore is the markdown storage adapter: one directory per issue
// under .prep/issues, records as markdown files with YAML frontmatter.
package mdstore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
	"gopkg.in/yaml.v3"
)

// Dir is the project data directory inside the repository.
const Dir = ".prep"

// Store reads and writes a .prep tree rooted at Root (the repository root).
type Store struct {
	Root string
	held bool // Hold has the exclusive checkout lock; Load and Apply must not lock again
}

// Open returns a store for the repository root.
func Open(root string) *Store { return &Store{Root: root} }

// Find walks up from dir to the directory containing .prep.
func Find(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(dir, Dir)); err == nil && st.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", &domain.Error{Code: domain.ErrNoProject, Message: "no .prep directory found; run prep init"}
		}
		dir = parent
	}
}

// Rel converts an absolute path below Root to a slash path relative to Root.
func (s *Store) Rel(p string) string {
	r, err := filepath.Rel(s.Root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

func (s *Store) abs(rel string) string { return filepath.Join(s.Root, filepath.FromSlash(rel)) }

// IssueDir returns the repository-relative directory of an issue.
func IssueDir(id string) string { return Dir + "/issues/" + id }

// read loads a file; the bool reports whether it exists.
func (s *Store) read(rel string) (string, bool, error) {
	b, err := os.ReadFile(s.abs(rel))
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

// Allowed top-level entries.
var (
	issueFiles   = map[string]bool{"issue.md": true, "acceptance.md": true, "context.md": true, "decisions.md": true, "history.md": true, "findings.md": true, "ready.md": true, "claim.md": true, "resolution.md": true}
	issueDirs    = map[string]bool{"baselines": true}
	projectFiles = map[string]bool{"project.md": true, "config.yaml": true, "config.yaml.dist": true, "issues": true, "knowledge": true, ".gitignore": true, activity.Dir: true}
)

// Load reads the project and every issue.
func (s *Store) Load() (domain.Project, []*domain.Issue, []domain.Diagnostic, error) {
	unlock, err := s.lock(false)
	if err != nil {
		return domain.Project{}, nil, nil, err
	}
	if unlock != nil {
		defer unlock()
	}
	return s.loadUnlocked()
}

// loadUnlocked reads the project while the caller holds the checkout lock.
func (s *Store) loadUnlocked() (domain.Project, []*domain.Issue, []domain.Diagnostic, error) {
	var diags []domain.Diagnostic
	diag := func(code string, sev domain.Severity, class domain.Class, issue, file, fix, format string, a ...any) {
		diags = append(diags, domain.Diagnostic{Code: code, Severity: sev, Class: class, Issue: issue, File: file, Message: fmt.Sprintf(format, a...), Fix: fix})
	}

	var project domain.Project
	raw, ok, err := s.read(Dir + "/project.md")
	switch {
	case err != nil:
		return project, nil, nil, err
	case !ok:
		diag(domain.CodeProjectMissing, domain.SevError, domain.ClassGuided, "", "project.md", "run prep init", "project.md is missing")
	default:
		project, err = parseProject(raw)
		if err != nil {
			diag(domain.CodeProjectMissing, domain.SevError, domain.ClassManual, "", "project.md", "fix the frontmatter", "project.md: %v", err)
		} else if c := canonicalProject(raw); c != "" && c != raw {
			diag(domain.CodeNotCanonical, domain.SevWarning, domain.ClassFixable, "", "project.md", "run prep fmt", "not in canonical format")
		}
	}
	// Both files are checked: the dist file is what a fresh clone uses, even
	// when this user's own config replaces it.
	for _, name := range []string{ConfigDist, ConfigOwn} {
		if _, ok, err := s.loadConfigFile(name); ok && err != nil {
			// The diagnostic names the file; the message need not repeat it.
			diag(domain.CodeConfigInvalid, domain.SevError, domain.ClassManual, "", name, "fix "+name, "%s", strings.TrimPrefix(err.Error(), name+": "))
		}
	}
	project.Config, _ = s.loadConfigUnlocked()
	if s.localUnignored() {
		diag(domain.CodeLocalTracked, domain.SevWarning, domain.ClassFixable, "", ".gitignore", "run prep fix, then commit", ".prep/.gitignore does not ignore %s, the per-machine activity stream", localIgnore)
	}

	entries, err := os.ReadDir(s.abs(Dir))
	if err != nil {
		return project, nil, nil, err
	}
	for _, e := range entries {
		if !projectFiles[e.Name()] && !strings.HasPrefix(e.Name(), tempPrefix) {
			diag(domain.CodeProjectUnknown, domain.SevError, domain.ClassManual, "", e.Name(), "remove the file; schema files are fixed", "unknown entry in .prep: %s", e.Name())
		}
	}

	dirs, err := os.ReadDir(s.abs(Dir + "/issues"))
	if errors.Is(err, fs.ErrNotExist) {
		return project, nil, diags, nil
	}
	if err != nil {
		return project, nil, nil, err
	}
	var issues []*domain.Issue
	for _, d := range dirs {
		if strings.HasPrefix(d.Name(), ".") {
			continue
		}
		if !d.IsDir() || !domain.ValidID(d.Name()) {
			diag(domain.CodeBadID, domain.SevError, domain.ClassManual, "", "issues/"+d.Name(), "issue directories are named by their ID; use prep new", "%s is not a valid issue directory: issue IDs are ULIDs (26 characters of Crockford base32)", d.Name())
			continue
		}
		i, ds, err := s.loadIssue(d.Name())
		if err != nil {
			return project, nil, nil, err
		}
		diags = append(diags, ds...)
		if i != nil {
			issues = append(issues, i)
		}
	}
	return project, issues, diags, nil
}

func (s *Store) loadIssue(id string) (*domain.Issue, []domain.Diagnostic, error) {
	var diags []domain.Diagnostic
	diag := func(code string, sev domain.Severity, class domain.Class, file, fix, format string, a ...any) {
		diags = append(diags, domain.Diagnostic{Code: code, Severity: sev, Class: class, Issue: id, File: file, Message: fmt.Sprintf(format, a...), Fix: fix})
	}
	dir := IssueDir(id)
	i := &domain.Issue{ID: id}
	entries, err := os.ReadDir(s.abs(dir))
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, tempPrefix) {
			continue // a write in flight, or left by a crash
		}
		i.Files = append(i.Files, n)
		switch {
		case e.IsDir() && issueDirs[n]:
		case !e.IsDir() && issueFiles[n]:
		default:
			diag(domain.CodeUnknownFile, domain.SevError, domain.ClassManual, n, "remove the unknown entry from the issue directory", "unknown entry %s", n)
		}
	}
	sort.Strings(i.Files)

	raw, ok, err := s.read(dir + "/issue.md")
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		diag(domain.CodeIssueMissing, domain.SevError, domain.ClassManual, "issue.md", "restore issue.md from git or remove the directory", "issue.md is missing")
		return nil, diags, nil
	}
	if err := parseIssue(raw, i); err != nil {
		diag(domain.CodeFrontmatter, domain.SevError, domain.ClassManual, "issue.md", "fix the frontmatter: title, kind, parent, depends_on", "%v", err)
		return nil, diags, nil
	}
	if renderIssue(i) != raw {
		diag(domain.CodeNotCanonical, domain.SevWarning, domain.ClassFixable, "issue.md", "run prep fmt", "not in canonical format")
	}
	// Repeated ## headings: prep reads only the first Open questions section,
	// so questions under another copy would not block define.
	dupes := func(name, text string) {
		for _, d := range duplicateHeadings(text) {
			sev, why := domain.SevWarning, ""
			if name == "issue.md" && d.key == "open questions" {
				sev, why = domain.SevError, "; prep reads only the first, so questions under the others do not block define"
			}
			class, fix := domain.ClassGuided, "move the content under one heading and delete the other copies"
			if d.mergeable {
				class, fix = domain.ClassFixable, "run prep fix to merge them"
			}
			diag(domain.CodeDuplicateHeading, sev, class, name, fix, "section %q appears %d times%s", d.heading, d.count, why)
		}
	}
	dupes("issue.md", i.Body)

	plain := func(name string, canon func(string) string) (string, bool, error) {
		raw, ok, err := s.read(dir + "/" + name)
		if err == nil && ok && canon(raw) != raw {
			diag(domain.CodeNotCanonical, domain.SevWarning, domain.ClassFixable, name, "run prep fmt", "not in canonical format")
		}
		return raw, ok, err
	}
	if raw, ok, err := plain("acceptance.md", canonicalAcceptance); err != nil {
		return nil, nil, err
	} else if ok {
		parseAcceptance(raw, i)
		dupes("acceptance.md", normalize(raw))
	}
	if raw, ok, err := plain("context.md", fileText); err != nil {
		return nil, nil, err
	} else if ok {
		parseContext(raw, i)
		dupes("context.md", normalize(raw))
	}
	if raw, ok, err := plain("history.md", fileText); err != nil {
		return nil, nil, err
	} else if ok {
		i.History = normalize(raw)
	}

	for _, name := range []string{"decisions.md", "findings.md"} {
		if i.HasFile(name) {
			diag(domain.CodeLegacyKnowledge, domain.SevWarning, domain.ClassManual, name, "convert to linked knowledge records", "legacy issue record awaits knowledge conversion")
		}
	}

	record := func(name string, parse func(string) error, render func() string) error {
		raw, ok, err := s.read(dir + "/" + name)
		if err != nil || !ok {
			return err
		}
		if err := parse(raw); err != nil {
			diag(domain.CodeRecordInvalid, domain.SevError, domain.ClassManual, name, "records are written by prep commands; restore it from git", "%v", err)
			return nil
		}
		if render() != raw {
			diag(domain.CodeNotCanonical, domain.SevWarning, domain.ClassFixable, name, "run prep fmt", "not in canonical format")
		}
		return nil
	}

	if bs, err := os.ReadDir(s.abs(dir + "/baselines")); err == nil {
		for _, b := range bs {
			n := b.Name()
			name := strings.TrimSuffix(n, ".md")
			if b.IsDir() || !strings.HasSuffix(n, ".md") || !domain.ValidStamp(name) {
				diag(domain.CodeRecordInvalid, domain.SevError, domain.ClassManual, "baselines/"+n, "baselines are written by prep define", "unexpected baseline entry %s", n)
				continue
			}
			var bl domain.Baseline
			if err := record("baselines/"+n, func(raw string) (err error) { bl, err = parseBaseline(name, raw); return }, func() string { return renderBaseline(&bl) }); err != nil {
				return nil, nil, err
			}
			if bl.Name != "" {
				i.Baselines = append(i.Baselines, bl)
			}
		}
		i.SortBaselines()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, err
	}
	if err := record("ready.md", func(raw string) (err error) { i.Ready, err = parseReady(raw); return }, func() string { return renderReady(i.Ready) }); err != nil {
		return nil, nil, err
	}
	if err := record("claim.md", func(raw string) (err error) { i.Claim, err = parseClaim(raw); return }, func() string { return renderClaim(i.Claim) }); err != nil {
		return nil, nil, err
	}
	if err := record("resolution.md", func(raw string) (err error) { i.Resolution, err = parseResolution(raw); return }, func() string { return renderResolution(i.Resolution) }); err != nil {
		return nil, nil, err
	}
	return i, diags, nil
}

func canonicalAcceptance(raw string) string { return fileText(normalizeChecklist(raw)) }

func canonicalProject(raw string) string {
	p, err := parseProject(raw)
	if err != nil {
		return ""
	}
	_, body, _, _ := splitFrontmatter(raw)
	return renderProject(p, body)
}

// --- config ---

type configFile struct {
	Views map[string]string `yaml:"views,omitempty"`
	Theme string            `yaml:"theme,omitempty"`
}

// Config files in .prep: each user's own config.yaml (gitignored) replaces
// the committed config.yaml.dist as a whole; the two are never merged.
const (
	ConfigOwn  = "config.yaml"
	ConfigDist = "config.yaml.dist"
)

// LoadConfig reads the configuration in effect: config.yaml, else
// config.yaml.dist; with neither, defaults.
func (s *Store) LoadConfig() (domain.Config, error) {
	unlock, err := s.lock(false)
	if err != nil {
		return domain.Config{}, err
	}
	if unlock != nil {
		defer unlock()
	}
	return s.loadConfigUnlocked()
}

func (s *Store) loadConfigUnlocked() (domain.Config, error) {
	cfg, ok, err := s.loadConfigFile(ConfigOwn)
	if ok || err != nil {
		return cfg, err
	}
	cfg, _, err = s.loadConfigFile(ConfigDist)
	return cfg, err
}

// tempPrefix starts the names of the temporary files atomic writes rename
// into place. Loading skips them: one may belong to a write in flight.
const tempPrefix = ".prep-tmp-"

// localIgnore is the .gitignore line for .prep/local, per-machine state
// such as the activity stream.
const localIgnore = activity.Dir + "/"

// localUnignored reports whether .prep/local exists without being
// ignored, so git would offer per-machine state for commit.
func (s *Store) localUnignored() bool {
	local := s.abs(Dir + "/" + activity.Dir)
	entries, err := os.ReadDir(local)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	if err != nil {
		return true
	}
	for _, entry := range entries {
		if entry.Name() != filepath.Base(lockFile) {
			return !s.ignores(localIgnore)
		}
	}
	return false
}

// ignores reports whether .prep/.gitignore lists name, with or without a
// leading slash.
func (s *Store) ignores(name string) bool {
	raw, _, _ := s.read(Dir + "/.gitignore")
	for l := range strings.SplitSeq(raw, "\n") {
		if l = strings.TrimSpace(l); l == name || l == "/"+name {
			return true
		}
	}
	return false
}

// loadConfigFile reads one config file in .prep and reports whether it exists.
func (s *Store) loadConfigFile(name string) (domain.Config, bool, error) {
	var cfg domain.Config
	raw, ok, err := s.read(Dir + "/" + name)
	if err != nil || !ok {
		return cfg, ok, err
	}
	var f configFile
	if err := decodeStrict(raw, &f); err != nil {
		return cfg, true, fmt.Errorf("%s: %v", name, err)
	}
	for view, flags := range f.Views {
		if _, err := domain.ParseFilter(strings.Fields(flags)); err != nil {
			return cfg, true, fmt.Errorf("%s: view %q: %v", name, view, err)
		}
	}
	if err := palette.Validate(f.Theme); err != nil {
		return cfg, true, fmt.Errorf("%s: %v", name, err)
	}
	cfg.Views = f.Views
	cfg.ViewOrder = viewOrder(raw)
	cfg.Theme = f.Theme
	return cfg, true, nil
}

// viewOrder returns the keys of the views mapping in file order.
func viewOrder(raw string) []string {
	var doc yaml.Node
	if yaml.Unmarshal([]byte(raw), &doc) != nil || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	for k := 0; k+1 < len(root.Content); k += 2 {
		if root.Content[k].Value != "views" {
			continue
		}
		var names []string
		v := root.Content[k+1]
		for j := 0; j+1 < len(v.Content); j += 2 {
			names = append(names, v.Content[j].Value)
		}
		return names
	}
	return nil
}

// renderConfig writes config.yaml from a configuration: the comment lines
// that lead the current file (or the default header), the views in their
// order and the theme, so the file reads like the one prep init wrote.
func renderConfig(current string, cfg domain.Config) (string, error) {
	if err := palette.Validate(cfg.Theme); err != nil {
		return "", &domain.Error{Code: domain.ErrUsage, Message: err.Error()}
	}
	var header []string
	for _, l := range strings.Split(current, "\n") {
		if !strings.HasPrefix(l, "#") {
			break
		}
		header = append(header, l)
	}
	if len(header) == 0 {
		header = strings.Split(strings.TrimSpace(defaultHeader()), "\n")
	}
	scalar := func(v string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Value: v} }
	views := &yaml.Node{Kind: yaml.MappingNode}
	for _, n := range cfg.ViewOrder {
		v := scalar(cfg.Views[n])
		if v.Value == "" {
			v.Style = yaml.DoubleQuotedStyle
		}
		views.Content = append(views.Content, scalar(n), v)
	}
	root := &yaml.Node{Kind: yaml.MappingNode}
	if len(cfg.ViewOrder) > 0 {
		root.Content = append(root.Content, scalar("views"), views)
	}
	if cfg.Theme != "" {
		root.Content = append(root.Content, scalar("theme"), scalar(cfg.Theme))
	}
	if len(root.Content) == 0 {
		return strings.Join(header, "\n") + "\n", nil
	}
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return "", err
	}
	return strings.Join(header, "\n") + "\n" + b.String(), nil
}

// defaultHeader is DefaultConfig's leading comment lines.
func defaultHeader() string {
	return DefaultConfig[:strings.Index(DefaultConfig, "\nviews:")+1]
}

// DefaultConfig is written by prep init as config.yaml.dist.
const DefaultConfig = `# prep configuration. config.yaml.dist is the project's shared default;
# copy it to config.yaml (gitignored) to make your own, which replaces it.
# views: saved queries (prep list flags), shown as tabs in prep tui.
views:
  Unresolved: --state open,defined,ready,in-progress
  All: ""
`

// DefaultProject is written by prep init: a new project with the given ID
// and name.
func DefaultProject(id, name string) string {
	return renderProject(domain.Project{Schema: domain.SchemaVersion, ID: id, Name: name}, "# Project\n\n## Definition of Done\n\n- prep check passes")
}

// Init creates the .prep skeleton for a project with the given ID and name.
func (s *Store) Init(id, name string) ([]string, error) {
	if _, err := os.Stat(s.abs(Dir + "/project.md")); err == nil {
		return nil, &domain.Error{Code: domain.ErrUsage, Message: "already initialized: .prep/project.md exists"}
	}
	release, err := s.Hold()
	if err != nil {
		return nil, err
	}
	defer release()
	var written []string
	files := []struct{ rel, content string }{
		{Dir + "/project.md", DefaultProject(id, name)},
		{Dir + "/" + ConfigDist, DefaultConfig},
		{Dir + "/.gitignore", ConfigOwn + "\n" + localIgnore + "\n"},
		{Dir + "/issues/.gitkeep", ""},
		{Dir + "/knowledge/.gitkeep", ""},
	}
	for _, f := range files {
		if _, err := os.Stat(s.abs(f.rel)); err == nil {
			continue
		}
		if err := s.write(f.rel, f.content); err != nil {
			return written, err
		}
		written = append(written, f.rel)
	}
	changed, err := ensureHistoryMerge(s)
	if err != nil {
		return written, err
	}
	return append(written, changed...), nil
}

// historyMerge makes parallel appends to a work log merge without conflict.
const historyMerge = ".prep/**/history.md merge=union"

// ensureHistoryMerge appends historyMerge to the repository's .gitattributes
// unless it is already there, and reports the file when it writes it.
func ensureHistoryMerge(s *Store) ([]string, error) {
	const rel = ".gitattributes"
	raw, _, err := s.read(rel)
	if err != nil {
		return nil, err
	}
	original := raw
	for _, rule := range []string{historyMerge, ".prep/knowledge/**/index.md merge=union"} {
		found := false
		for _, line := range strings.Split(raw, "\n") {
			if strings.TrimSpace(line) == rule {
				found = true
			}
		}
		if !found {
			if raw != "" && !strings.HasSuffix(raw, "\n") {
				raw += "\n"
			}
			raw += rule + "\n"
		}
	}
	if raw == original {
		return nil, nil
	}
	if raw != "" && !strings.HasSuffix(raw, "\n") {
		raw += "\n"
	}
	if err := s.write(rel, raw); err != nil {
		return nil, err
	}
	return []string{rel}, nil
}
