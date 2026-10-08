package domain

import (
	"fmt"
	"path"
	"strings"
	"time"
)

// KnowledgeEdit creates or updates one knowledge entry. Nil fields stay as
// they are on update. The path never changes: entries are never renamed.
type KnowledgeEdit struct {
	Path            string // bundle-relative, normalized by PlanKnowledge
	New             bool
	Type            *string
	Title           *string
	Description     *string
	Status          *string
	Scope           *[]string
	ConfirmedCommit *string
	Body            *string
	Actor           string
	At              time.Time
}

// PlanKnowledge checks a knowledge write as far as the domain can without
// the adapter. The adapter then renders the entry (KnowledgeStore.Render)
// and CheckWrite validates the tree with it, including links.
func (t *Tree) PlanKnowledge(in KnowledgeEdit) (*Change, error) {
	p := NormalizeEntryPath(in.Path)
	if strings.Contains(p, "..") || strings.Contains(p, "//") || p == "/.md" || strings.ContainsAny(p, " \t\\") {
		return nil, &Error{Code: ErrUsage, Message: fmt.Sprintf("%q is not a valid entry path; use a bundle path such as /components/export.md", in.Path)}
	}
	if b := path.Base(p); b == "index.md" || b == "log.md" {
		return nil, &Error{Code: ErrUsage, Message: fmt.Sprintf("%s is reserved: prep generates index.md and keeps log.md for a chronological log; name the entry differently", p)}
	}
	in.Path = p
	in.At = in.At.UTC().Truncate(time.Second)
	if in.Body != nil && strings.TrimSpace(*in.Body) == "" {
		return nil, &Error{Code: ErrUsage, Message: "the body is empty; an entry needs content (nothing was written)"}
	}
	exists := t.Knowledge[p] != nil
	empty := func(s *string) bool { return s == nil || strings.TrimSpace(*s) == "" }
	switch {
	case in.New && exists:
		return nil, &Error{Code: ErrInvalid, Message: fmt.Sprintf("entry %s already exists; use prep knowledge update", p)}
	case in.New:
		var unmet []Unmet
		for _, f := range []struct {
			v    *string
			name string
		}{{in.Type, "--type"}, {in.Title, "--title"}, {in.Description, "--description"}, {in.Body, "--body or --body-file"}} {
			if empty(f.v) {
				unmet = append(unmet, Unmet{GateRequirement, f.name + " is required"})
			}
		}
		if len(unmet) > 0 {
			return nil, &Error{Code: ErrUsage, Message: "cannot create " + p, Unmet: unmet}
		}
	case !exists:
		return nil, &Error{Code: ErrNotFound, Message: fmt.Sprintf("no knowledge entry %s; prep knowledge new creates one", p)}
	case in.Type == nil && in.Title == nil && in.Description == nil && in.Status == nil && in.Scope == nil && in.ConfirmedCommit == nil && in.Body == nil:
		return nil, &Error{Code: ErrUsage, Message: "nothing to change: pass a field, --scope or --body"}
	}
	if in.New && p != OverviewEntry && !t.Bootstrapped() {
		return nil, &Error{Code: ErrGate, Message: "cannot create " + p, Unmet: []Unmet{{GateBootstrap, t.BootstrapAlert() + " Entries come after the overview."}}}
	}
	if in.ConfirmedCommit != nil && *in.ConfirmedCommit != "" {
		scope := t.scopeOf(p, in.Scope)
		if len(scope) == 0 {
			return nil, &Error{Code: ErrInvalid, Message: fmt.Sprintf("%s has no scope; confirmed_commit only means something for scoped entries", p)}
		}
	}
	return &Change{Op: "knowledge", Knowledge: &in}, nil
}

func (t *Tree) scopeOf(p string, edit *[]string) []string {
	if edit != nil {
		return *edit
	}
	if e := t.Knowledge[p]; e != nil {
		return e.Scope
	}
	return nil
}
