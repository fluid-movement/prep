package activity

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Usage sums what request events report: model requests with their
// tokens, the cost, and the newest context fill.
type Usage struct {
	Requests int     `json:"requests"`
	Tokens   Tokens  `json:"tokens"`
	CostUSD  float64 `json:"cost_usd"`
	// ContextUsed and ContextWindow are the newest reported context fill.
	ContextUsed   int `json:"-"`
	ContextWindow int `json:"-"`
}

// Add counts one request event; other kinds are ignored. A turn's end
// reports context and cost without tokens, so only events with tokens
// count as model requests.
func (u *Usage) Add(e Event) {
	if e.Kind != KindRequest && e.Kind != KindSummary {
		return
	}
	if e.Kind == KindRequest && e.Tokens != nil {
		u.Requests++
		u.Tokens.Input += e.Tokens.Input
		u.Tokens.Output += e.Tokens.Output
		u.Tokens.CacheRead += e.Tokens.CacheRead
		u.Tokens.CacheWrite += e.Tokens.CacheWrite
	}
	if e.Context != nil {
		u.ContextUsed, u.ContextWindow = e.Context.Tokens, e.Context.Window
	}
	u.CostUSD += e.CostUSD
}

// Report summarizes recorded sessions: what the hooks did, whether the
// workflow rules were followed, which commands ran and failed, and what
// the harnesses reported as cost. A session is the events of one harness
// session id; a turn runs from one prompt hook event to the next.
type Report struct {
	Since    *time.Time `json:"since,omitempty"`
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Events   int        `json:"events"`
	Sessions int        `json:"sessions"`
	Turns    int        `json:"turns"`
	// Unattributed are events with no session: commands run outside a
	// harness. Gap markers also count toward coverage, without attribution.
	Unattributed int            `json:"events_without_session"`
	Hooks        HookReport     `json:"hooks"`
	Rules        RuleReport     `json:"rules"`
	Handoff      HandoffReport  `json:"handoff"`
	Commands     []CommandCount `json:"commands"`
	Failures     []FailureCount `json:"failures"`
	Cost         CostReport     `json:"cost"`
}

// HookReport counts hook outcomes.
type HookReport struct {
	PromptsSent    int `json:"prompts_sent"`
	PromptsSkipped int `json:"prompts_skipped"`
	// Stop is the finish hook's question, in total and by the number of
	// changed files the event records.
	Stop          StopCount    `json:"stop"`
	StopByFiles   []StopBucket `json:"stop_by_files"`
	DeniedReads   int          `json:"denied_reads"`
	DeniedWrites  int          `json:"denied_writes"`
	EditNotices   int          `json:"edit_notices"`
	Errors        int          `json:"errors"`
	ErrorsByHook  []NameCount  `json:"errors_by_hook"`
	InjectedChars int          `json:"injected_chars"`
}

// StopCount counts Stop questions. Followed are asked questions with a
// prep decide or finding in the same session before the next prompt.
type StopCount struct {
	Asked    int `json:"asked"`
	Skipped  int `json:"skipped"`
	Followed int `json:"followed"`
}

// StopBucket is StopCount for one range of changed files: "<3", "3-8",
// ">8", or "unknown" where the event records no count.
type StopBucket struct {
	Files string `json:"files"`
	StopCount
}

// NameCount is a count under a name.
type NameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// RuleReport gives the rule-following figures per turn, per session and
// per issue. Per turn alone they look worse than they are, because an
// issue read in one turn is worked on in the following ones.
type RuleReport struct {
	Turns    RuleTurns    `json:"turns"`
	Sessions RuleSessions `json:"sessions"`
	Issues   RuleIssues   `json:"issues"`
	// Edits are code edits; EditsNoIssue those with no earlier prep event
	// naming an issue in their session.
	Edits        int `json:"code_edits"`
	EditsNoIssue int `json:"code_edits_without_issue"`
	// Guides are guide reads; GuidesThenShow those followed by a show of
	// the same issue in the same session before its next guide.
	Guides         int `json:"guides"`
	GuidesThenShow int `json:"guides_then_show"`
}

// RuleTurns counts turns that edited code.
type RuleTurns struct {
	Edited int `json:"edited_code"`
	// NoIssue: no prep event named an issue earlier in the session.
	NoIssue int `json:"without_issue"`
	// IssueInTurn and KnowledgeInTurn: an issue was named, or knowledge
	// looked up, in the turn before its first code edit.
	IssueInTurn     int `json:"issue_named_in_turn_before_edit"`
	KnowledgeInTurn int `json:"knowledge_lookup_before_edit"`
	// NoPrep: the turn ran no prep command at all.
	NoPrep int `json:"without_prep_command"`
}

// RuleSessions counts sessions that edited code.
type RuleSessions struct {
	Edited int `json:"edited_code"`
	// NoIssue: at least one code edit came before any issue was named.
	NoIssue int `json:"without_issue"`
	// Knowledge: knowledge was looked up before the first code edit.
	Knowledge      int `json:"knowledge_lookup_before_edit"`
	Guides         int `json:"with_guide"`
	GuidesThenShow int `json:"with_guide_then_show"`
}

// RuleIssues counts issues. A code edit belongs to the issue most recently
// named in its session.
type RuleIssues struct {
	Edited int `json:"edited_code"`
	// Knowledge: knowledge was looked up while the issue was the current
	// one, before its first code edit.
	Knowledge      int `json:"knowledge_lookup_before_edit"`
	Guided         int `json:"guided"`
	GuidesThenShow int `json:"guided_then_show"`
}

// HandoffReport counts the hand-off commands: prep brief and prep handback.
type HandoffReport struct {
	Briefs   int `json:"briefs"`
	Accepted int `json:"handbacks_accepted"`
	Refused  int `json:"handbacks_refused"`
	// RefusedBy are the refusals by error code.
	RefusedBy []NameCount `json:"handbacks_refused_by_reason"`
	// FailedChecks are accepted hand-backs with no later completion of
	// their issue.
	FailedChecks int `json:"failed_checks"`
}

// CommandCount counts one prep command.
type CommandCount struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
	Failed  int    `json:"failed"`
}

// FailureCount counts one command's failures with one error code.
type FailureCount struct {
	Command string `json:"command"`
	Code    string `json:"code"`
	Count   int    `json:"count"`
}

// CostReport is what request events carry. Sessions and subagents with no
// request events get no figure: nothing is estimated for them.
type CostReport struct {
	Usage
	Gaps                 int          `json:"gaps"`
	Coverage             string       `json:"coverage"`
	ByAgent              []AgentUsage `json:"by_agent"`
	Sessions             int          `json:"sessions"`
	SessionsWithRequests int          `json:"sessions_with_requests"`
	// Subagents are the subagents seen in the sessions (events with an
	// agent id), and how many of them have request events.
	Subagents             int          `json:"subagents"`
	SubagentsWithRequests int          `json:"subagents_with_requests"`
	ByActor               []ActorCover `json:"by_actor"`
}

// AgentUsage is scoped to a session; summary costs are never allocated here.
type AgentUsage struct {
	Session   string `json:"session"`
	AgentID   string `json:"agent_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	Role      string `json:"role"`
	Usage
	Gaps int `json:"gaps"`
}

// ActorCover is the request coverage of one actor's sessions.
type ActorCover struct {
	Actor                string `json:"actor"`
	Sessions             int    `json:"sessions"`
	SessionsWithRequests int    `json:"sessions_with_requests"`
}

var (
	stopFiles = regexp.MustCompile(`(\d+) (?:changed |code )?files?\b`)
	editOps   = map[string]bool{"Edit": true, "Write": true, "NotebookEdit": true, "apply_patch": true}
)

func stopBucket(target string) int {
	m := stopFiles.FindStringSubmatch(target)
	if m == nil {
		return 3
	}
	switch n, _ := strconv.Atoi(m[1]); {
	case n < 3:
		return 0
	case n <= 8:
		return 1
	}
	return 2
}

type issueState struct {
	edited, knowledge, guided, shown bool
}

type sessionState struct {
	counted                         bool
	actor                           string
	issue                           string // most recently named
	named                           bool   // some prep event named an issue
	knowledge                       bool   // knowledge was looked up
	edited, noIssue, knowBeforeEdit bool
	guided, guideShown              bool
	requests                        bool
	// The turn in progress.
	turn, turnEdited, turnNamed, turnKnowledge, turnPrep bool
	turnNoIssue, turnIssueFirst, turnKnowFirst           bool
	// stop is the bucket of an asked Stop question still open, else -1.
	stop int
	// issueKnow: knowledge looked up since the current issue was named.
	issueKnow bool
	// pending are guides not yet followed by a show, by issue.
	pending   map[string]int
	subagents map[string]bool // agent id: has request events
}

// BuildReport summarizes events (oldest first). With a non-zero since only
// events from then on are counted, while earlier ones still tell what a
// session had read before. Bridge-disable markers remain relevant for any
// session with events in the reporting period.
func BuildReport(events []Event, since time.Time) Report {
	r := Report{Commands: []CommandCount{}, Failures: []FailureCount{}}
	if !since.IsZero() {
		r.Since = &since
	}
	r.Hooks.StopByFiles = []StopBucket{{Files: "<3"}, {Files: "3-8"}, {Files: ">8"}, {Files: "unknown"}}
	sessions := map[string]*sessionState{}
	var order []*sessionState
	issues := map[string]*issueState{}
	issueOf := func(id string) *issueState {
		if issues[id] == nil {
			issues[id] = &issueState{}
		}
		return issues[id]
	}
	commands := map[string]*CommandCount{}
	failures := map[[2]string]int{}
	hookErrors, refused := map[string]int{}, map[string]int{}
	handbacks := map[string]int{} // issue: accepted hand-backs not yet followed by a completion
	agents := map[[2]string]*AgentUsage{}
	activeSessions := map[string]bool{}
	if !since.IsZero() {
		for _, e := range events {
			if e.Session != "" && !e.At.Before(since) {
				activeSessions[e.Session] = true
			}
		}
	}

	closeTurn := func(s *sessionState) {
		if s.turn {
			r.Turns++
			if s.turnEdited {
				r.Rules.Turns.Edited++
				add(&r.Rules.Turns.NoIssue, s.turnNoIssue)
				add(&r.Rules.Turns.IssueInTurn, s.turnIssueFirst)
				add(&r.Rules.Turns.KnowledgeInTurn, s.turnKnowFirst)
				add(&r.Rules.Turns.NoPrep, !s.turnPrep)
			}
		}
		s.turn, s.turnEdited, s.turnNamed, s.turnKnowledge, s.turnPrep = false, false, false, false, false
		s.turnNoIssue, s.turnIssueFirst, s.turnKnowFirst = false, false, false
	}

	for _, e := range events {
		in := !e.At.Before(since)
		gapIn := e.Kind == KindGap && (in || e.Op == "bridge-disabled" && activeSessions[e.Session])
		missingUsage := in && e.Kind == KindRequest && e.Tokens == nil && e.Context == nil && e.CostUSD == 0
		if gapIn || missingUsage {
			r.Cost.Gaps++
		}
		if in {
			r.Events++
			at := e.At
			if r.From == nil {
				r.From = &at
			}
			r.To = &at
		}
		if (e.Kind == KindPrep || e.Kind == KindFocus) && in {
			op := e.Op
			if e.Kind == KindFocus {
				op = "focus"
			}
			c := commands[op]
			if c == nil {
				c = &CommandCount{Command: op}
				commands[op] = c
			}
			c.Count++
			if e.Failed {
				c.Failed++
				failures[[2]string{e.Op, e.Verb}]++
			}
			switch {
			case e.Op == "brief" && !e.Failed:
				r.Handoff.Briefs++
			case e.Op == "handback" && e.Failed:
				r.Handoff.Refused++
				refused[e.Verb]++
			case e.Op == "handback":
				r.Handoff.Accepted++
				handbacks[e.Issue]++
			case e.Op == "complete" && !e.Failed:
				delete(handbacks, e.Issue)
			}
		}
		if e.Session == "" {
			if in {
				r.Unattributed++
			}
			continue
		}
		s := sessions[e.Session]
		if s == nil {
			s = &sessionState{stop: -1, pending: map[string]int{}, subagents: map[string]bool{}}
			sessions[e.Session] = s
			order = append(order, s)
		}
		if in && !s.counted {
			s.counted = true
			s.actor = e.Actor
		}
		if e.AgentID != "" && in {
			s.subagents[e.AgentID] = s.subagents[e.AgentID] || (e.Kind == KindRequest && e.Tokens != nil)
		}
		if in && e.Kind == KindRequest && e.Tokens != nil || gapIn || missingUsage {
			key := [2]string{e.Session, e.AgentID}
			a := agents[key]
			if a == nil {
				role := "main"
				if e.AgentID != "" {
					role = "subagent"
				}
				a = &AgentUsage{Session: e.Session, AgentID: e.AgentID, Role: role}
				agents[key] = a
			}
			if e.AgentType != "" {
				a.AgentType = e.AgentType
			}
			if e.Tokens == nil {
				a.Gaps++
			} else {
				a.Usage.Add(e)
			}
		}
		switch e.Kind {
		case KindRequest, KindSummary:
			if in {
				r.Cost.Usage.Add(e)
				s.requests = s.requests || e.Kind == KindRequest && e.Tokens != nil
			}
		case KindHook:
			if in {
				r.Hooks.InjectedChars += e.Chars
			}
			if e.Op == "prompt" {
				// Every prompt starts a turn, also one whose hook failed.
				closeTurn(s)
				s.stop = -1
				s.turn = in
			}
			switch {
			case e.Verb == HookError:
				if in {
					r.Hooks.Errors++
					hookErrors[e.Op]++
				}
			case e.Op == "prompt" && e.Verb == "sent":
				add(&r.Hooks.PromptsSent, in)
			case e.Op == "prompt":
				add(&r.Hooks.PromptsSkipped, in)
			case e.Op == "finish" && e.Verb == "asked":
				s.stop = -1
				if in {
					s.stop = stopBucket(e.Target)
					r.Hooks.Stop.Asked++
					r.Hooks.StopByFiles[s.stop].Asked++
				}
			case e.Op == "finish" && e.Verb == "skipped":
				if in {
					r.Hooks.Stop.Skipped++
					r.Hooks.StopByFiles[stopBucket(e.Target)].Skipped++
				}
			case e.Op == "before-tool" && e.Verb == "denied read":
				add(&r.Hooks.DeniedReads, in)
			case e.Op == "before-tool" && e.Verb == "denied write":
				add(&r.Hooks.DeniedWrites, in)
			case e.Op == "before-tool" && e.Verb == "context":
				add(&r.Hooks.EditNotices, in)
			}
		case KindPrep, KindFocus:
			if in {
				s.turn, s.turnPrep = true, true
			}
			if e.Kind == KindPrep && !e.Failed {
				switch {
				case e.Op == "decide" || e.Op == "finding":
					if s.stop >= 0 {
						r.Hooks.Stop.Followed++
						r.Hooks.StopByFiles[s.stop].Followed++
						s.stop = -1
					}
				case strings.HasPrefix(e.Op, "knowledge"):
					s.knowledge, s.turnKnowledge, s.issueKnow = true, true, true
				case e.Op == "guide" && e.Issue != "" && in:
					r.Rules.Guides++
					s.guided = true
					s.pending[e.Issue] = 1
					issueOf(e.Issue).guided = true
				case e.Op == "show" && s.pending[e.Issue] > 0:
					delete(s.pending, e.Issue)
					if in {
						r.Rules.GuidesThenShow++
						s.guideShown = true
						issueOf(e.Issue).shown = true
					}
				}
			}
			if e.Issue != "" {
				if e.Issue != s.issue {
					s.issue, s.issueKnow = e.Issue, false
				}
				s.named, s.turnNamed = true, true
			}
		case KindTool:
			if in {
				s.turn = true
			}
			if !in || !editOps[e.Op] || e.Area != AreaCode {
				break
			}
			r.Rules.Edits++
			if !s.named {
				r.Rules.EditsNoIssue++
				s.noIssue = true
			}
			if !s.edited {
				s.edited, s.knowBeforeEdit = true, s.knowledge
			}
			if !s.turnEdited {
				s.turnEdited = true
				s.turnNoIssue, s.turnIssueFirst, s.turnKnowFirst = !s.named, s.turnNamed, s.turnKnowledge
			}
			if s.issue != "" {
				if i := issueOf(s.issue); !i.edited {
					i.edited, i.knowledge = true, s.issueKnow
				}
			}
		}
	}

	actors := map[string]*ActorCover{}
	for _, s := range order {
		closeTurn(s)
		if !s.counted {
			continue
		}
		r.Sessions++
		if s.edited {
			r.Rules.Sessions.Edited++
			add(&r.Rules.Sessions.NoIssue, s.noIssue)
			add(&r.Rules.Sessions.Knowledge, s.knowBeforeEdit)
		}
		add(&r.Rules.Sessions.Guides, s.guided)
		add(&r.Rules.Sessions.GuidesThenShow, s.guideShown)
		a := actors[s.actor]
		if a == nil {
			a = &ActorCover{Actor: s.actor}
			actors[s.actor] = a
		}
		a.Sessions++
		add(&a.SessionsWithRequests, s.requests)
		add(&r.Cost.SessionsWithRequests, s.requests)
		for _, has := range s.subagents {
			r.Cost.Subagents++
			add(&r.Cost.SubagentsWithRequests, has)
		}
	}
	r.Cost.Sessions = r.Sessions
	r.Cost.Coverage = "unknown"
	if r.Cost.Requests > 0 || r.Cost.Gaps > 0 {
		r.Cost.Coverage = "partial"
	}
	r.Cost.ByAgent = []AgentUsage{}
	for _, a := range agents {
		r.Cost.ByAgent = append(r.Cost.ByAgent, *a)
	}
	sort.Slice(r.Cost.ByAgent, func(i, j int) bool {
		a, b := r.Cost.ByAgent[i], r.Cost.ByAgent[j]
		if a.Session != b.Session {
			return a.Session < b.Session
		}
		return a.AgentID < b.AgentID
	})
	r.Cost.ByActor = []ActorCover{}
	for _, a := range actors {
		r.Cost.ByActor = append(r.Cost.ByActor, *a)
	}
	sort.Slice(r.Cost.ByActor, func(i, j int) bool { return r.Cost.ByActor[i].Actor < r.Cost.ByActor[j].Actor })

	for _, i := range issues {
		add(&r.Rules.Issues.Edited, i.edited)
		add(&r.Rules.Issues.Knowledge, i.edited && i.knowledge)
		add(&r.Rules.Issues.Guided, i.guided)
		add(&r.Rules.Issues.GuidesThenShow, i.shown)
	}
	for _, n := range handbacks {
		r.Handoff.FailedChecks += n
	}
	r.Handoff.RefusedBy = nameCounts(refused)
	r.Hooks.ErrorsByHook = nameCounts(hookErrors)
	for _, c := range commands {
		r.Commands = append(r.Commands, *c)
	}
	sort.Slice(r.Commands, func(i, j int) bool {
		a, b := r.Commands[i], r.Commands[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Command < b.Command
	})
	for k, n := range failures {
		r.Failures = append(r.Failures, FailureCount{Command: k[0], Code: k[1], Count: n})
	}
	sort.Slice(r.Failures, func(i, j int) bool {
		a, b := r.Failures[i], r.Failures[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		if a.Command != b.Command {
			return a.Command < b.Command
		}
		return a.Code < b.Code
	})
	return r
}

func add(n *int, ok bool) {
	if ok {
		*n++
	}
}

func nameCounts(m map[string]int) []NameCount {
	out := []NameCount{}
	for name, n := range m {
		out = append(out, NameCount{name, n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}
