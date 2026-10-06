package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Tree is the whole project as loaded from storage: every issue plus the
// knowledge base. Derived facts are computed on demand and never stored.
type Tree struct {
	Project   Project
	Issues    map[string]*Issue
	Knowledge map[string]*Entry // keyed by bundle-relative path
	// ParseDiags holds format-level diagnostics reported by the adapters.
	ParseDiags []Diagnostic

	children map[string][]string
}

// NewTree builds a tree from issues and knowledge entries.
func NewTree(p Project, issues []*Issue, entries []*Entry, parseDiags []Diagnostic) *Tree {
	t := &Tree{Project: p, Issues: map[string]*Issue{}, Knowledge: map[string]*Entry{}, ParseDiags: parseDiags}
	for _, i := range issues {
		t.Issues[i.ID] = i
	}
	for _, e := range entries {
		t.Knowledge[e.Path] = e
	}
	t.index()
	return t
}

func (t *Tree) index() {
	t.children = map[string][]string{}
	for _, id := range t.IDs() {
		i := t.Issues[id]
		if i.Parent != "" {
			t.children[i.Parent] = append(t.children[i.Parent], id)
		}
	}
}

// IDs returns all issue IDs in chronological order.
func (t *Tree) IDs() []string {
	ids := make([]string, 0, len(t.Issues))
	for id := range t.Issues {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

var stampPattern = regexp.MustCompile(`^\d{8}-\d{6}$`)

// ValidStamp reports whether s has the YYYYMMDD-HHMMSS shape of baseline
// names.
func ValidStamp(s string) bool { return stampPattern.MatchString(s) }

// Resolve maps a full ID or any unique suffix to an issue ID.
func (t *Tree) Resolve(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", &Error{Code: ErrUsage, Message: "issue id required"}
	}
	if _, ok := t.Issues[ref]; ok {
		return ref, nil
	}
	var matches []string
	for _, id := range t.IDs() {
		if strings.HasSuffix(id, ref) {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 0:
		return "", &Error{Code: ErrNotFound, Message: fmt.Sprintf("no issue matches %q", ref)}
	case 1:
		return matches[0], nil
	}
	return "", &Error{Code: ErrAmbiguous, Message: fmt.Sprintf("%q matches %d issues: %s", ref, len(matches), strings.Join(matches, ", "))}
}

// Children returns the direct children of an issue.
func (t *Tree) Children(id string) []string { return t.children[id] }

// IsParent reports whether an issue has children.
func (t *Tree) IsParent(id string) bool { return len(t.children[id]) > 0 }

// Descendants returns all issues below id, depth first.
func (t *Tree) Descendants(id string) []string {
	var out []string
	seen := map[string]bool{id: true}
	var walk func(string)
	walk = func(p string) {
		for _, c := range t.children[p] {
			if seen[c] {
				continue
			}
			seen[c] = true
			out = append(out, c)
			walk(c)
		}
	}
	walk(id)
	return out
}

// Ancestors returns the parent chain of an issue, nearest first. It stops at
// missing parents and cycles.
func (t *Tree) Ancestors(id string) []string {
	var out []string
	seen := map[string]bool{id: true}
	cur := t.Issues[id]
	for cur != nil && cur.Parent != "" && !seen[cur.Parent] {
		seen[cur.Parent] = true
		out = append(out, cur.Parent)
		cur = t.Issues[cur.Parent]
	}
	return out
}

// Blocks returns the issues that depend on id. It is computed, never stored.
func (t *Tree) Blocks(id string) []string {
	var out []string
	for _, oid := range t.IDs() {
		for _, d := range t.Issues[oid].DependsOn {
			if d == id {
				out = append(out, oid)
				break
			}
		}
	}
	return out
}

// ReadyValid reports whether ready.md exists and references the newest baseline.
func (t *Tree) ReadyValid(i *Issue) bool {
	b := i.LatestBaseline()
	return i.Ready != nil && b != nil && i.Ready.Baseline == b.Name
}

// State derives an issue's state from which records exist.
func (t *Tree) State(id string) State {
	i := t.Issues[id]
	if i == nil {
		return ""
	}
	switch {
	case i.Resolution != nil && i.Resolution.Outcome == OutcomeDropped:
		return StateDropped
	case i.Resolution != nil:
		return StateDone
	case i.Claim != nil && t.ReadyValid(i):
		return StateInProgress
	case t.ReadyValid(i):
		return StateReady
	case len(i.Baselines) > 0:
		return StateDefined
	}
	return StateOpen
}

// Stale reports whether the requirement or kind differs from the newest
// baseline. It applies to defined, ready and in-progress issues.
func (t *Tree) Stale(id string) bool {
	i := t.Issues[id]
	if i == nil {
		return false
	}
	switch t.State(id) {
	case StateDefined, StateReady, StateInProgress:
	default:
		return false
	}
	return t.Drifted(i)
}

// Drifted reports whether an issue's requirement or kind differs from its newest baseline.
func (t *Tree) Drifted(i *Issue) bool {
	b := i.LatestBaseline()
	if b == nil {
		return false
	}
	return b.Requirement != i.Body || b.Kind != i.Kind
}

// UnresolvedDeps returns dependencies that are not done.
func (t *Tree) UnresolvedDeps(id string) []string {
	var out []string
	for _, d := range t.Issues[id].DependsOn {
		if t.State(d) != StateDone {
			out = append(out, d)
		}
	}
	return out
}

// Blocked reports whether an unresolved dependency prevents claiming.
func (t *Tree) Blocked(id string) bool {
	return !t.State(id).Terminal() && len(t.UnresolvedDeps(id)) > 0
}

// Actionable = ready, not stale, dependencies done, unclaimed, and a leaf
// (parents are not implemented themselves).
func (t *Tree) Actionable(id string) bool {
	return t.State(id) == StateReady && !t.Stale(id) && len(t.UnresolvedDeps(id)) == 0 && !t.IsParent(id)
}

// Progress counts resolved direct children of a parent.
type Progress struct {
	Done    int `json:"done"`
	Dropped int `json:"dropped"`
	Total   int `json:"total"`
}

// ChildProgress computes progress over the direct children of id.
func (t *Tree) ChildProgress(id string) Progress {
	var p Progress
	for _, c := range t.children[id] {
		p.Total++
		switch t.State(c) {
		case StateDone:
			p.Done++
		case StateDropped:
			p.Dropped++
		}
	}
	return p
}

// EffectiveDoD cascades the Definition of Done project → ancestors → issue,
// applying opt-outs. It returns the remaining items and the opt-outs applied.
func (t *Tree) EffectiveDoD(id string) ([]string, []OptOut) {
	chain := []*Issue{}
	anc := t.Ancestors(id)
	for k := len(anc) - 1; k >= 0; k-- {
		if a := t.Issues[anc[k]]; a != nil {
			chain = append(chain, a)
		}
	}
	if i := t.Issues[id]; i != nil {
		chain = append(chain, i)
	}
	items := append([]string(nil), t.Project.DoD...)
	var applied []OptOut
	for _, i := range chain {
		for _, o := range i.DoDOptOuts {
			for k, it := range items {
				if it == o.Item {
					items = append(items[:k], items[k+1:]...)
					applied = append(applied, o)
					break
				}
			}
		}
		items = append(items, i.DoDAdd...)
	}
	return items, applied
}

// CheckedAll reports whether every acceptance criterion is checked.
func CheckedAll(cs []Criterion) bool {
	for _, c := range cs {
		if !c.Checked {
			return false
		}
	}
	return true
}

// HasOutcome reports whether decisions contain an outcome entry.
func HasOutcome(ds []Decision) bool {
	for _, d := range ds {
		if d.Outcome {
			return true
		}
	}
	return false
}
