package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/projects"
)

// validate is domain.Validate plus what only the other projects on this
// machine can tell (projects.Check): every check the CLI and the TUI show.
func validate(t *domain.Tree) []domain.Diagnostic {
	ds := append(domain.Validate(t), projects.Check(t)...)
	domain.SortDiagnostics(ds)
	return ds
}

// cmdLinks lists the project's links and where each points on this
// machine.
func cmdLinks(a *app, args []string) error {
	if len(args) > 0 {
		return usageErr("usage: prep links")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	rs, err := projects.ResolveLinks(t.Project)
	if err != nil {
		return err
	}
	a.noteRead(activity.Event{Op: "links", Verb: "listed the links", Area: activity.AreaPrep})
	if a.json {
		a.emit(map[string]any{"links": rs})
		return nil
	}
	if len(rs) == 0 {
		a.printf("no links; prep link add <project> links one (prep projects lists them)\n")
		return nil
	}
	for _, r := range rs {
		where := "not checked out here"
		if r.Remote != "" {
			where += "; clone " + r.Remote
		}
		if c := r.Checkout; c != nil {
			where = c.Root
			if c.Name != r.Name {
				where += " (named " + c.Name + " there)"
			}
		}
		a.printf("%-16s %s  %s\n", r.Name, r.ID[len(r.ID)-6:], where)
	}
	return nil
}

// cmdLink changes the project's links: add, rename (with every reference)
// and remove (refused while referenced).
func cmdLink(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("usage: prep link add <project> [--as name] | rename <old> <new> | remove <name>")
	}
	sub, args := args[0], args[1:]
	fs := flag.NewFlagSet("link "+sub, flag.ContinueOnError)
	as := fs.String("as", "", "the local name (default: the project's own name)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	var c *domain.Change
	var name, verb string
	switch {
	case sub == "add" && len(pos) == 1:
		l, err := linkTo(t, pos[0], *as)
		if err != nil {
			return err
		}
		if c, err = t.PlanLinks(append(append([]domain.ProjectLink{}, t.Project.Links...), l)); err != nil {
			return err
		}
		name, verb = l.Name, "linked project "+l.Name
	case sub == "rename" && len(pos) == 2 && *as == "":
		if c, err = t.PlanLinkRename(pos[0], pos[1]); err != nil {
			return err
		}
		name, verb = pos[1], "renamed link "+pos[0]+" to "+pos[1]
	case sub == "remove" && len(pos) == 1 && *as == "":
		if c, err = t.PlanLinkRemove(pos[0]); err != nil {
			return err
		}
		name, verb = pos[0], "removed link "+pos[0]
	default:
		return usageErr("usage: prep link add <project> [--as name] | rename <old> <new> | remove <name>")
	}
	files, err := a.applyAll(t, []*domain.Change{c})
	if err != nil {
		return err
	}
	a.note(activity.Event{Op: "link " + sub, Verb: verb, Target: name, Area: activity.AreaPrep})
	a.reportWrite(writeResult{OK: true, Op: "link " + sub, Entry: name, Files: files})
	return nil
}

// linkTo builds the link to the registered project query names, under name
// or else the project's own name.
func linkTo(t *domain.Tree, query, name string) (domain.ProjectLink, error) {
	p, err := projects.Resolve(query)
	if err != nil {
		return domain.ProjectLink{}, err
	}
	if p.ID == "" {
		return domain.ProjectLink{}, usageErr("%s has no project ID; run prep fix there (prep fix --project %s) and commit it", p.Root, p.Root)
	}
	if name == "" {
		name = p.Name
	}
	if l, ok := t.Project.LinkByID(p.ID); ok {
		return domain.ProjectLink{}, usageErr("%s is linked already as %s", p.Name, l.Name)
	}
	if _, ok := t.Project.Link(name); ok {
		return domain.ProjectLink{}, usageErr("a link named %s exists already; choose another with --as", name)
	}
	if !domain.ValidProjectName(name) {
		return domain.ProjectLink{}, usageErr("link name %q is invalid: use %s", name, domain.ProjectNameRule)
	}
	return domain.ProjectLink{Name: name, ID: p.ID, Remote: p.Remote}, nil
}

// fixLinkRemotes updates clone hints to their checkouts' remotes (X004).
func (a *app) fixLinkRemotes() ([]string, error) {
	t, err := a.load(false)
	if err != nil {
		return nil, err
	}
	links, changed := projects.FixRemotes(t.Project)
	if !changed {
		return nil, nil
	}
	c, err := t.PlanLinks(links)
	if err != nil {
		return nil, fmt.Errorf("links: %w", err)
	}
	return a.store.Apply(c)
}

// resolveDeps resolves --depends-on values: local IDs or suffixes, and
// cross-project ones (<project>:<ID or suffix>) through the registry. It
// records what it found in t.External and returns the links the
// dependencies need that the project does not have yet.
func (a *app) resolveDeps(t *domain.Tree, deps []string) ([]string, []domain.ProjectLink, error) {
	var ids []string
	var add []domain.ProjectLink
	for _, d := range deps {
		name, rest, ok := domain.SplitRef(d)
		if !ok {
			id, err := t.Resolve(d)
			if err != nil {
				var de *domain.Error
				if errors.As(err, &de) && de.Code == domain.ErrNotFound && !strings.Contains(de.Message, "<project>") {
					de.Message += "; an issue in another project is written <project>:<id> (prep projects lists the projects, --depends-on api:01ABC...)"
				}
				return nil, nil, err
			}
			ids = append(ids, id)
			continue
		}
		view := *t
		view.Project.Links = append(append([]domain.ProjectLink{}, t.Project.Links...), add...)
		ref, l, x, err := projects.ResolveDep(&view, name, rest)
		if err != nil {
			return nil, nil, err
		}
		if l != nil {
			add = append(add, *l)
		}
		if t.External == nil {
			t.External = map[string]domain.XIssue{}
		}
		t.External[ref] = x
		ids = append(ids, ref)
	}
	return ids, add, nil
}

// withLinks plans adding links the dependencies need: the change to apply
// first and the tree to plan the rest against. Without new links it
// returns nil and t.
func withLinks(t *domain.Tree, add []domain.ProjectLink) (*domain.Change, *domain.Tree, error) {
	if len(add) == 0 {
		return nil, t, nil
	}
	c, err := t.PlanLinks(append(append([]domain.ProjectLink{}, t.Project.Links...), add...))
	if err != nil {
		return nil, nil, err
	}
	return c, t.Apply(c), nil
}
