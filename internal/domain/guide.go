package domain

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

// Candidate is a knowledge entry suggested for an issue, listed by title and
// description only; the agent opens what it needs.
type Candidate struct {
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Via         string `json:"via"`
}

// Candidates resolves knowledge context for an issue. Entry points, in order:
// links from the issue's context, the parent chain's links, entries whose
// scope overlaps the files the issue touches, and the overview entry.
func (t *Tree) Candidates(id string, touched []string) []Candidate {
	var out []Candidate
	seen := map[string]bool{}
	addEntry := func(p, via string) {
		e := t.Knowledge[p]
		if e == nil || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, Candidate{Path: p, Title: e.Title, Description: e.Description, Via: via})
	}
	i := t.Issues[id]
	for _, l := range i.ContextLinks {
		addEntry(l, "context")
	}
	for _, a := range t.Ancestors(id) {
		if ai := t.Issues[a]; ai != nil {
			for _, l := range ai.ContextLinks {
				addEntry(l, "parent:"+a)
			}
		}
	}
	files := append(append([]string(nil), i.ContextPaths...), touched...)
	for _, p := range sortedKeys(t.Knowledge) {
		e := t.Knowledge[p]
		for _, s := range e.Scope {
			if overlaps(s, files) {
				addEntry(p, "scope:"+s)
				break
			}
		}
	}
	addEntry(OverviewEntry, "overview")
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func overlaps(scope string, files []string) bool {
	for _, f := range files {
		if ScopeMatch(scope, f) || ScopeMatch(f, scope) {
			return true
		}
	}
	return false
}

// ScopeMatch reports whether a path matches a scope pattern. Patterns support
// *, ? and **; a plain directory matches everything below it.
func ScopeMatch(pattern, p string) bool {
	pattern = strings.TrimPrefix(strings.TrimSpace(pattern), "./")
	p = strings.TrimPrefix(strings.TrimSpace(p), "./")
	if pattern == "" || p == "" {
		return false
	}
	if !strings.ContainsAny(pattern, "*?[") {
		dir := strings.TrimSuffix(pattern, "/")
		return p == dir || strings.HasPrefix(p, dir+"/")
	}
	return globRegexp(pattern).MatchString(p)
}

var globCache = map[string]*regexp.Regexp{}

func globRegexp(pattern string) *regexp.Regexp {
	if re, ok := globCache[pattern]; ok {
		return re
	}
	var b strings.Builder
	b.WriteString("^")
	for k := 0; k < len(pattern); k++ {
		c := pattern[k]
		switch {
		case c == '*' && k+1 < len(pattern) && pattern[k+1] == '*':
			k++
			if k+1 < len(pattern) && pattern[k+1] == '/' {
				k++
				b.WriteString("(?:.*/)?")
			} else {
				b.WriteString(".*")
			}
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("(?:/.*)?$")
	re := regexp.MustCompile(b.String())
	globCache[pattern] = re
	return re
}

// Pointer is an input the agent may read, by path.
type Pointer struct {
	Path    string `json:"path"`
	What    string `json:"what"`
	Command string `json:"command,omitempty"` // the write command for this record
}

// Guide is the step contract for an issue: where it is, what can happen
// next, what blocks it, what to read and where outputs go.
type Guide struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Kind         Kind         `json:"kind"`
	State        State        `json:"state"`
	Stale        bool         `json:"stale"`
	Parent       bool         `json:"parent"`
	Actionable   bool         `json:"actionable"`
	Step         string       `json:"step"`
	Instructions []string     `json:"instructions"`
	Transitions  []Transition `json:"transitions"`
	Inputs       []Pointer    `json:"inputs"`
	Knowledge    []Candidate  `json:"knowledge"`
	Outputs      []Pointer    `json:"outputs"`
	DoD          []string     `json:"definition_of_done,omitempty"`
	Diagnostics  []Diagnostic `json:"diagnostics,omitempty"`
}

// BuildGuide assembles the step contract. issueDir maps an ID to its
// directory path for pointers; touched lists files the issue touches.
func (t *Tree) BuildGuide(id string, issueDir func(string) string, touched []string) Guide {
	i := t.Issues[id]
	s := t.State(id)
	stale := t.Stale(id)
	parent := t.IsParent(id)
	g := Guide{ID: id, Title: i.Title, Kind: i.Kind, State: s, Stale: stale, Parent: parent, Actionable: t.Actionable(id)}
	dir := issueDir(id)
	f := func(name string) string { return path.Join(dir, name) }

	for _, op := range Ops {
		if !t.Applicable(id, op) {
			continue
		}
		unmet, needs := t.Gates(id, op, nil)
		cmd := "prep " + string(op) + " " + id
		for _, n := range needs {
			cmd += " " + n
		}
		g.Transitions = append(g.Transitions, Transition{Op: op, Command: cmd, Allowed: len(unmet) == 0, Unmet: unmet, Needs: needs})
	}

	// Inputs: the issue's own records first, then upstream outputs.
	g.Inputs = append(g.Inputs, Pointer{f("issue.md"), "requirement and open questions", ""})
	if s != StateOpen {
		for _, p := range []Pointer{{f("acceptance.md"), "acceptance criteria and Definition of Done additions", ""}, {f("context.md"), "implementation context", ""}, {f("decisions.md"), "decision records", ""}} {
			if i.HasFile(path.Base(p.Path)) {
				g.Inputs = append(g.Inputs, p)
			}
		}
		if i.HasFile("history.md") && (s == StateInProgress || s.Terminal()) {
			g.Inputs = append(g.Inputs, Pointer{f("history.md"), "work log", ""})
		}
	}
	for _, a := range t.Ancestors(id) {
		if ai := t.Issues[a]; ai != nil && ai.HasFile("context.md") && strings.TrimSpace(ai.Context) != "" {
			g.Inputs = append(g.Inputs, Pointer{path.Join(issueDir(a), "context.md"), "parent context: " + ai.Title, ""})
		}
	}
	for _, d := range i.DependsOn {
		di := t.Issues[d]
		if di == nil {
			continue
		}
		switch {
		case di.Kind == KindResearch && di.Findings != nil:
			g.Inputs = append(g.Inputs, Pointer{path.Join(issueDir(d), "findings.md"), "findings of dependency: " + di.Title, ""})
		case di.Kind == KindDecision && len(di.Decisions) > 0:
			g.Inputs = append(g.Inputs, Pointer{path.Join(issueDir(d), "decisions.md"), "outcome of dependency: " + di.Title, ""})
		case t.State(d) == StateDone:
			g.Inputs = append(g.Inputs, Pointer{path.Join(issueDir(d), "resolution.md"), "resolution of dependency: " + di.Title, ""})
		}
	}
	g.Knowledge = t.Candidates(id, touched)
	if s == StateInProgress || s == StateReady {
		g.DoD, _ = t.EffectiveDoD(id)
	}
	for _, d := range Validate(t) {
		if d.Issue == id {
			g.Diagnostics = append(g.Diagnostics, d)
		}
	}
	g.Step, g.Instructions, g.Outputs = stepContract(t, i, s, stale, parent, f)
	return g
}

func stepContract(t *Tree, i *Issue, s State, stale, parent bool, f func(string) string) (string, []string, []Pointer) {
	if stale {
		return "resolve-drift", []string{
			"The requirement or kind changed since the newest baseline (" + i.LatestBaseline().Name + "). Compare issue.md with " + f("baselines/"+i.LatestBaseline().Name+".md") + ".",
			"Trivial change (typo, rewording): run prep ack " + i.ID + ". A new baseline is written; enrichment stays valid.",
			"Real change: run prep define " + i.ID + ", then re-enrich: update the context (prep context) and criteria (prep criterion), and record new decisions with prep decide --supersedes <id> for the ones they replace. Never delete decisions. Finish with prep ready " + i.ID + ".",
			"An issue in progress must be released (prep release) before it can be re-defined.",
		}, []Pointer{{f("baselines/"), "new baseline, written by prep ack or prep define", ""}}
	}
	switch s {
	case StateOpen:
		return "define", []string{
			"Settle what and why with the user; the user has authority over the requirement.",
			"Write the requirement as prose with prep edit " + i.ID + " --body-file - (stdin). List anything unresolved under '## Open questions'.",
			"Checking code is fine when it tests whether a requirement makes sense; do not plan the implementation yet.",
			"When the Open questions section is empty, run prep define " + i.ID + " to write the baseline.",
		}, []Pointer{{f("issue.md"), "requirement prose, empty Open questions section", "prep edit " + i.ID + " --body-file -"}}
	case StateDefined:
		ins := []string{
			"Enrich: settle how. The agent drafts, the user agrees.",
			"Context: implementation context derived from the requirement, with prep context " + i.ID + " --body-file -. Link relevant knowledge entries with bundle-relative links such as [Export](/components/export.md); mention code paths in backticks.",
			"Decisions: prep decide " + i.ID + " --title <title> --body-file - appends a dated entry with rationale and alternatives considered; --supersedes D<n> replaces an earlier one.",
			"Acceptance: prep criterion " + i.ID + " --add <text> (repeatable) for checkable criteria; prep dod " + i.ID + " --add <item> or --opt-out <item> --reason <text> for the Definition of Done.",
		}
		switch {
		case parent:
			ins = append(ins, "This is a parent: it completes when all children are resolved and its own criteria are checked. Context is optional.")
		case i.Kind == KindResearch:
			ins = append(ins, "Research: context.md states the question, the scope, and what counts as answered.")
		case i.Kind == KindDecision:
			ins = append(ins, "Decision: context.md states the question and the options.")
		case i.Kind == KindManual:
			ins = append(ins, "Manual: enrichment is optional; acceptance criteria are required.")
		case i.Kind == KindCode:
			ins = append(ins, "Code: context from the codebase is required.")
		}
		ins = append(ins, "Then run prep ready "+i.ID+" to sign off against the current baseline.")
		return "enrich", ins, []Pointer{
			{f("context.md"), "implementation context", "prep context " + i.ID + " --body-file -"},
			{f("decisions.md"), "decision records", "prep decide " + i.ID + " --title <title> --body-file -"},
			{f("acceptance.md"), "acceptance criteria", "prep criterion " + i.ID + " --add <text>"},
		}
	case StateReady:
		if parent {
			return "complete-parent", []string{
				"Work happens in the children; use prep next --under " + i.ID + ".",
				"When all children are done or dropped and the criteria in acceptance.md are checked, run prep complete " + i.ID + " with a documentation decision.",
			}, []Pointer{{f("acceptance.md"), "check criteria", "prep criterion " + i.ID + " --check <n>"}}
		}
		if deps := t.UnresolvedDeps(i.ID); len(deps) > 0 {
			return "wait", []string{"Blocked by unresolved dependencies: " + strings.Join(deps, ", ") + "."}, nil
		}
		return "claim", []string{
			"Claim the issue before implementing: prep claim " + i.ID + ".",
		}, []Pointer{{f("claim.md"), "written by prep claim", ""}}
	case StateInProgress:
		ins := []string{
			"Implement against the context and acceptance criteria. Log notable steps with prep log " + i.ID + " <text>.",
			"Check criteria as they are met with prep criterion " + i.ID + " --check <n> (numbers as prep show lists them). Writes during implementation are reviewed as a batch at completion.",
			"Documentation step: update or create knowledge entries in .prep/knowledge for what is now true, or decide there is no impact.",
		}
		var outs []Pointer
		switch {
		case parent:
		case i.Kind == KindCode:
			ins = append(ins, "Commit the code, then run prep complete "+i.ID+" --commit <hash> --docs <entry>... (or --no-impact <reason>).")
		case i.Kind == KindResearch:
			ins = append(ins, "Write the answer with prep findings "+i.ID+" --body-file -, then run prep complete "+i.ID+" --docs <entry>... (or --no-impact <reason>).")
			outs = append(outs, Pointer{f("findings.md"), "research findings", "prep findings " + i.ID + " --body-file -"})
		case i.Kind == KindDecision:
			ins = append(ins, "After agreement, record the decision with prep decide "+i.ID+" --title <title> --body-file - --outcome, then run prep complete "+i.ID+" --docs <entry>... (or --no-impact <reason>).")
		case i.Kind == KindManual:
			ins = append(ins, "Manual issues are completed by the user, usually in the TUI; with the CLI: prep complete "+i.ID+" --no-impact <reason> or --docs <entry>.")
		}
		ins = append(ins, "Stopping without finishing: prep release "+i.ID+" --reason <why>.")
		outs = append(outs, Pointer{f("history.md"), "work log", "prep log " + i.ID + " <text>"}, Pointer{f("acceptance.md"), "checked criteria", "prep criterion " + i.ID + " --check <n>"}, Pointer{".prep/knowledge/", "knowledge entries changed by this issue", ""}, Pointer{f("resolution.md"), "written by prep complete", ""})
		return "implement", ins, outs
	case StateDone, StateDropped:
		return "resolved", []string{"The issue is resolved; its records are history. Create a new issue for further change."}, nil
	}
	return "", nil, nil
}
