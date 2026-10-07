package domain

import "sort"

// Flag kinds in FlagInfo.
const (
	FlagList = "list" // takes values; a comma-separated list matches any of them
	FlagBool = "bool" // a switch; booleans accept =false
	FlagText = "text" // free text
)

// FlagKind returns the kind of a query flag (FlagList, FlagBool or
// FlagText), or "" for an unknown flag; readers of typed queries use it to
// know which flags take a value.
func FlagKind(name string) string {
	for _, f := range NewTree(Project{}, nil, nil, nil).QueryFlags() {
		if f.Name == name {
			return f.Kind
		}
	}
	return ""
}

// FlagInfo describes one query flag and the values it accepts now.
type FlagInfo struct {
	Name   string      `json:"name"`
	Kind   string      `json:"kind"`
	Desc   string      `json:"description"`
	Values []FlagValue `json:"values,omitempty"`
	// Count is how many issues a boolean flag selects.
	Count *int `json:"count,omitempty"`
}

// FlagValue is one value of a list flag, with the number of issues it
// matches and a label (the title, for issue references).
type FlagValue struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
	Count int    `json:"count"`
}

// QueryFlags describes every flag ParseFilter accepts, in its order, with
// the values available in the tree. prep flags prints it; the TUI filter
// bar offers the same values.
func (t *Tree) QueryFlags() []FlagInfo {
	count := func(f Filter) int {
		ids, _ := t.Query(f)
		return len(ids)
	}
	yes := true
	boolean := func(name, desc string, f Filter) FlagInfo {
		n := count(f)
		return FlagInfo{Name: name, Kind: FlagBool, Desc: desc, Count: &n}
	}

	var states, kinds, prios, tags, under []FlagValue
	for _, s := range States {
		states = append(states, FlagValue{Value: string(s), Count: count(Filter{States: []State{s}})})
	}
	for _, k := range Kinds {
		kinds = append(kinds, FlagValue{Value: string(k), Count: count(Filter{Kinds: []Kind{k}})})
	}
	for _, p := range Priorities {
		prios = append(prios, FlagValue{Value: string(p), Count: count(Filter{Priorities: []Priority{p}})})
	}
	seen := map[string]bool{}
	for _, i := range t.Issues {
		for _, tag := range i.Tags {
			seen[tag] = true
		}
	}
	names := make([]string, 0, len(seen))
	for tag := range seen {
		names = append(names, tag)
	}
	sort.Strings(names)
	for _, tag := range names {
		tags = append(tags, FlagValue{Value: tag, Count: count(Filter{Tags: []string{tag}})})
	}
	for _, id := range t.IDs() {
		if t.IsParent(id) {
			under = append(under, FlagValue{Value: id, Label: t.Issues[id].Title, Count: len(t.Descendants(id))})
		}
	}

	return []FlagInfo{
		{Name: "state", Kind: FlagList, Desc: "lifecycle state", Values: states},
		{Name: "kind", Kind: FlagList, Desc: "issue kind", Values: kinds},
		{Name: "tag", Kind: FlagList, Desc: "tags in use", Values: tags},
		{Name: "priority", Kind: FlagList, Desc: "priority; medium matches issues without one", Values: prios},
		{Name: "under", Kind: FlagList, Desc: "descendants of an issue (any ID or unique suffix; parents listed)", Values: under},
		{Name: "text", Kind: FlagText, Desc: "case-insensitive text in the title or requirement"},
		boolean("stale", "requirement changed since the newest baseline", Filter{Stale: &yes}),
		boolean("actionable", "ready, not stale, dependencies done, unclaimed, not a parent", Filter{Actionable: &yes}),
		boolean("blocked", "has unresolved dependencies", Filter{Blocked: &yes}),
		boolean("parent", "has children", Filter{Parent: &yes}),
		boolean("leaf", "has no children", Filter{Parent: new(bool)}),
		boolean("top", "no parent", Filter{TopLevel: true}),
	}
}
