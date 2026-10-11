package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// Resolution is where a link points on this machine.
type Resolution struct {
	domain.ProjectLink
	// Checkout is the most recently used checkout with the link's ID; nil
	// when the project is not checked out here (or not registered: prep
	// projects add registers a checkout).
	Checkout *Project `json:"checkout,omitempty"`
}

// ResolveLinks resolves each of the project's links through the registry.
func ResolveLinks(p domain.Project) ([]Resolution, error) {
	ps, err := List()
	if err != nil {
		return nil, err
	}
	out := make([]Resolution, 0, len(p.Links))
	for _, l := range p.Links {
		r := Resolution{ProjectLink: l}
		for _, q := range ps {
			if q.ID == l.ID {
				q := q
				r.Checkout = &q
				break
			}
		}
		out = append(out, r)
	}
	return out, nil
}

// Check reports what only the other projects on this machine can tell
// about t's links: a clone hint that differs from the checkout's remote. A
// project that is not checked out is never reported.
func Check(t *domain.Tree) []domain.Diagnostic {
	if len(t.Project.Links) == 0 {
		return nil
	}
	rs, err := ResolveLinks(t.Project)
	if err != nil {
		return nil
	}
	var ds []domain.Diagnostic
	for _, id := range t.IDs() {
		if t.State(id).Terminal() {
			continue
		}
		for _, d := range t.Issues[id].DependsOn {
			x, ok := t.External[d]
			switch {
			case !ok:
			case x.Status == domain.XMissing:
				ds = append(ds, domain.Diagnostic{Code: domain.CodeXDepMissing, Severity: domain.SevWarning, Class: domain.ClassManual, Issue: id, File: "issue.md",
					Message: fmt.Sprintf("dependency %s: no such issue in %s (%s)", d, x.Project, x.Root),
					Fix:     "pull that checkout if it is behind, else fix or remove the dependency"})
			case x.Status == domain.XFound && x.State == domain.StateDropped:
				ds = append(ds, domain.Diagnostic{Code: domain.CodeXDepDropped, Severity: domain.SevWarning, Class: domain.ClassGuided, Issue: id, File: "issue.md",
					Message: fmt.Sprintf("depends on %s, dropped in %s", d, x.Project),
					Fix:     "remove the dependency or replace it with the issue that supersedes it"})
			}
		}
	}
	checkouts := map[string]*Project{}
	for _, r := range rs {
		checkouts[r.Name] = r.Checkout
	}
	for _, ref := range t.LinkRefs() {
		co := checkouts[ref.Link]
		if ref.Dep || co == nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(co.Root, mdstore.Dir, "knowledge", filepath.FromSlash(strings.TrimPrefix(ref.Target, "/")))); err == nil {
			continue
		}
		ds = append(ds, domain.Diagnostic{Code: domain.CodeXKnowledgeLink, Severity: domain.SevWarning, Class: domain.ClassManual, Issue: ref.Issue, File: ref.File,
			Message: fmt.Sprintf("%s:%s: no such entry in %s (%s)", ref.Link, ref.Target, co.Name, co.Root),
			Fix:     "pull that checkout if it is behind, else fix the link"})
	}
	for _, r := range rs {
		if r.Checkout == nil || r.Checkout.Remote == "" || r.Checkout.Remote == r.Remote {
			continue
		}
		ds = append(ds, domain.Diagnostic{Code: domain.CodeLinkRemote, Severity: domain.SevWarning, Class: domain.ClassFixable, File: "project.md",
			Message: fmt.Sprintf("link %s: the clone hint is %q, its checkout's remote is %s", r.Name, r.Remote, r.Checkout.Remote),
			Fix:     "run prep fix"})
	}
	return ds
}

// FixRemotes returns t's links with each clone hint set to its checkout's
// remote, and whether any changed.
func FixRemotes(p domain.Project) ([]domain.ProjectLink, bool) {
	rs, err := ResolveLinks(p)
	if err != nil {
		return nil, false
	}
	changed := false
	var out []domain.ProjectLink
	for _, r := range rs {
		l := r.ProjectLink
		if r.Checkout != nil && r.Checkout.Remote != "" && r.Checkout.Remote != l.Remote {
			l.Remote, changed = r.Checkout.Remote, true
		}
		out = append(out, l)
	}
	return out, changed
}

// issueTrees loads other projects' issues (no knowledge) once per root.
type issueTrees map[string]*domain.Tree

func (c issueTrees) get(root string) (*domain.Tree, error) {
	if t, ok := c[root]; ok {
		return t, nil
	}
	p, issues, _, err := mdstore.Open(root).Load()
	if err != nil {
		return nil, err
	}
	t := domain.NewTree(p, issues, nil, nil)
	c[root] = t
	return t, nil
}

// LoadTree returns the issues of the project at root, without its
// knowledge base: what a cross-project reference needs.
func LoadTree(root string) (*domain.Tree, error) { return issueTrees{}.get(root) }

// ResolveExternal fills t.External with every cross-project dependency in
// t: found (with state and title), missing in the checked-out project, or
// unknown when the project is not checked out here. It loads each other
// project once, and nothing when t has no cross-project dependencies.
func ResolveExternal(t *domain.Tree) {
	refs := t.XRefs()
	if len(refs) == 0 {
		return
	}
	rs, err := ResolveLinks(t.Project)
	if err != nil {
		return
	}
	byName := map[string]Resolution{}
	for _, r := range rs {
		byName[r.Name] = r
	}
	trees := issueTrees{}
	t.External = map[string]domain.XIssue{}
	for _, ref := range refs {
		x := t.XIssueOf(ref)
		r, ok := byName[x.Link]
		if ok && r.Checkout != nil {
			x.Project, x.Root = r.Checkout.Name, r.Checkout.Root
			if other, err := trees.get(r.Checkout.Root); err == nil {
				x.Status = domain.XMissing
				if i := other.Issues[x.ID]; i != nil {
					x.Status, x.State, x.Title = domain.XFound, other.State(x.ID), i.Title
				}
			}
		}
		t.External[ref] = x
	}
}

// ResolveDep turns a cross-project dependency as given ("<project>:<ID or
// suffix>") into its stored form "<link>:<full ID>". The project part is a
// link of t, else a registered project, which then needs a new link
// (returned for the caller to add). A suffix is resolved in the project's
// checkout; without one the ID must be complete. Dropped issues are refused.
func ResolveDep(t *domain.Tree, link, rest string) (ref string, add *domain.ProjectLink, x domain.XIssue, err error) {
	l, ok := t.Project.Link(link)
	if !ok {
		p, err := Resolve(link)
		if err != nil {
			return "", nil, x, err
		}
		if p.ID == "" {
			return "", nil, x, &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf("%s has no project ID; run prep fix there and commit it", p.Root)}
		}
		if have, ok := t.Project.LinkByID(p.ID); ok {
			l = have
		} else {
			name := link
			if !domain.ValidProjectName(name) {
				name = p.Name
			}
			if _, taken := t.Project.Link(name); taken {
				return "", nil, x, &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf("a link named %s points at another project; prep link add %s --as <name> first", name, p.Root)}
			}
			l = domain.ProjectLink{Name: name, ID: p.ID, Remote: p.Remote}
			add = &l
		}
	}
	x = domain.XIssue{Link: l.Name, Status: domain.XUnknown, Remote: l.Remote}
	co, found, err := ByID(l.ID)
	if err != nil {
		return "", nil, x, err
	}
	if !found {
		if !domain.ValidID(strings.ToUpper(rest)) {
			return "", nil, x, &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf("project %s is not checked out here%s; give the full issue ID", l.Name, cloneHint(l.Remote))}
		}
		x.ID = strings.ToUpper(rest)
	} else {
		other, err := LoadTree(co.Root)
		if err != nil {
			return "", nil, x, err
		}
		id, err := other.Resolve(rest)
		if err != nil {
			return "", nil, x, &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("in project %s: %v", l.Name, err)}
		}
		if other.State(id) == domain.StateDropped {
			return "", nil, x, &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf("cannot depend on dropped issue %s:%s", l.Name, id)}
		}
		x.ID, x.Status, x.State, x.Title, x.Project, x.Root = id, domain.XFound, other.State(id), other.Issues[id].Title, co.Name, co.Root
	}
	x.Ref = l.Name + ":" + x.ID
	return x.Ref, add, x, nil
}

func cloneHint(remote string) string {
	if remote == "" {
		return ""
	}
	return " (clone " + remote + ")"
}

// Dependent is an issue in another project that depends on one of this
// project's issues.
type Dependent struct {
	Project string       `json:"project"`
	Root    string       `json:"root"`
	ID      string       `json:"id"`
	Title   string       `json:"title"`
	State   domain.State `json:"state"`
}

// Dependents scans the registered checkouts of other projects for issues
// that depend on issue id of the project with the given ID.
func Dependents(projectID, id string) ([]Dependent, error) {
	ps, err := List()
	if err != nil {
		return nil, err
	}
	var out []Dependent
	seen := map[string]bool{}
	for _, p := range ps {
		if p.ID == projectID || seen[p.key()] {
			continue
		}
		seen[p.key()] = true
		other, err := LoadTree(p.Root)
		if err != nil {
			continue
		}
		l, ok := other.Project.LinkByID(projectID)
		if !ok {
			continue
		}
		for _, oid := range other.IDs() {
			for _, d := range other.Issues[oid].DependsOn {
				if d == l.Name+":"+id {
					out = append(out, Dependent{Project: p.Name, Root: p.Root, ID: oid, Title: other.Issues[oid].Title, State: other.State(oid)})
				}
			}
		}
	}
	return out, nil
}
