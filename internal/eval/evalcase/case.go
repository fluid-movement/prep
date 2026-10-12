// Package evalcase loads agent behavior test cases: a prompt, the fixture it
// runs in and what to expect. A case is validated completely at load time.
package evalcase

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/fixture"
)

// Case is one test.
type Case struct {
	Name        string   `yaml:"name"`
	Area        string   `yaml:"area"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Prompt      string   `yaml:"prompt"`

	// Exactly one of the four says where the agent runs. Fixture and
	// Workspace name a file under eval/fixtures; Spec and WorkspaceSpec are
	// written in the case.
	Fixture       string                 `yaml:"fixture"`
	Workspace     string                 `yaml:"workspace"`
	Spec          *fixture.Spec          `yaml:"spec"`
	WorkspaceSpec *fixture.WorkspaceSpec `yaml:"workspace_spec"`
	// Primary overrides the workspace's primary project.
	Primary string `yaml:"primary"`

	// Models and Variants restrict where the case runs; empty means any.
	Models   []string `yaml:"models"`
	Variants []string `yaml:"variants"`

	Limits Limits `yaml:"limits"`
	Expect Expect `yaml:"expect"`

	// File is where the case was loaded from.
	File string `yaml:"-"`

	resolved *fixture.WorkspaceSpec
	tools    []*yaml.Node
	outcome  []*yaml.Node
}

// Limits bound one run of the case.
type Limits struct {
	MaxTurns     int      `yaml:"max_turns"`
	MaxBudgetUSD float64  `yaml:"max_budget_usd"`
	Timeout      Duration `yaml:"timeout"`
}

// Duration is a time.Duration written as "5m".
type Duration time.Duration

func (d *Duration) UnmarshalYAML(n *yaml.Node) error {
	v, err := time.ParseDuration(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", n.Line, err)
	}
	*d = Duration(v)
	return nil
}

// String formats the duration like time.Duration.
func (d Duration) String() string { return time.Duration(d).String() }

// Default limits for a case that sets none.
var Default = Limits{MaxTurns: 30, MaxBudgetUSD: 1, Timeout: Duration(10 * time.Minute)}

// Expect is what the case expects. Tools and Outcome hold assertions, kept
// as nodes so their {{aliases}} can be expanded once the fixture is built.
type Expect struct {
	Tools   yaml.Node `yaml:"tools"`
	Outcome yaml.Node `yaml:"outcome"`
	Rubric  []string  `yaml:"rubric"`
}

var aliasRef = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_./-]+)\s*\}\}`)

// Parse reads one case from r; file names it in errors. fixtures is the
// directory fixture and workspace references resolve in ("" disallows them).
func Parse(r io.Reader, file, fixtures string) (*Case, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var c Case
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	c.File = file
	if err := c.finish(fixtures); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return &c, nil
}

func (c *Case) finish(fixtures string) error {
	if c.Name == "" {
		return fmt.Errorf("no name")
	}
	if strings.TrimSpace(c.Prompt) == "" {
		return fmt.Errorf("no prompt")
	}
	// The fixture.
	n := 0
	for _, set := range []bool{c.Fixture != "", c.Workspace != "", c.Spec != nil, c.WorkspaceSpec != nil} {
		if set {
			n++
		}
	}
	if n != 1 {
		return fmt.Errorf("exactly one of fixture, workspace, spec and workspace_spec is needed (found %d)", n)
	}
	switch {
	case c.Fixture != "" || c.Workspace != "":
		name := c.Fixture + c.Workspace
		if fixtures == "" {
			return fmt.Errorf("fixture %q referenced but there is no fixtures directory", name)
		}
		f, err := os.Open(filepath.Join(fixtures, name+".yaml"))
		if err != nil {
			return fmt.Errorf("fixture %q: %w", name, err)
		}
		defer f.Close()
		if c.Fixture != "" {
			s, err := fixture.LoadSpec(f)
			if err != nil {
				return fmt.Errorf("fixture %q: %w", name, err)
			}
			if s.Name == "" {
				s.Name = "project"
			}
			c.resolved = &fixture.WorkspaceSpec{Projects: map[string]*fixture.Spec{s.Name: s}, Primary: s.Name}
		} else if c.resolved, err = fixture.LoadWorkspace(f); err != nil {
			return fmt.Errorf("workspace %q: %w", name, err)
		}
	case c.Spec != nil:
		if c.Spec.Name == "" {
			c.Spec.Name = "project"
		}
		if err := c.Spec.Validate(); err != nil {
			return fmt.Errorf("spec: %w", err)
		}
		c.resolved = &fixture.WorkspaceSpec{Projects: map[string]*fixture.Spec{c.Spec.Name: c.Spec}, Primary: c.Spec.Name}
	default:
		if err := c.WorkspaceSpec.Validate(); err != nil {
			return fmt.Errorf("workspace_spec: %w", err)
		}
		c.resolved = c.WorkspaceSpec
	}
	if c.Primary != "" {
		if _, ok := c.resolved.Projects[c.Primary]; !ok {
			return fmt.Errorf("primary %q is not a project of the fixture", c.Primary)
		}
		c.resolved.Primary = c.Primary
	}
	// Limits.
	if c.Limits.MaxTurns == 0 {
		c.Limits.MaxTurns = Default.MaxTurns
	}
	if c.Limits.MaxBudgetUSD == 0 {
		c.Limits.MaxBudgetUSD = Default.MaxBudgetUSD
	}
	if c.Limits.Timeout == 0 {
		c.Limits.Timeout = Default.Timeout
	}
	// Assertions.
	var err error
	if c.tools, err = items(&c.Expect.Tools, "expect.tools"); err != nil {
		return err
	}
	if c.outcome, err = items(&c.Expect.Outcome, "expect.outcome"); err != nil {
		return err
	}
	identity := func(s string) (string, error) { return s, nil }
	if _, _, err := c.compileAssertions(identity); err != nil {
		return err
	}
	// Aliases used anywhere must exist in the fixture.
	known := c.Aliases()
	refs := aliasRefs(c.Prompt)
	for _, nd := range append(append([]*yaml.Node{}, c.tools...), c.outcome...) {
		refs = append(refs, nodeRefs(nd)...)
	}
	refs = append(refs, aliasRefsAll(c.Expect.Rubric)...)
	for _, r := range refs {
		if !known[r] {
			return fmt.Errorf("alias {{%s}} is not in the fixture", r)
		}
	}
	return nil
}

// items returns the assertion nodes of an expect list, or none when the key
// is absent.
func items(n *yaml.Node, where string) ([]*yaml.Node, error) {
	if n.Kind == 0 {
		return nil, nil
	}
	if n.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("line %d: %s must be a list", n.Line, where)
	}
	return n.Content, nil
}

// Aliases lists the aliases a prompt or assertion may use: plain ones from
// the primary project and project/alias from any.
func (c *Case) Aliases() map[string]bool {
	out := map[string]bool{}
	for name, s := range c.resolved.Projects {
		for _, is := range s.Issues {
			if is.Alias == "" {
				continue
			}
			out[name+"/"+is.Alias] = true
			if name == c.resolved.Primary {
				out[is.Alias] = true
			}
		}
	}
	return out
}

// Resolved returns the workspace spec the case runs in.
func (c *Case) Resolved() *fixture.WorkspaceSpec { return c.resolved }

func aliasRefs(s string) []string {
	var out []string
	for _, m := range aliasRef.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

func aliasRefsAll(ss []string) []string {
	var out []string
	for _, s := range ss {
		out = append(out, aliasRefs(s)...)
	}
	return out
}

func nodeRefs(n *yaml.Node) []string {
	var out []string
	if n.Kind == yaml.ScalarNode {
		out = append(out, aliasRefs(n.Value)...)
	}
	for _, c := range n.Content {
		out = append(out, nodeRefs(c)...)
	}
	return out
}

// Compiled is a case ready to check a run: aliases expanded, assertions
// decoded.
type Compiled struct {
	Case    *Case
	Prompt  string
	Tools   []assert.Assertion
	Outcome []assert.Assertion
	Rubric  []string
}

// Compile expands aliases with expand (the built fixture's Expand) and
// decodes the assertions.
func (c *Case) Compile(expand func(string) (string, error)) (*Compiled, error) {
	prompt, err := expand(c.Prompt)
	if err != nil {
		return nil, fmt.Errorf("%s: prompt: %w", c.File, err)
	}
	tools, outcome, err := c.compileAssertions(expand)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.File, err)
	}
	var rubric []string
	for _, r := range c.Expect.Rubric {
		e, err := expand(r)
		if err != nil {
			return nil, fmt.Errorf("%s: rubric: %w", c.File, err)
		}
		rubric = append(rubric, e)
	}
	return &Compiled{Case: c, Prompt: prompt, Tools: tools, Outcome: outcome, Rubric: rubric}, nil
}

func (c *Case) compileAssertions(expand func(string) (string, error)) (tools, outcome []assert.Assertion, err error) {
	for _, g := range []struct {
		group assert.Group
		nodes []*yaml.Node
		out   *[]assert.Assertion
	}{{assert.Tools, c.tools, &tools}, {assert.Outcome, c.outcome, &outcome}} {
		for _, n := range g.nodes {
			cp, err := expandNode(n, expand)
			if err != nil {
				return nil, nil, err
			}
			a, err := assert.Decode(g.group, cp)
			if err != nil {
				return nil, nil, err
			}
			*g.out = append(*g.out, a)
		}
	}
	return tools, outcome, nil
}

// expandNode copies a node tree, expanding aliases in its scalars.
func expandNode(n *yaml.Node, expand func(string) (string, error)) (*yaml.Node, error) {
	cp := *n
	if n.Kind == yaml.ScalarNode {
		v, err := expand(n.Value)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n.Line, err)
		}
		cp.Value = v
	}
	cp.Content = nil
	for _, c := range n.Content {
		e, err := expandNode(c, expand)
		if err != nil {
			return nil, err
		}
		cp.Content = append(cp.Content, e)
	}
	return &cp, nil
}

// Load reads every case under root/cases (any depth) and checks them
// against root/fixtures. Names must be unique.
func Load(root string) ([]*Case, error) {
	var files []string
	err := filepath.WalkDir(filepath.Join(root, "cases"), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml")) {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []*Case
	seen := map[string]string{}
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		c, err := Parse(fh, f, filepath.Join(root, "fixtures"))
		fh.Close()
		if err != nil {
			return nil, err
		}
		if c.Area == "" {
			c.Area = filepath.Base(filepath.Dir(f))
		}
		if prev, dup := seen[c.Name]; dup {
			return nil, fmt.Errorf("%s: case name %q is also used in %s", f, c.Name, prev)
		}
		seen[c.Name] = f
		out = append(out, c)
	}
	return out, nil
}

// Filter selects cases; empty fields match everything.
type Filter struct {
	Areas []string
	// Names are globs matched against the case name.
	Names []string
	// Tags select cases having any of them.
	Tags []string
}

// Select returns the cases matching f, in their order.
func Select(cases []*Case, f Filter) []*Case {
	var out []*Case
	for _, c := range cases {
		if len(f.Areas) > 0 && !contains(f.Areas, c.Area) {
			continue
		}
		if len(f.Names) > 0 {
			ok := false
			for _, g := range f.Names {
				if m, _ := path.Match(g, c.Name); m {
					ok = true
				}
			}
			if !ok {
				continue
			}
		}
		if len(f.Tags) > 0 {
			ok := false
			for _, t := range f.Tags {
				if contains(c.Tags, t) {
					ok = true
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
