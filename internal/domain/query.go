package domain

import (
	"fmt"
	"strings"
)

// Filter is a query over issues. Different fields AND; repeated values within
// a field OR. The CLI and the TUI share it, so both agree on what each derived
// state means.
type Filter struct {
	States     []State
	Kinds      []Kind
	Under      []string
	Stale      *bool
	Actionable *bool
	Blocked    *bool
	Parent     *bool // true: only parents, false: only leaves
	TopLevel   bool
	Text       []string // case-insensitive title substring
}

// ParseFilter parses query flags such as `--state ready --kind code --under <id>`.
// Boolean flags accept an optional =false (e.g. --stale=false).
func ParseFilter(args []string) (Filter, error) {
	var f Filter
	boolp := func(v string) (*bool, error) {
		switch v {
		case "", "true":
			b := true
			return &b, nil
		case "false":
			b := false
			return &b, nil
		}
		return nil, fmt.Errorf("invalid boolean %q", v)
	}
	for k := 0; k < len(args); k++ {
		a := args[k]
		if !strings.HasPrefix(a, "--") {
			return f, fmt.Errorf("unexpected argument %q", a)
		}
		name, val, hasVal := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		takeVal := func() (string, error) {
			if hasVal {
				return val, nil
			}
			if k+1 >= len(args) {
				return "", fmt.Errorf("--%s needs a value", name)
			}
			k++
			return args[k], nil
		}
		var err error
		switch name {
		case "state":
			var v string
			if v, err = takeVal(); err == nil {
				for _, s := range strings.Split(v, ",") {
					st := State(strings.ReplaceAll(strings.TrimSpace(s), "-", "_"))
					if !hasState(States, st) {
						return f, fmt.Errorf("unknown state %q", s)
					}
					f.States = append(f.States, st)
				}
			}
		case "kind":
			var v string
			if v, err = takeVal(); err == nil {
				for _, s := range strings.Split(v, ",") {
					kd := Kind(strings.TrimSpace(s))
					if !kd.Valid() {
						return f, fmt.Errorf("unknown kind %q", s)
					}
					f.Kinds = append(f.Kinds, kd)
				}
			}
		case "under":
			var v string
			if v, err = takeVal(); err == nil {
				f.Under = append(f.Under, v)
			}
		case "text":
			var v string
			if v, err = takeVal(); err == nil {
				f.Text = append(f.Text, strings.ToLower(v))
			}
		case "stale":
			f.Stale, err = boolp(val)
		case "actionable":
			f.Actionable, err = boolp(val)
		case "blocked":
			f.Blocked, err = boolp(val)
		case "parent":
			f.Parent, err = boolp(val)
		case "leaf":
			var b *bool
			if b, err = boolp(val); err == nil {
				nb := !*b
				f.Parent = &nb
			}
		case "top":
			f.TopLevel = true
		default:
			return f, fmt.Errorf("unknown query flag --%s", name)
		}
		if err != nil {
			return f, err
		}
	}
	return f, nil
}

// Query returns the IDs matching the filter, in chronological order.
// Under references are resolved as ID suffixes.
func (t *Tree) Query(f Filter) ([]string, error) {
	var under map[string]bool
	if len(f.Under) > 0 {
		under = map[string]bool{}
		for _, ref := range f.Under {
			id, err := t.Resolve(ref)
			if err != nil {
				return nil, err
			}
			for _, d := range t.Descendants(id) {
				under[d] = true
			}
		}
	}
	var out []string
	for _, id := range t.IDs() {
		i := t.Issues[id]
		if len(f.States) > 0 && !hasState(f.States, t.State(id)) {
			continue
		}
		if len(f.Kinds) > 0 && !hasKind(f.Kinds, i.Kind) {
			continue
		}
		if under != nil && !under[id] {
			continue
		}
		if f.Stale != nil && t.Stale(id) != *f.Stale {
			continue
		}
		if f.Actionable != nil && t.Actionable(id) != *f.Actionable {
			continue
		}
		if f.Blocked != nil && t.Blocked(id) != *f.Blocked {
			continue
		}
		if f.Parent != nil && t.IsParent(id) != *f.Parent {
			continue
		}
		if f.TopLevel && i.Parent != "" && t.Issues[i.Parent] != nil {
			continue
		}
		if len(f.Text) > 0 {
			hit := false
			for _, s := range f.Text {
				if strings.Contains(strings.ToLower(i.Title), s) {
					hit = true
				}
			}
			if !hit {
				continue
			}
		}
		out = append(out, id)
	}
	return out, nil
}

func hasKind(ks []Kind, k Kind) bool {
	for _, x := range ks {
		if x == k {
			return true
		}
	}
	return false
}

// WithAncestors expands a result set with every ancestor, for tree display.
func (t *Tree) WithAncestors(ids []string) []string {
	set := map[string]bool{}
	for _, id := range ids {
		set[id] = true
		for _, a := range t.Ancestors(id) {
			set[a] = true
		}
	}
	var out []string
	for _, id := range t.IDs() {
		if set[id] {
			out = append(out, id)
		}
	}
	return out
}

// Depth returns the number of ancestors of an issue.
func (t *Tree) Depth(id string) int { return len(t.Ancestors(id)) }

// TreeOrder orders ids depth first under their parents, roots chronological.
func (t *Tree) TreeOrder(ids []string) []string {
	set := map[string]bool{}
	for _, id := range ids {
		set[id] = true
	}
	var out []string
	done := map[string]bool{}
	var walk func(string)
	walk = func(id string) {
		if done[id] {
			return
		}
		done[id] = true
		if set[id] {
			out = append(out, id)
		}
		for _, c := range t.children[id] {
			walk(c)
		}
	}
	for _, id := range t.IDs() {
		p := t.Issues[id].Parent
		if p == "" || t.Issues[p] == nil || !set[p] {
			// a root of the displayed forest
			isRoot := true
			for _, a := range t.Ancestors(id) {
				if set[a] {
					isRoot = false
				}
			}
			if isRoot {
				walk(id)
			}
		}
	}
	for _, id := range ids { // anything left (cycles)
		if !done[id] {
			out = append(out, id)
		}
	}
	return out
}
