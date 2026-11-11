package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// A brief is the self-contained hand-off text for an agent that knows
// nothing of prep: an implementer, or with check a checker. It names no
// prep command and no record path except the file the report goes to.

// Roles a brief is written for.
const (
	RoleImplementer = "implementer"
	RoleChecker     = "checker"
)

// BriefDecisionsBound bounds the decisions section of a brief, in bytes.
const BriefDecisionsBound = 6000

// How a decision comes to apply to an issue.
const (
	ViaIssue    = "issue"    // linked to the issue itself
	ViaAncestor = "ancestor" // linked to a parent up the chain
	ViaScope    = "scope"    // its scope matches a path the context names
)

// AppliedDecision is a current decision that applies to an issue. From is
// the ancestor's ID for ViaAncestor and the matching scope for ViaScope.
type AppliedDecision struct {
	Entry *Entry
	Via   string
	From  string
}

// LinkedDecisions returns the current decisions linked to an issue and to
// its ancestors: the issue's own first, then each ancestor's, nearest
// first. Superseded decisions are left out.
func (t *Tree) LinkedDecisions(id string) []AppliedDecision {
	var out []AppliedDecision
	add := func(issue, via, from string) {
		for _, e := range t.IssueRecords(issue) {
			if e.Type == "decision" && !e.Superseded {
				out = append(out, AppliedDecision{Entry: e, Via: via, From: from})
			}
		}
	}
	add(id, ViaIssue, "")
	for _, a := range t.Ancestors(id) {
		add(a, ViaAncestor, a)
	}
	return out
}

// ApplicableDecisions returns the decisions that apply to an issue:
// LinkedDecisions, then the current decisions whose scope matches a path the
// issue's context names, most specific scope first.
func (t *Tree) ApplicableDecisions(id string) []AppliedDecision {
	out := t.LinkedDecisions(id)
	seen := map[string]bool{}
	for _, d := range out {
		seen[d.Entry.Path] = true
	}
	i := t.Issues[id]
	if i == nil {
		return out
	}
	var scoped []AppliedDecision
	for _, p := range sortedKeys(t.Knowledge) {
		e := t.Knowledge[p]
		if e.Type != "decision" || e.Superseded || seen[p] {
			continue
		}
		best := ""
		for _, s := range e.Scope {
			for _, f := range i.ContextPaths {
				if ScopeMatch(s, f) && (best == "" || scopeSpecificity(s) > scopeSpecificity(best)) {
					best = s
				}
			}
		}
		if best != "" {
			scoped = append(scoped, AppliedDecision{Entry: e, Via: ViaScope, From: best})
		}
	}
	sort.SliceStable(scoped, func(a, b int) bool {
		return scopeSpecificity(scoped[a].From) > scopeSpecificity(scoped[b].From)
	})
	return append(out, scoped...)
}

// scopeSpecificity ranks a scope pattern: the more literal characters, the
// narrower the code it names.
func scopeSpecificity(scope string) int {
	n := 0
	for _, c := range strings.TrimPrefix(strings.TrimSpace(scope), "./") {
		if c != '*' && c != '?' {
			n++
		}
	}
	return n
}

// LastFailedCheck returns the newest "Check failed" entry of a work log when
// no "ready for check" entry follows it, without the entry's date and actor;
// else "".
func LastFailedCheck(history string) string {
	lines := strings.Split(strings.TrimSpace(history), "\n")
	for k := len(lines) - 1; k >= 0; k-- {
		l := strings.ToLower(lines[k])
		switch {
		case strings.Contains(l, "check failed"):
			return strings.TrimSpace(lines[k][strings.Index(l, "check failed"):])
		case strings.Contains(l, "ready for check"):
			return ""
		}
	}
	return ""
}

// BriefDecision is one decision in a brief. The issue's own decisions carry
// their body; the others are given by title and description.
type BriefDecision struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body,omitempty"`
	Via         string `json:"via"`
	From        string `json:"from,omitempty"`
	Cut         bool   `json:"cut,omitempty"` // the body was cut at the section's bound
}

// Brief holds the parts of a hand-off text; Text renders them.
type Brief struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Kind        Kind            `json:"kind"`
	State       State           `json:"state"`
	Role        string          `json:"role"`
	Task        string          `json:"task"`
	Requirement string          `json:"requirement"`
	Context     string          `json:"context"`
	Criteria    []Criterion     `json:"criteria"`
	Decisions   []BriefDecision `json:"decisions"`
	// DecisionsDropped counts applicable decisions left out at the bound.
	DecisionsDropped int      `json:"decisions_dropped,omitempty"`
	FailedCheck      string   `json:"failed_check,omitempty"`
	DoD              []string `json:"definition_of_done"`
	Rules            []string `json:"rules"`
	ReportSections   []string `json:"report_sections"`
	ReportFormat     string   `json:"report_format"`
	ReportPath       string   `json:"report_path"`
}

// BriefGate reports why an issue cannot be handed off by brief, or nil. A
// brief is given for a code or research issue that is ready or in progress,
// not stale, not a parent, with its dependencies done.
func (t *Tree) BriefGate(id string) error {
	i := t.Issues[id]
	if i == nil {
		return &Error{Code: ErrNotFound, Message: "no issue " + id}
	}
	var unmet []Unmet
	add := func(code, format string, a ...any) { unmet = append(unmet, Unmet{code, fmt.Sprintf(format, a...)}) }
	if i.Kind != KindCode && i.Kind != KindResearch {
		add(GateKind, "a brief is for a code or research issue, this one is %s", i.Kind)
	}
	if t.IsParent(id) {
		add(GateParent, "the issue is a parent; the work is in its children")
	}
	if s := t.State(id); s != StateReady && s != StateInProgress {
		add(GateState, "the issue must be ready or in progress, it is %s", s)
	}
	if t.Stale(id) {
		add(GateStale, "the requirement changed since its baseline; resolve the drift first")
	}
	if deps := t.UnresolvedDeps(id); len(deps) > 0 {
		add(GateDeps, "unresolved dependencies: %s", strings.Join(deps, ", "))
	}
	if len(unmet) == 0 {
		return nil
	}
	return &Error{Code: ErrGate, Message: "no brief for " + id, Unmet: unmet}
}

// BuildBrief assembles the hand-off for an issue: the implementer's, or with
// check the checker's. reportPath is the absolute path of the file the agent
// writes its report to. It fails with BriefGate's error.
func (t *Tree) BuildBrief(id string, check bool, reportPath string) (*Brief, error) {
	if err := t.BriefGate(id); err != nil {
		return nil, err
	}
	i := t.Issues[id]
	b := &Brief{
		ID: id, Title: i.Title, Kind: i.Kind, State: t.State(id), Role: RoleImplementer,
		Requirement: briefRequirement(i.Body),
		Context:     strings.TrimSpace(i.Context),
		Criteria:    append([]Criterion{}, i.Criteria...),
		Decisions:   []BriefDecision{},
		DoD:         briefDoD(t.Project.DoD),
		ReportPath:  reportPath,
	}
	b.Decisions, b.DecisionsDropped = boundDecisions(t.ApplicableDecisions(id), BriefDecisionsBound)
	b.ReportSections = ReportSectionsFor(i.Kind, check, len(i.Criteria) > 0)
	b.ReportFormat = ReportFormat(b.ReportSections)
	switch {
	case check:
		b.Role = RoleChecker
		b.Task = "You are asked to check work that another agent did on the task below: verify it against the requirement and the acceptance criteria, change nothing, and report your verdict."
		b.Rules = checkerRules(i.Kind)
	case i.Kind == KindResearch:
		b.Task = "You are asked to answer the research question below and to report what you found."
		b.Rules = implementerRules(i.Kind)
		b.FailedCheck = LastFailedCheck(i.History)
	default:
		b.Task = "You are asked to implement the change described below in this repository and to report what you did."
		b.Rules = implementerRules(i.Kind)
		b.FailedCheck = LastFailedCheck(i.History)
	}
	return b, nil
}

var prepCommandRe = regexp.MustCompile(`(^|[^.\w])prep [a-z]`)

// briefDoD is the Definition of Done without the items that name a prep
// command (the default "prep check passes"): the agent has no prep and does
// not touch the records, so those stay with the orchestrator.
func briefDoD(dod []string) []string {
	out := []string{}
	for _, d := range dod {
		if !prepCommandRe.MatchString(d) {
			out = append(out, d)
		}
	}
	return out
}

// briefRequirement is the requirement without an empty trailing "Open
// questions" section.
func briefRequirement(body string) string {
	body = strings.TrimSpace(body)
	if k := strings.LastIndex(body, "\n## "); k >= 0 && strings.EqualFold(strings.TrimSpace(body[k+1:]), "## Open questions") {
		body = strings.TrimSpace(body[:k])
	}
	return body
}

// boundDecisions turns the applicable decisions into the brief's section,
// kept to max bytes as Text renders it: in order, each is taken whole while
// it fits; an own decision that does not fit is cut when room is left, and
// whatever else does not fit is dropped and counted.
func boundDecisions(ds []AppliedDecision, max int) (out []BriefDecision, dropped int) {
	out = []BriefDecision{}
	const reserve, minBody = 200, 200 // room for the section's notes; the least body worth showing
	left := max - reserve
	for _, d := range ds {
		bd := BriefDecision{ID: d.Entry.ID, Title: d.Entry.Title, Description: d.Entry.Description, Via: d.Via, From: d.From}
		if d.Via == ViaIssue {
			bd.Body = strings.TrimSpace(d.Entry.Body)
		}
		if cost := len(bd.text()); cost <= left {
			left -= cost
			out = append(out, bd)
			continue
		}
		full := bd.Body
		bd.Body = ""
		if room := left - len(bd.text()) - 80; full != "" && room >= minBody {
			for room > 0 && !utf8.RuneStart(full[room]) {
				room--
			}
			bd.Body = fmt.Sprintf("%s\n[Cut: the first %d of %d characters of this decision are shown.]", strings.TrimSpace(full[:room]), room, len(full))
			bd.Cut = true
			left -= len(bd.text())
			out = append(out, bd)
			continue
		}
		dropped++
	}
	return out, dropped
}

// text renders one decision: an own one under its title with its body, any
// other as one line.
func (d BriefDecision) text() string {
	if d.Via != ViaIssue {
		return "- " + d.Title + ": " + d.Description + "\n"
	}
	s := "### " + d.Title + "\n" + d.Description + "\n"
	if d.Body != "" {
		s += "\n" + d.Body + "\n"
	}
	return s + "\n"
}

// decisionsText renders the decisions section's content, within
// BriefDecisionsBound.
func (b *Brief) decisionsText() string {
	var s strings.Builder
	other := false
	for _, d := range b.Decisions {
		if d.Via != ViaIssue && !other {
			other = true
			s.WriteString("Also in force, by title and summary:\n")
		}
		s.WriteString(d.text())
	}
	if b.DecisionsDropped > 0 {
		if other {
			s.WriteString("\n")
		}
		fmt.Fprintf(&s, "[Left out to keep this section short: %d more that apply.]\n", b.DecisionsDropped)
	}
	return strings.TrimSpace(s.String())
}

// Text renders the brief.
func (b *Brief) Text() string {
	var s strings.Builder
	section := func(title, body string) {
		if strings.TrimSpace(body) != "" {
			s.WriteString("\n## " + title + "\n\n" + strings.TrimSpace(body) + "\n")
		}
	}
	list := func(items []string) string {
		var l strings.Builder
		for _, it := range items {
			l.WriteString("- " + it + "\n")
		}
		return l.String()
	}
	s.WriteString(b.Task + "\n\n# " + b.Title + "\n")
	section("Requirement", b.Requirement)
	section("Implementation context", b.Context)
	var crit strings.Builder
	for n, c := range b.Criteria {
		fmt.Fprintf(&crit, "%d. %s\n", n+1, c.Text)
	}
	section("Acceptance criteria", crit.String())
	section("Decisions that apply", b.decisionsText())
	section("Last failed check", b.FailedCheck)
	section("Definition of Done", list(b.DoD))
	section("Rules", list(b.Rules))
	section("Report", "````markdown\n"+strings.TrimSpace(b.ReportFormat)+"\n````")
	s.WriteString("\nWrite the report to this file: " + b.ReportPath + "\nThen answer with one line saying the report is written. Do not repeat the report in your answer.\n")
	return s.String()
}

// implementerRules are the implementer's rules, worded without any command
// of the tracker: the agent reports, the orchestrator keeps the records.
func implementerRules(kind Kind) []string {
	if kind == KindResearch {
		return []string{
			"Answer exactly the question asked, within the scope the context gives, nothing beyond.",
			"State each finding with how you know it: the file, command, measurement or source. Say what you could not establish. The report must hold your findings.",
			"Change no code unless the context asks for it. Never run git commit or git push.",
			"If the question cannot be answered as asked, or the context conflicts with what you find, stop and report it as a problem.",
			"Do not read or change anything in the `.prep` directory, except to write your report to the file named below.",
		}
	}
	return []string{
		"Implement exactly the requirement and context, nothing beyond.",
		"Run the checks the context and the Definition of Done list, and make them pass.",
		"Never run git commit or git push; leave the change in the working tree.",
		"If the context conflicts with the code or tests, or a criterion cannot be met as written, stop and report the conflict as a problem. Do not bend tests or the requirement to fit.",
		"Do not read or change anything in the `.prep` directory, except to write your report to the file named below.",
		"State per criterion whether it is met, with the evidence. Report the decisions you made while implementing, with the reason.",
	}
}

// checkerRules are the checker's rules, worded without any command of the
// tracker.
func checkerRules(kind Kind) []string {
	verdict := "State per criterion whether it is met, with the evidence you saw yourself. A criterion you could not verify is not met."
	if kind == KindResearch {
		verdict = "Say whether the findings answer the question and rest on evidence you could confirm; state per criterion, where there are any, whether it is met."
	}
	return []string{
		"You are the checker, not the implementer: verify each criterion and the Definition of Done yourself, not from the implementer's account.",
		"Read the change from the worktree or branch you were pointed to, else from the uncommitted diff of this repository.",
		"Run the checks yourself and say which commands you ran.",
		verdict,
		"Change nothing: do not fix what you find, report it as a problem. Never run git commit or git push.",
		"Do not read or change anything in the `.prep` directory, except to write your report to the file named below.",
	}
}
