package projects

import (
	"fmt"

	"github.com/fluid-movement/prep/internal/domain"
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
