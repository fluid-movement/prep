package domain

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// EntryIssues lists the issues whose completion changed an entry (their
// documentation decision names it), newest first.
func (t *Tree) EntryIssues(path string) []string {
	var out []string
	for _, id := range t.IDs() {
		r := t.Issues[id].Resolution
		if r != nil && r.Documentation != nil && slices.Contains(r.Documentation.Entries, path) {
			out = append(out, id)
		}
	}
	slices.Reverse(out)
	return out
}

// IssueEntries lists the knowledge entries in play for an issue: those its
// context links, and once resolved, those its completion changed. Only
// existing entries, each once, in that order.
func (t *Tree) IssueEntries(id string) []string {
	i := t.Issues[id]
	if i == nil {
		return nil
	}
	var out []string
	add := func(p string) {
		if t.Knowledge[p] != nil && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	for _, p := range i.ContextLinks {
		add(p)
	}
	if r := i.Resolution; r != nil && r.Documentation != nil {
		for _, p := range r.Documentation.Entries {
			add(p)
		}
	}
	return out
}

// KnowledgeFilter is a query over knowledge entries. Different fields AND;
// repeated values within a field OR.
type KnowledgeFilter struct {
	Types    []string
	Statuses []string
	Scopes   []string // a path the entry's scope covers or lies under
	Text     []string // case-insensitive title substring
}

// ParseKnowledgeFilter parses `--type component --status draft --scope
// internal/cli words`; bare words join into one title phrase.
func ParseKnowledgeFilter(args []string) (KnowledgeFilter, error) {
	var f KnowledgeFilter
	var words []string
	for k := 0; k < len(args); k++ {
		a := args[k]
		if !strings.HasPrefix(a, "--") {
			words = append(words, a)
			continue
		}
		name, val, hasVal := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		if !hasVal {
			if k+1 >= len(args) {
				return f, fmt.Errorf("--%s needs a value", name)
			}
			k++
			val = args[k]
		}
		vals := strings.Split(strings.ToLower(val), ",")
		switch name {
		case "type":
			for _, v := range vals {
				if !slices.Contains(KnowledgeTypes, v) {
					return f, fmt.Errorf("--type must be one of %s", strings.Join(KnowledgeTypes, ", "))
				}
			}
			f.Types = append(f.Types, vals...)
		case "status":
			f.Statuses = append(f.Statuses, vals...)
		case "scope":
			f.Scopes = append(f.Scopes, strings.Split(val, ",")...)
		default:
			return f, fmt.Errorf("unknown knowledge flag --%s (use --type, --status, --scope or words)", name)
		}
	}
	if len(words) > 0 {
		f.Text = []string{strings.ToLower(strings.Join(words, " "))}
	}
	return f, nil
}

// QueryKnowledge returns the paths of the entries matching f, ordered by path.
func (t *Tree) QueryKnowledge(f KnowledgeFilter) []string {
	var out []string
	for p, e := range t.Knowledge {
		if len(f.Types) > 0 && !slices.Contains(f.Types, e.Type) {
			continue
		}
		if len(f.Statuses) > 0 && !slices.Contains(f.Statuses, e.Status) {
			continue
		}
		if len(f.Scopes) > 0 && !slices.ContainsFunc(f.Scopes, func(path string) bool {
			// Either side may be the wider one: --scope internal finds an entry
			// scoped to internal/cli, --scope internal/cli/x.go one scoped to internal/cli.
			return slices.ContainsFunc(e.Scope, func(s string) bool { return ScopeMatch(s, path) || ScopeMatch(path, s) })
		}) {
			continue
		}
		if len(f.Text) > 0 && !strings.Contains(strings.ToLower(e.Title), f.Text[0]) {
			continue
		}
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
