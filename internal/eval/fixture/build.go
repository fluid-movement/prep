package fixture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Options configure a build.
type Options struct {
	// Binary is the prep binary that builds the fixture (required).
	Binary string
	// Start is the clock of the first call; each call advances it a second.
	// Defaults to 2026-01-05T09:00:00Z.
	Start time.Time
	// Seed seeds the random bits of issue IDs; each project of a workspace
	// adds its position to it. Defaults to 1.
	Seed int64
	// Actor is the --by of every setup call. Defaults to "fixture".
	Actor string
}

// DefaultStart is the clock of the first call of a build.
var DefaultStart = time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)

// Project is one built fixture project.
type Project struct {
	Name string
	Dir  string
	Spec *Spec
	// IDs maps each issue alias to its ID.
	IDs map[string]string

	ws *Workspace
}

// Workspace is a set of built projects in one directory.
type Workspace struct {
	Dir string
	// Primary is the project the agent starts in.
	Primary string
	// Order is the order the projects were built in.
	Order    []string
	Projects map[string]*Project
	// State is the workspace's own registry directory (XDG_STATE_HOME).
	State string
	// Home is the HOME every setup call ran with.
	Home string

	opts  Options
	ticks int
}

// Build builds one project in dir/<name> and returns it. The directory is
// the workspace root; it must exist or be creatable.
func Build(spec *Spec, dir string, opts Options) (*Project, error) {
	if spec.Name == "" {
		spec.Name = "project"
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	ws := &WorkspaceSpec{Projects: map[string]*Spec{spec.Name: spec}, Primary: spec.Name}
	w, err := BuildWorkspace(ws, dir, opts)
	if err != nil {
		return nil, err
	}
	return w.Projects[spec.Name], nil
}

// BuildWorkspace builds every project of ws in dir/<name>, linked as the
// specs say, with a registry of its own so cross-project commands work.
func BuildWorkspace(ws *WorkspaceSpec, dir string, opts Options) (*Workspace, error) {
	if opts.Binary == "" {
		return nil, fmt.Errorf("fixture: no prep binary given")
	}
	if opts.Start.IsZero() {
		opts.Start = DefaultStart
	}
	if opts.Seed == 0 {
		opts.Seed = 1
	}
	if opts.Actor == "" {
		opts.Actor = "fixture"
	}
	if err := ws.Validate(); err != nil {
		return nil, err
	}
	order, err := ws.order()
	if err != nil {
		return nil, err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	w := &Workspace{
		Dir:      dir,
		Primary:  ws.primary(),
		Order:    order,
		Projects: map[string]*Project{},
		State:    filepath.Join(dir, ".state"),
		Home:     filepath.Join(dir, ".home"),
		opts:     opts,
	}
	for _, d := range []string{w.State, w.Home} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	for i, name := range order {
		p := &Project{Name: name, Dir: filepath.Join(dir, name), Spec: ws.Projects[name], IDs: map[string]string{}, ws: w}
		w.Projects[name] = p
		if err := p.build(opts.Seed + int64(i)); err != nil {
			return w, fmt.Errorf("project %s: %w", name, err)
		}
	}
	return w, nil
}

// Primary returns the primary project.
func (w *Workspace) PrimaryProject() *Project { return w.Projects[w.Primary] }

var aliasRef = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_./-]+)\s*\}\}`)

// Expand replaces {{alias}} with the ID of an issue of this project and
// {{project/alias}} with the ID of an issue of another; an unknown alias is
// an error.
func (p *Project) Expand(s string) (string, error) {
	var err error
	out := aliasRef.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.TrimSpace(m[2 : len(m)-2])
		id, e := p.lookup(key)
		if e != nil && err == nil {
			err = e
		}
		return id
	})
	return out, err
}

func (p *Project) lookup(key string) (string, error) {
	if i := strings.Index(key, "/"); i >= 0 {
		o, ok := p.ws.Projects[key[:i]]
		if !ok {
			return "", fmt.Errorf("unknown project in alias %q", key)
		}
		id, ok := o.IDs[key[i+1:]]
		if !ok {
			return "", fmt.Errorf("unknown alias %q", key)
		}
		return id, nil
	}
	id, ok := p.IDs[key]
	if !ok {
		return "", fmt.Errorf("unknown alias %q", key)
	}
	return id, nil
}

// Prep runs the binary in the project the way the builder does and returns
// its output; it advances the clock like any setup call.
func (p *Project) Prep(args ...string) (string, error) {
	return p.ws.prep(p.Dir, p.seed(), args...)
}

// Query runs a read command in the project without a --by, so the call is
// not recorded in the activity stream and the agent never sees it.
func (p *Project) Query(args ...string) (string, error) {
	return p.ws.query(p.Dir, p.seed(), args...)
}

func (w *Workspace) query(dir string, seed int64, args ...string) (string, error) {
	cmd := exec.Command(w.opts.Binary, append([]string{"--root", dir}, args...)...)
	cmd.Dir = dir
	cmd.Env = w.env(seed)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("prep %s: %v\n%s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()+out.String()))
	}
	return out.String(), nil
}

// Git runs git in the project and returns its output.
func (p *Project) Git(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", p.Dir}, args...)...)
	cmd.Env = p.ws.env(0)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// HasRepo reports whether the project is a git repository.
func (p *Project) HasRepo() bool {
	_, err := os.Stat(filepath.Join(p.Dir, ".git"))
	return err == nil
}

func (p *Project) seed() int64 {
	for i, n := range p.ws.Order {
		if n == p.Name {
			return p.ws.opts.Seed + int64(i)
		}
	}
	return p.ws.opts.Seed
}

func (w *Workspace) env(seed int64) []string {
	t := w.opts.Start.Add(time.Duration(w.ticks) * time.Second)
	w.ticks++
	return []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + w.Home,
		"XDG_STATE_HOME=" + w.State,
		"XDG_CONFIG_HOME=" + filepath.Join(w.Home, ".config"),
		"TMPDIR=" + os.TempDir(),
		"LANG=C",
		"PREP_TEST_NOW=" + t.Format(time.RFC3339),
		"PREP_TEST_SEED=" + strconv.FormatInt(seed, 10),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
		"GIT_AUTHOR_DATE=" + t.Format(time.RFC3339), "GIT_COMMITTER_DATE=" + t.Format(time.RFC3339),
	}
}

func (w *Workspace) prep(dir string, seed int64, args ...string) (string, error) {
	full := append([]string{"--root", dir, "--by", w.opts.Actor}, args...)
	for _, a := range args {
		if a == "--project" { // runs from the project's directory instead
			full = append([]string{"--by", w.opts.Actor}, args...)
			break
		}
	}
	cmd := exec.Command(w.opts.Binary, full...)
	cmd.Dir = dir
	cmd.Env = w.env(seed)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("prep %s: %v\n%s", strings.Join(args, " "), err, strings.TrimSpace(out.String()))
	}
	return out.String(), nil
}

func (w *Workspace) git(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	cmd.Env = w.env(0)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func (p *Project) build(seed int64) error {
	w := p.ws
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}
	s := p.Spec
	hasRepo := s.Repo != nil
	if hasRepo {
		for rel, content := range s.Repo.Files {
			path := filepath.Join(p.Dir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return err
			}
		}
		if err := w.git(p.Dir, "init", "-q"); err != nil {
			return err
		}
		if s.Repo.Commit == nil || *s.Repo.Commit {
			if err := w.git(p.Dir, "add", "-A"); err != nil {
				return err
			}
			if err := w.git(p.Dir, "commit", "-q", "--allow-empty", "-m", "Initial commit"); err != nil {
				return err
			}
		}
	}
	run := func(args ...string) (string, error) { return w.prep(p.Dir, seed, args...) }
	if _, err := run("init", "--no-bootstrap"); err != nil {
		return err
	}
	if _, err := w.prep(p.Dir, seed, "projects", "add", p.Dir); err != nil {
		return err
	}
	if s.Config != nil {
		data, err := yaml.Marshal(s.Config)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(p.Dir, ".prep", "config.yaml"), data, 0o644); err != nil {
			return err
		}
	}
	for _, l := range s.Links {
		if _, err := run("link", "add", l); err != nil {
			return err
		}
	}
	if err := p.buildIssues(run); err != nil {
		return err
	}
	if err := p.buildKnowledge(run); err != nil {
		return err
	}
	if hasRepo && (s.Repo.Commit == nil || *s.Repo.Commit) {
		if err := w.git(p.Dir, "add", "-A"); err != nil {
			return err
		}
		if err := w.git(p.Dir, "commit", "-q", "--allow-empty", "-m", "Project records"); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) buildKnowledge(run func(...string) (string, error)) error {
	entries := p.Spec.Knowledge
	hasOverview := false
	for _, k := range entries {
		if k.Path == "/overview.md" {
			hasOverview = true
		}
	}
	if !hasOverview {
		entries = append([]KnowledgeSpec{{Path: "/overview.md", Type: "overview", Title: "Project overview", Description: "Entry point to the knowledge base.", Body: "The fixture project " + p.Name + "."}}, entries...)
	} else {
		// the overview must come first
		for i, k := range entries {
			if k.Path == "/overview.md" {
				entries = append([]KnowledgeSpec{k}, append(append([]KnowledgeSpec{}, entries[:i]...), entries[i+1:]...)...)
				break
			}
		}
	}
	for _, k := range entries {
		if k.Path == "" || k.Title == "" {
			return fmt.Errorf("knowledge entry needs a path and a title")
		}
		typ, desc := k.Type, k.Description
		if typ == "" {
			typ = "component"
		}
		if desc == "" {
			desc = k.Title + "."
		}
		body, err := p.Expand(k.Body)
		if err != nil {
			return err
		}
		if body == "" {
			body = k.Title + "."
		}
		if _, err := run("knowledge", "new", k.Path, "--type", typ, "--title", k.Title, "--description", desc, "--body", body); err != nil {
			return err
		}
	}
	return nil
}

type writeOut struct {
	ID string `json:"id"`
}

func (p *Project) buildIssues(run func(...string) (string, error)) error {
	issues := p.Spec.Issues
	idx := map[string]int{}
	for i, is := range issues {
		if is.Alias != "" {
			idx[is.Alias] = i
		}
	}
	children := map[string][]int{}
	for i, is := range issues {
		if is.Parent != "" {
			children[is.Parent] = append(children[is.Parent], i)
		}
	}
	stateOf := func(i int) string {
		if issues[i].State == "" {
			return "open"
		}
		return issues[i].State
	}
	for i, is := range issues {
		if len(children[is.Alias]) > 0 && is.Alias != "" && stateOf(i) == "in-progress" {
			return fmt.Errorf("issue %q is a parent and cannot be in-progress", is.Title)
		}
	}

	// Creation order: parents and local dependencies first.
	create, err := topo(len(issues), func(i int) []int {
		var r []int
		if issues[i].Parent != "" {
			r = append(r, idx[issues[i].Parent])
		}
		for _, d := range issues[i].DependsOn {
			if !strings.Contains(d, "/") {
				r = append(r, idx[d])
			}
		}
		return r
	})
	if err != nil {
		return fmt.Errorf("issues: %w", err)
	}
	ids := make([]string, len(issues))
	for _, i := range create {
		is := issues[i]
		kind := is.Kind
		if kind == "" {
			kind = "code"
		}
		args := []string{"new", "--json", "--title", is.Title, "--kind", kind}
		if is.Parent != "" {
			args = append(args, "--parent", ids[idx[is.Parent]])
		}
		for _, d := range is.DependsOn {
			if j := strings.Index(d, "/"); j >= 0 {
				id, err := p.lookup(d)
				if err != nil {
					return err
				}
				args = append(args, "--depends-on", d[:j]+":"+id)
			} else {
				args = append(args, "--depends-on", ids[idx[d]])
			}
		}
		for _, t := range is.Tags {
			args = append(args, "--tag", t)
		}
		if is.Priority != "" {
			args = append(args, "--priority", is.Priority)
		}
		body := is.Body
		if body == "" || strings.Contains(body, "{{") {
			body = "Fixture requirement: " + is.Title + "."
		}
		args = append(args, "--body", body)
		out, err := run(args...)
		if err != nil {
			return err
		}
		var r writeOut
		if err := json.Unmarshal([]byte(out), &r); err != nil || r.ID == "" {
			return fmt.Errorf("issue %q: no ID in %q", is.Title, out)
		}
		ids[i] = r.ID
		if is.Alias != "" {
			p.IDs[is.Alias] = r.ID
		}
	}

	// Bodies, contexts and criteria with aliases in them, now every ID is known.
	for i, is := range issues {
		if strings.Contains(is.Body, "{{") {
			body, err := p.Expand(is.Body)
			if err != nil {
				return fmt.Errorf("issue %q: %w", is.Title, err)
			}
			if _, err := run("edit", ids[i], "--body", body); err != nil {
				return err
			}
		}
	}

	// Transition order: dependencies first, children before their parent.
	trans, err := topo(len(issues), func(i int) []int {
		var r []int
		for _, d := range issues[i].DependsOn {
			if !strings.Contains(d, "/") {
				r = append(r, idx[d])
			}
		}
		if issues[i].Alias != "" {
			r = append(r, children[issues[i].Alias]...)
		}
		return r
	})
	if err != nil {
		return fmt.Errorf("issues: %w", err)
	}
	for _, i := range trans {
		if err := p.reach(run, issues[i], ids[i], stateOf(i)); err != nil {
			return fmt.Errorf("issue %q: %w", issues[i].Title, err)
		}
	}
	return nil
}

var stateRank = map[string]int{"open": 0, "defined": 1, "ready": 2, "in-progress": 3, "done": 4, "dropped": 5}

func (p *Project) reach(run func(...string) (string, error), is IssueSpec, id, state string) error {
	rank := stateRank[state]
	kind := is.Kind
	if kind == "" {
		kind = "code"
	}
	crit := is.Criteria
	var err error
	for i := range crit {
		if crit[i], err = p.Expand(crit[i]); err != nil {
			return err
		}
	}
	if rank >= 2 && rank != 5 {
		if len(crit) == 0 {
			crit = []string{"Done as described"}
		}
	}
	for _, c := range crit {
		if _, err := run("criterion", id, "--add", c); err != nil {
			return err
		}
	}
	ctx := is.Context
	if ctx != "" {
		if ctx, err = p.Expand(ctx); err != nil {
			return err
		}
	}
	if ctx == "" && kind != "manual" && rank >= 2 && rank != 5 {
		ctx = "Fixture context for " + is.Title + "."
	}
	if ctx != "" {
		if _, err := run("context", id, "--body", ctx); err != nil {
			return err
		}
	}
	if state == "dropped" {
		_, err := run("drop", id, "--reason", "fixture")
		return err
	}
	if rank >= 1 {
		if _, err := run("define", id); err != nil {
			return err
		}
	}
	if rank >= 2 {
		if _, err := run("ready", id); err != nil {
			return err
		}
	}
	if state == "in-progress" || state == "done" {
		isParent := false
		for _, o := range p.Spec.Issues {
			if o.Parent != "" && o.Parent == is.Alias && is.Alias != "" {
				isParent = true
			}
		}
		if !isParent {
			if _, err := run("claim", id); err != nil {
				return err
			}
		}
	}
	if state == "done" {
		for n := 1; n <= len(crit); n++ {
			if _, err := run("criterion", id, "--check", strconv.Itoa(n)); err != nil {
				return err
			}
		}
		switch kind {
		case "research":
			if _, err := run("findings", id, "--body", "Fixture findings."); err != nil {
				return err
			}
		case "decision":
			if _, err := run("decide", id, "--title", "Fixture decision", "--body", "Decided for the fixture.", "--outcome"); err != nil {
				return err
			}
		}
		if _, err := run("complete", id, "--no-impact", "fixture"); err != nil {
			return err
		}
	}
	return nil
}

// topo orders 0..n-1 so every node follows the ones deps names, keeping the
// given order otherwise; a cycle is an error.
func topo(n int, deps func(int) []int) ([]int, error) {
	placed := make([]bool, n)
	var out []int
	for len(out) < n {
		progress := false
		for i := 0; i < n; i++ {
			if placed[i] {
				continue
			}
			ok := true
			for _, d := range deps(i) {
				if !placed[d] {
					ok = false
					break
				}
			}
			if ok {
				placed[i] = true
				out = append(out, i)
				progress = true
			}
		}
		if !progress {
			return nil, fmt.Errorf("parents and dependencies form a cycle")
		}
	}
	return out, nil
}
