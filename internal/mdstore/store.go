// Package mdstore is the markdown storage adapter: one directory per issue
// under .prep/issues, records as markdown files with YAML frontmatter.
package mdstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
	"gopkg.in/yaml.v3"
)

// Dir is the project data directory inside the repository.
const Dir = ".prep"

// Store reads and writes a .prep tree rooted at Root (the repository root).
type Store struct {
	Root   string
	hashes map[string]string // relative path -> sha256 of content at load; "" = absent
}

// Open returns a store for the repository root.
func Open(root string) *Store { return &Store{Root: root, hashes: map[string]string{}} }

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

func hashOf(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// read loads a file and records its hash for compare-and-swap.
func (s *Store) read(rel string) (string, bool, error) {
	b, err := os.ReadFile(s.abs(rel))
	if errors.Is(err, fs.ErrNotExist) {
		s.hashes[rel] = ""
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	s.hashes[rel] = hashOf(b)
	return string(b), true, nil
}

// Allowed top-level entries.
var (
	issueFiles   = map[string]bool{"issue.md": true, "acceptance.md": true, "context.md": true, "decisions.md": true, "history.md": true, "findings.md": true, "ready.md": true, "claim.md": true, "resolution.md": true}
	issueDirs    = map[string]bool{"baselines": true, "attachments": true}
	schemaFiles  = []string{"acceptance.md", "context.md", "decisions.md", "history.md"}
	projectFiles = map[string]bool{"project.md": true, "config.yaml": true, "config.yaml.dist": true, "issues": true, "knowledge": true, ".gitignore": true}
)

// Load reads the project and every issue.
func (s *Store) Load() (domain.Project, []*domain.Issue, []domain.Diagnostic, error) {
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
			diag(domain.CodeConfigInvalid, domain.SevError, domain.ClassManual, "", name, "fix "+name, "%v", err)
		}
	}
	project.Config, _ = s.LoadConfig()

	entries, err := os.ReadDir(s.abs(Dir))
	if err != nil {
		return project, nil, nil, err
	}
	for _, e := range entries {
		if !projectFiles[e.Name()] {
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
		i.Files = append(i.Files, n)
		switch {
		case e.IsDir() && issueDirs[n]:
		case !e.IsDir() && issueFiles[n]:
		default:
			diag(domain.CodeUnknownFile, domain.SevError, domain.ClassManual, n, "move free-form files into attachments/", "unknown entry %s", n)
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

	for _, n := range schemaFiles {
		if !i.HasFile(n) {
			diag(domain.CodeSchemaFileMissing, domain.SevWarning, domain.ClassFixable, n, "run prep fix to create it", "%s is missing", n)
		}
	}

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
	if raw, ok, err := plain("decisions.md", canonicalDecisions); err != nil {
		return nil, nil, err
	} else if ok {
		var problems []string
		i.Decisions, problems = parseDecisions(raw)
		for _, p := range problems {
			diag(domain.CodeDecisionInvalid, domain.SevError, domain.ClassManual, "decisions.md", "entries are '## <id>: <title>' followed by date: YYYY-MM-DD", "%s", p)
		}
	}
	if raw, ok, err := plain("history.md", fileText); err != nil {
		return nil, nil, err
	} else if ok {
		i.History = normalize(raw)
	}
	if raw, ok, err := plain("findings.md", fileText); err != nil {
		return nil, nil, err
	} else if ok {
		f := normalize(raw)
		i.Findings = &f
		dupes("findings.md", f)
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
				diag(domain.CodeRecordInvalid, domain.SevError, domain.ClassManual, "baselines/"+n, "baselines are written by prep define and prep ack", "unexpected baseline entry %s", n)
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
	return renderProject(p.Schema, body)
}

// --- config ---

type configFile struct {
	Views  map[string]string          `yaml:"views,omitempty"`
	Theme  string                     `yaml:"theme,omitempty"`
	Themes map[string]domain.ThemeDef `yaml:"themes,omitempty"`
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
	cfg, ok, err := s.loadConfigFile(ConfigOwn)
	if ok || err != nil {
		return cfg, err
	}
	cfg, _, err = s.loadConfigFile(ConfigDist)
	return cfg, err
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
	if err := palette.Validate(f.Theme, f.Themes); err != nil {
		return cfg, true, fmt.Errorf("%s: %v", name, err)
	}
	cfg.Views = f.Views
	cfg.ViewOrder = viewOrder(raw)
	cfg.Theme, cfg.Themes = f.Theme, f.Themes
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
// order, the theme, and custom themes by name with tokens in palette
// order, so the file reads like the one prep init wrote.
func renderConfig(current string, cfg domain.Config) (string, error) {
	if err := palette.Validate(cfg.Theme, cfg.Themes); err != nil {
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
		header = strings.Split(strings.TrimSpace(DefaultConfig[:strings.Index(DefaultConfig, "views:")]), "\n")
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
	if len(cfg.Themes) > 0 {
		themes := &yaml.Node{Kind: yaml.MappingNode}
		for _, name := range palette.Names(cfg.Themes) {
			def, ok := cfg.Themes[name]
			if !ok {
				continue
			}
			body := &yaml.Node{Kind: yaml.MappingNode}
			if def.Base != "" {
				body.Content = append(body.Content, scalar("base"), scalar(def.Base))
			}
			for _, bg := range []struct {
				name   string
				values map[string]string
			}{{"dark", def.Dark}, {"light", def.Light}} {
				if len(bg.values) == 0 {
					continue
				}
				toks := &yaml.Node{Kind: yaml.MappingNode}
				for _, tok := range palette.Tokens() {
					if v, ok := bg.values[tok]; ok {
						val := scalar(v)
						val.Style = yaml.DoubleQuotedStyle
						toks.Content = append(toks.Content, scalar(tok), val)
					}
				}
				body.Content = append(body.Content, scalar(bg.name), toks)
			}
			themes.Content = append(themes.Content, scalar(name), body)
		}
		root.Content = append(root.Content, scalar("themes"), themes)
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

// DefaultConfig is written by prep init as config.yaml.dist.
const DefaultConfig = `# prep configuration. config.yaml.dist is the project's shared default;
# copy it to config.yaml (gitignored) to make your own, which replaces it.
# views: saved queries (prep list flags), shown as tabs in prep tui.
views:
  Unresolved: --state open,defined,ready,in_progress
  All: ""
`

// DefaultProject is written by prep init.
func DefaultProject() string {
	return renderProject(domain.SchemaVersion, "# Project\n\n## Definition of Done\n\n- prep check passes")
}

// Init creates the .prep skeleton.
func (s *Store) Init() ([]string, error) {
	if _, err := os.Stat(s.abs(Dir + "/project.md")); err == nil {
		return nil, &domain.Error{Code: domain.ErrUsage, Message: "already initialized: .prep/project.md exists"}
	}
	var written []string
	files := []struct{ rel, content string }{
		{Dir + "/project.md", DefaultProject()},
		{Dir + "/" + ConfigDist, DefaultConfig},
		{Dir + "/.gitignore", ConfigOwn + "\n"},
		{Dir + "/issues/.gitkeep", ""},
		{Dir + "/knowledge/.gitkeep", ""},
	}
	for _, f := range files {
		if _, err := os.Stat(s.abs(f.rel)); err == nil {
			continue
		}
		s.hashes[f.rel] = ""
		if err := s.write(f.rel, f.content); err != nil {
			return written, err
		}
		written = append(written, f.rel)
	}
	return written, nil
}
