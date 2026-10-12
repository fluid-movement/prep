// Package fixture builds fake prep projects from a declarative spec, for
// the agent behavior tests. It drives the real prep binary for everything
// the CLI can do, so a fixture is valid by construction and reaches every
// state through the real transitions.
package fixture

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec describes one fixture project.
type Spec struct {
	// Name is the project's name; it is the directory name too. Defaults to
	// "project" (or the key in a workspace spec).
	Name      string          `yaml:"name"`
	Issues    []IssueSpec     `yaml:"issues"`
	Knowledge []KnowledgeSpec `yaml:"knowledge"`
	// Links lists the workspace projects this one links to.
	Links []string `yaml:"links"`
	// Config is written as .prep/config.yaml (views, theme); the CLI has no
	// command for it.
	Config map[string]any `yaml:"config"`
	// Repo makes the project a git repository with stub files.
	Repo *RepoSpec `yaml:"repo"`
	// Bootstrap runs prep init the way a user does, creating the knowledge
	// base bootstrap issues, and leaves the knowledge base empty (no
	// overview): the starting point of "set up prep in this project".
	Bootstrap bool `yaml:"bootstrap"`
	// Generate adds synthetic issues and knowledge entries, to make a project
	// as big as a real one.
	Generate *GenSpec `yaml:"generate"`
}

// GenSpec describes synthetic bulk: deterministic, varied in state, kind,
// priority, parent and dependencies, and free of any alias.
type GenSpec struct {
	Issues    int `yaml:"issues"`
	Knowledge int `yaml:"knowledge"`
}

// IssueSpec describes one issue. Strings may use {{alias}} (or
// {{project/alias}}) for the IDs of other fixture issues.
type IssueSpec struct {
	Alias     string   `yaml:"alias"`
	Title     string   `yaml:"title"`
	Kind      string   `yaml:"kind"`  // default code
	State     string   `yaml:"state"` // open (default), defined, ready, in-progress, done, dropped
	Parent    string   `yaml:"parent"`
	DependsOn []string `yaml:"depends_on"` // alias, or project/alias in a workspace
	Tags      []string `yaml:"tags"`
	Priority  string   `yaml:"priority"`
	Body      string   `yaml:"body"`
	Context   string   `yaml:"context"`
	Criteria  []string `yaml:"criteria"`
}

// KnowledgeSpec describes one knowledge entry.
type KnowledgeSpec struct {
	Path        string `yaml:"path"` // /components/x.md
	Type        string `yaml:"type"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Body        string `yaml:"body"`
	// Scope lists the paths or globs the entry applies to, which prep uses to
	// notice drift when they change.
	Scope []string `yaml:"scope"`
}

// RepoSpec describes the optional git repository around the project.
type RepoSpec struct {
	// Files maps relative paths to contents.
	Files map[string]string `yaml:"files"`
	// Executable lists files that get the executable bit.
	Executable []string `yaml:"executable"`
	// Commit makes the initial commit (default true).
	Commit *bool `yaml:"commit"`
}

// WorkspaceSpec describes several projects that know each other.
type WorkspaceSpec struct {
	Projects map[string]*Spec `yaml:"projects"`
	// Primary is the project the agent starts in. Defaults to the only one,
	// else the first name in sorted order.
	Primary string `yaml:"primary"`
}

// LoadSpec reads a single-project spec; unknown keys are errors.
func LoadSpec(r io.Reader) (*Spec, error) {
	var s Spec
	if err := decodeStrict(r, &s); err != nil {
		return nil, err
	}
	return &s, s.Validate()
}

// LoadWorkspace reads a workspace spec; unknown keys are errors.
func LoadWorkspace(r io.Reader) (*WorkspaceSpec, error) {
	var w WorkspaceSpec
	if err := decodeStrict(r, &w); err != nil {
		return nil, err
	}
	return &w, w.Validate()
}

func decodeStrict(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil && err != io.EOF {
		return err
	}
	return nil
}

var validStates = map[string]bool{"open": true, "defined": true, "ready": true, "in-progress": true, "done": true, "dropped": true}

// Validate checks a spec on its own, without building it.
func (s *Spec) Validate() error {
	seen := map[string]bool{}
	for i := range s.Issues {
		is := &s.Issues[i]
		if is.Title == "" {
			return fmt.Errorf("issue %d has no title", i+1)
		}
		if is.Alias != "" {
			if seen[is.Alias] {
				return fmt.Errorf("alias %q is used twice", is.Alias)
			}
			seen[is.Alias] = true
		}
		if is.State != "" && !validStates[is.State] {
			return fmt.Errorf("issue %q: unknown state %q", is.Title, is.State)
		}
	}
	for _, is := range s.Issues {
		if is.Parent != "" && !seen[is.Parent] {
			return fmt.Errorf("issue %q: parent %q is not an alias in this project", is.Title, is.Parent)
		}
		for _, d := range is.DependsOn {
			if !strings.Contains(d, "/") && !seen[d] {
				return fmt.Errorf("issue %q: depends_on %q is not an alias in this project", is.Title, d)
			}
		}
	}
	return nil
}

// Validate checks every project and the links between them.
func (w *WorkspaceSpec) Validate() error {
	if len(w.Projects) == 0 {
		return fmt.Errorf("workspace has no projects")
	}
	for name, s := range w.Projects {
		if s == nil {
			w.Projects[name] = &Spec{}
			s = w.Projects[name]
		}
		if s.Name != "" && s.Name != name {
			return fmt.Errorf("project %q: name %q differs from its key", name, s.Name)
		}
		s.Name = name
		if err := s.Validate(); err != nil {
			return fmt.Errorf("project %s: %w", name, err)
		}
		for _, l := range s.Links {
			if _, ok := w.Projects[l]; !ok {
				return fmt.Errorf("project %s links to %q, which is not in the workspace", name, l)
			}
		}
		for _, is := range s.Issues {
			for _, d := range is.DependsOn {
				if i := strings.Index(d, "/"); i >= 0 {
					p, a := d[:i], d[i+1:]
					t, ok := w.Projects[p]
					if !ok {
						return fmt.Errorf("project %s: %q depends on project %q, which is not in the workspace", name, is.Title, p)
					}
					if !t.hasAlias(a) {
						return fmt.Errorf("project %s: %q depends on %s/%s, which has no such alias", name, is.Title, p, a)
					}
				}
			}
		}
	}
	if w.Primary != "" {
		if _, ok := w.Projects[w.Primary]; !ok {
			return fmt.Errorf("primary %q is not in the workspace", w.Primary)
		}
	}
	if _, err := w.order(); err != nil {
		return err
	}
	return nil
}

func (s *Spec) hasAlias(a string) bool {
	for _, is := range s.Issues {
		if is.Alias == a {
			return true
		}
	}
	return false
}

// primary returns the name of the primary project.
func (w *WorkspaceSpec) primary() string {
	if w.Primary != "" {
		return w.Primary
	}
	names := w.names()
	return names[0]
}

func (w *WorkspaceSpec) names() []string {
	var n []string
	for k := range w.Projects {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

// order returns the project names so every project comes after the ones it
// links to or depends on; a cycle is an error.
func (w *WorkspaceSpec) order() ([]string, error) {
	deps := map[string]map[string]bool{}
	for name, s := range w.Projects {
		deps[name] = map[string]bool{}
		for _, l := range s.Links {
			deps[name][l] = true
		}
		for _, is := range s.Issues {
			for _, d := range is.DependsOn {
				if i := strings.Index(d, "/"); i >= 0 {
					deps[name][d[:i]] = true
				}
			}
		}
	}
	var out []string
	state := map[string]int{}
	var visit func(n string, path []string) error
	visit = func(n string, path []string) error {
		switch state[n] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("projects form a cycle: %s", strings.Join(append(path, n), " -> "))
		}
		state[n] = 1
		var ds []string
		for d := range deps[n] {
			ds = append(ds, d)
		}
		sort.Strings(ds)
		for _, d := range ds {
			if d == n {
				continue
			}
			if err := visit(d, append(path, n)); err != nil {
				return err
			}
		}
		state[n] = 2
		out = append(out, n)
		return nil
	}
	for _, n := range w.names() {
		if err := visit(n, nil); err != nil {
			return nil, err
		}
	}
	return out, nil
}
