package assert

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/fixture"
)

// Snapshot is the state of a workspace at one moment, read through prep's
// --json output, never from the files.
type Snapshot struct {
	Projects map[string]*ProjectSnapshot `json:"projects"`
}

// ProjectSnapshot is one project's state.
type ProjectSnapshot struct {
	Issues    map[string]*Issue `json:"issues"`
	Knowledge map[string]*Entry `json:"knowledge"`
	Check     Check             `json:"check"`
	Git       *GitState         `json:"git,omitempty"`
}

// Issue is what prep show reports about one issue.
type Issue struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	State       string   `json:"state"`
	Stale       bool     `json:"stale,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Requirement string   `json:"requirement,omitempty"`
	// OpenQuestions is the requirement's ## Open questions section, which prep
	// reports apart from the rest of the requirement.
	OpenQuestions string      `json:"open_questions,omitempty"`
	Context       string      `json:"context,omitempty"`
	Criteria      []Criterion `json:"criteria,omitempty"`
	ClaimedBy     string      `json:"claimed_by,omitempty"`
	History       string      `json:"history,omitempty"`
}

// Criterion is an acceptance criterion.
type Criterion struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// Entry is a knowledge entry.
type Entry struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Text        string `json:"text"`
}

// Check is the result of prep check.
type Check struct {
	Errors      int          `json:"errors"`
	Warnings    int          `json:"warnings"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// Diagnostic is one finding of prep check.
type Diagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Message  string `json:"message"`
}

// GitState is the repository's state.
type GitState struct {
	Head string `json:"head"`
	// Porcelain lists `git status --porcelain` lines.
	Porcelain []string `json:"porcelain,omitempty"`
}

// Capture reads the state of every project of the workspace.
func Capture(w *fixture.Workspace) (*Snapshot, error) {
	s := &Snapshot{Projects: map[string]*ProjectSnapshot{}}
	for _, name := range w.Order {
		ps, err := capture(w.Projects[name])
		if err != nil {
			return nil, fmt.Errorf("snapshot of %s: %w", name, err)
		}
		s.Projects[name] = ps
	}
	return s, nil
}

func capture(p *fixture.Project) (*ProjectSnapshot, error) {
	ps := &ProjectSnapshot{Issues: map[string]*Issue{}, Knowledge: map[string]*Entry{}}
	out, err := p.Query("list", "--json")
	if err != nil {
		return nil, err
	}
	var list struct {
		Issues []struct {
			ID string `json:"id"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	for _, li := range list.Issues {
		out, err := p.Query("show", li.ID, "--json")
		if err != nil {
			return nil, err
		}
		is, err := parseShow(out)
		if err != nil {
			return nil, fmt.Errorf("show %s: %w", li.ID, err)
		}
		ps.Issues[is.ID] = is
	}
	out, err = p.Query("knowledge", "list", "--json")
	if err != nil {
		return nil, err
	}
	var kl struct {
		Entries []struct {
			Path string `json:"path"`
		} `json:"entries"`
	}
	if err := json.Unmarshal([]byte(out), &kl); err != nil {
		return nil, fmt.Errorf("knowledge list: %w", err)
	}
	for _, ke := range kl.Entries {
		out, err := p.Query("knowledge", "show", ke.Path, "--json")
		if err != nil {
			return nil, err
		}
		var ks struct {
			Entries []Entry `json:"entries"`
		}
		if err := json.Unmarshal([]byte(out), &ks); err != nil || len(ks.Entries) == 0 {
			return nil, fmt.Errorf("knowledge show %s: %v", ke.Path, err)
		}
		e := ks.Entries[0]
		ps.Knowledge[e.Path] = &e
	}
	// check exits non-zero on errors; its output is still the report.
	out, _ = p.Query("check", "--json")
	var ck struct {
		Errors      int          `json:"errors"`
		Warnings    int          `json:"warnings"`
		Diagnostics []Diagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &ck); err != nil {
		return nil, fmt.Errorf("check: %w\n%s", err, out)
	}
	ps.Check = Check{Errors: ck.Errors, Warnings: ck.Warnings, Diagnostics: ck.Diagnostics}
	if p.HasRepo() {
		g := &GitState{}
		if h, err := p.Git("rev-parse", "HEAD"); err == nil {
			g.Head = strings.TrimSpace(h)
		}
		if st, err := p.Git("status", "--porcelain"); err == nil {
			for _, l := range strings.Split(strings.TrimRight(st, "\n"), "\n") {
				if l != "" {
					g.Porcelain = append(g.Porcelain, l)
				}
			}
		}
		ps.Git = g
	}
	return ps, nil
}

func parseShow(out string) (*Issue, error) {
	var d struct {
		State    string `json:"state"`
		Stale    bool   `json:"stale"`
		Priority string `json:"priority"`
		Context  string `json:"context"`
		History  string `json:"history"`
		Issue    struct {
			ID          string      `json:"id"`
			Title       string      `json:"title"`
			Kind        string      `json:"kind"`
			Parent      string      `json:"parent"`
			DependsOn   []string    `json:"depends_on"`
			Tags        []string    `json:"tags"`
			Requirement string      `json:"requirement"`
			OpenQ       string      `json:"open_questions"`
			Criteria    []Criterion `json:"criteria"`
			Claim       *struct {
				By string `json:"by"`
			} `json:"claim"`
		} `json:"issue"`
	}
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	i := d.Issue
	is := &Issue{ID: i.ID, Title: i.Title, Kind: i.Kind, State: d.State, Stale: d.Stale, Parent: i.Parent,
		DependsOn: i.DependsOn, Tags: i.Tags, Priority: d.Priority, Requirement: i.Requirement, OpenQuestions: i.OpenQ,
		Context: d.Context, Criteria: i.Criteria, History: d.History}
	if i.Claim != nil {
		is.ClaimedBy = i.Claim.By
	}
	if is.ID == "" {
		return nil, fmt.Errorf("no issue in the output")
	}
	return is, nil
}

// Diff is what changed between two snapshots.
type Diff struct {
	Projects map[string]*ProjectDiff `json:"projects"`
}

// ProjectDiff is what changed in one project.
type ProjectDiff struct {
	Created []string `json:"created,omitempty"`
	Removed []string `json:"removed,omitempty"`
	// Changed maps an issue ID to the fields that differ.
	Changed map[string][]string `json:"changed,omitempty"`
	// Knowledge changes by entry path.
	EntriesCreated []string `json:"entries_created,omitempty"`
	EntriesChanged []string `json:"entries_changed,omitempty"`
	EntriesRemoved []string `json:"entries_removed,omitempty"`
	// HeadMoved is true when the repository has new commits; Files are the
	// paths changed in them or left modified.
	HeadMoved bool     `json:"head_moved,omitempty"`
	Files     []string `json:"files,omitempty"`
}

// Empty reports whether nothing changed.
func (d *ProjectDiff) Empty() bool {
	return len(d.Created) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0 &&
		len(d.EntriesCreated) == 0 && len(d.EntriesChanged) == 0 && len(d.EntriesRemoved) == 0 && !d.HeadMoved && len(d.Files) == 0
}

// Compare computes the diff from before to after. An issue's history
// counts as a field: a work log line or any other write shows up as a change
// even when no other field differs.
func Compare(w *fixture.Workspace, before, after *Snapshot) *Diff {
	d := &Diff{Projects: map[string]*ProjectDiff{}}
	for name, a := range after.Projects {
		b := before.Projects[name]
		if b == nil {
			b = &ProjectSnapshot{Issues: map[string]*Issue{}, Knowledge: map[string]*Entry{}}
		}
		pd := &ProjectDiff{Changed: map[string][]string{}}
		for id, ai := range a.Issues {
			bi, ok := b.Issues[id]
			if !ok {
				pd.Created = append(pd.Created, id)
				continue
			}
			if f := issueFields(bi, ai); len(f) > 0 {
				pd.Changed[id] = f
			}
		}
		for id := range b.Issues {
			if _, ok := a.Issues[id]; !ok {
				pd.Removed = append(pd.Removed, id)
			}
		}
		for path, ae := range a.Knowledge {
			be, ok := b.Knowledge[path]
			switch {
			case !ok:
				pd.EntriesCreated = append(pd.EntriesCreated, path)
			case hash(be) != hash(ae):
				pd.EntriesChanged = append(pd.EntriesChanged, path)
			}
		}
		for path := range b.Knowledge {
			if _, ok := a.Knowledge[path]; !ok {
				pd.EntriesRemoved = append(pd.EntriesRemoved, path)
			}
		}
		if a.Git != nil && b.Git != nil {
			pd.HeadMoved = a.Git.Head != b.Git.Head
			files := map[string]bool{}
			if pd.HeadMoved {
				if p := w.Projects[name]; p != nil {
					if out, err := p.Git("diff", "--name-only", b.Git.Head, a.Git.Head); err == nil {
						for _, f := range strings.Fields(out) {
							files[f] = true
						}
					}
				}
			}
			for _, l := range a.Git.Porcelain {
				if len(l) > 3 {
					files[strings.TrimSpace(l[3:])] = true
				}
			}
			for _, l := range b.Git.Porcelain { // already dirty before the agent: not its change
				if len(l) > 3 && a.Git.contains(l) {
					delete(files, strings.TrimSpace(l[3:]))
				}
			}
			for f := range files {
				pd.Files = append(pd.Files, f)
			}
		}
		sort.Strings(pd.Created)
		sort.Strings(pd.Removed)
		sort.Strings(pd.EntriesCreated)
		sort.Strings(pd.EntriesChanged)
		sort.Strings(pd.EntriesRemoved)
		sort.Strings(pd.Files)
		d.Projects[name] = pd
	}
	return d
}

func (g *GitState) contains(line string) bool {
	for _, l := range g.Porcelain {
		if l == line {
			return true
		}
	}
	return false
}

func hash(v any) string {
	b, _ := json.Marshal(v)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func issueFields(b, a *Issue) []string {
	var f []string
	add := func(name string, differs bool) {
		if differs {
			f = append(f, name)
		}
	}
	add("title", b.Title != a.Title)
	add("kind", b.Kind != a.Kind)
	add("state", b.State != a.State)
	add("parent", b.Parent != a.Parent)
	add("depends_on", strings.Join(b.DependsOn, ",") != strings.Join(a.DependsOn, ","))
	add("tags", strings.Join(b.Tags, ",") != strings.Join(a.Tags, ","))
	add("priority", b.Priority != a.Priority)
	add("requirement", b.Requirement != a.Requirement || b.OpenQuestions != a.OpenQuestions)
	add("context", b.Context != a.Context)
	add("criteria", hash(b.Criteria) != hash(a.Criteria))
	add("claim", b.ClaimedBy != a.ClaimedBy)
	add("history", b.History != a.History)
	return f
}

// Summary describes the diff in words, for reports and the judge.
func (d *Diff) Summary(after *Snapshot) string {
	var names []string
	for n := range d.Projects {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		pd := d.Projects[n]
		if pd.Empty() {
			continue
		}
		fmt.Fprintf(&b, "project %s:\n", n)
		ps := after.Projects[n]
		for _, id := range pd.Created {
			is := ps.Issues[id]
			fmt.Fprintf(&b, "  created %s %q (%s, %s)\n", id, is.Title, is.Kind, is.State)
		}
		ids := make([]string, 0, len(pd.Changed))
		for id := range pd.Changed {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintf(&b, "  changed %s %q: %s\n", id, ps.Issues[id].Title, strings.Join(pd.Changed[id], ", "))
		}
		for _, id := range pd.Removed {
			fmt.Fprintf(&b, "  removed %s\n", id)
		}
		for _, p := range pd.EntriesCreated {
			fmt.Fprintf(&b, "  created knowledge entry %s\n", p)
		}
		for _, p := range pd.EntriesChanged {
			fmt.Fprintf(&b, "  changed knowledge entry %s\n", p)
		}
		for _, p := range pd.EntriesRemoved {
			fmt.Fprintf(&b, "  removed knowledge entry %s\n", p)
		}
		if pd.HeadMoved {
			fmt.Fprintf(&b, "  new commits\n")
		}
		for _, f := range pd.Files {
			fmt.Fprintf(&b, "  file changed: %s\n", f)
		}
	}
	if b.Len() == 0 {
		return "nothing changed\n"
	}
	return b.String()
}

// Detail writes out the text of everything the run created or changed — the
// requirement, open questions, context and criteria of issues, the text of
// knowledge entries — each clipped to max characters, so a judge reads what
// the agent wrote and not only that it wrote something.
func (d *Diff) Detail(after *Snapshot, max int) string {
	clip := func(s string) string {
		s = strings.TrimSpace(s)
		if len(s) > max {
			return s[:max] + "…"
		}
		return s
	}
	var names []string
	for n := range d.Projects {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		pd, ps := d.Projects[n], after.Projects[n]
		if pd == nil || ps == nil {
			continue
		}
		ids := append([]string{}, pd.Created...)
		for id := range pd.Changed {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			is := ps.Issues[id]
			if is == nil {
				continue
			}
			fmt.Fprintf(&b, "--- issue %s %q (%s, %s, priority %s, tags %v, parent %q, depends on %v)\n", is.ID, is.Title, is.Kind, is.State, is.Priority, is.Tags, is.Parent, is.DependsOn)
			if is.Requirement != "" {
				fmt.Fprintf(&b, "requirement:\n%s\n", clip(is.Requirement))
			}
			if is.OpenQuestions != "" {
				fmt.Fprintf(&b, "open questions:\n%s\n", clip(is.OpenQuestions))
			}
			if is.Context != "" {
				fmt.Fprintf(&b, "context:\n%s\n", clip(is.Context))
			}
			for i, c := range is.Criteria {
				mark := " "
				if c.Checked {
					mark = "x"
				}
				fmt.Fprintf(&b, "criterion %d [%s] %s\n", i+1, mark, clip(c.Text))
			}
		}
		for _, path := range append(append([]string{}, pd.EntriesCreated...), pd.EntriesChanged...) {
			if e := ps.Knowledge[path]; e != nil {
				fmt.Fprintf(&b, "--- knowledge entry %s (%s) %q\n%s\n", e.Path, e.Type, e.Title, clip(e.Text))
			}
		}
	}
	return b.String()
}
