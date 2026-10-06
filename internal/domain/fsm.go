package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Op is a lifecycle transition, one write command each.
type Op string

const (
	OpDefine   Op = "define"
	OpAck      Op = "ack"
	OpReady    Op = "ready"
	OpClaim    Op = "claim"
	OpRelease  Op = "release"
	OpComplete Op = "complete"
	OpDrop     Op = "drop"
)

// Ops lists transitions in lifecycle order.
var Ops = []Op{OpDefine, OpAck, OpReady, OpClaim, OpRelease, OpComplete, OpDrop}

// Unmet is one gate condition that blocks a transition.
type Unmet struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Gate condition codes.
const (
	GateState         = "G_STATE"
	GateStale         = "G_STALE"
	GateNotStale      = "G_NOT_STALE"
	GateOpenQuestions = "G_OPEN_QUESTIONS"
	GateRequirement   = "G_REQUIREMENT"
	GateTitle         = "G_TITLE"
	GateKind          = "G_KIND"
	GateCriteria      = "G_CRITERIA"
	GateContext       = "G_CONTEXT"
	GateParent        = "G_PARENT"
	GateDeps          = "G_DEPS"
	GateUnchecked     = "G_UNCHECKED"
	GateChildren      = "G_CHILDREN"
	GateEvidence      = "G_EVIDENCE"
	GateFindings      = "G_FINDINGS"
	GateOutcome       = "G_OUTCOME"
	GateDocs          = "G_DOCS"
	GateDocEntry      = "G_DOC_ENTRY"
	GateReason        = "G_REASON"
	GateActor         = "G_ACTOR"
	GateTags          = "G_TAGS"
	GatePriority      = "G_PRIORITY"
	GateConfig        = "G_CONFIG"
)

// Input carries the arguments of a transition. Nil fields in guide mode mean
// "not provided yet"; gates that depend on them are reported as needs.
type Input struct {
	Actor    string
	Now      time.Time
	Note     string
	Reason   string
	Commit   string
	Docs     []string
	NoImpact string
}

// Change is the set of records a transition writes. Storage adapters render it.
type Change struct {
	Op          Op             `json:"op"`
	IssueID     string         `json:"id"`
	NewIssue    *Issue         `json:"-"`
	Baseline    *Baseline      `json:"-"`
	Ready       *Ready         `json:"-"`
	Claim       *Claim         `json:"-"`
	RemoveClaim bool           `json:"-"`
	Resolution  *Resolution    `json:"-"`
	Edit        *IssueEdit     `json:"-"`
	Context     *string        `json:"-"`
	Findings    *string        `json:"-"`
	Decision    *Decision      `json:"-"`
	Acceptance  []AcceptanceOp `json:"-"`
	// Knowledge is a knowledge write; KnowledgeEntry is the entry the
	// adapter rendered from it, used to validate before writing.
	Knowledge      *KnowledgeEdit `json:"-"`
	KnowledgeEntry *Entry         `json:"-"`
	// Config is a new project configuration (settings).
	Config  *Config `json:"-"`
	History string  `json:"-"`
}

// IssueEdit is the new content of an issue's requirement record after
// prep edit. Fields not edited keep their current values.
type IssueEdit struct {
	Title     string
	Kind      Kind
	Parent    string
	DependsOn []string
	Tags      []string
	Priority  Priority
	Body      *string // new requirement text; nil keeps the current one
	Fields    []string
}

// Transition describes one possible transition for guide output.
type Transition struct {
	Op      Op       `json:"op"`
	Command string   `json:"command"`
	Allowed bool     `json:"allowed"`
	Unmet   []Unmet  `json:"unmet,omitempty"`
	Needs   []string `json:"needs,omitempty"`
}

// fromStates lists the states each transition leaves from.
func fromStates(op Op) []State {
	switch op {
	case OpDefine:
		return []State{StateOpen, StateDefined, StateReady}
	case OpAck:
		return []State{StateDefined, StateReady, StateInProgress}
	case OpReady:
		return []State{StateDefined}
	case OpClaim:
		return []State{StateReady}
	case OpRelease:
		return []State{StateInProgress}
	case OpComplete:
		return []State{StateInProgress}
	case OpDrop:
		return []State{StateOpen, StateDefined, StateReady, StateInProgress}
	}
	return nil
}

func hasState(ss []State, s State) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// Applicable reports whether op can leave the issue's current state at all.
func (t *Tree) Applicable(id string, op Op) bool {
	s := t.State(id)
	if op == OpComplete && t.IsParent(id) {
		return s == StateReady || s == StateInProgress
	}
	return hasState(fromStates(op), s)
}

// Gates evaluates every gate condition for a transition. With in == nil,
// conditions that depend on command arguments are returned as needs.
func (t *Tree) Gates(id string, op Op, in *Input) (unmet []Unmet, needs []string) {
	i := t.Issues[id]
	s := t.State(id)
	stale := t.Stale(id)
	parent := t.IsParent(id)
	add := func(code, format string, a ...any) {
		unmet = append(unmet, Unmet{Code: code, Message: fmt.Sprintf(format, a...)})
	}
	if !t.Applicable(id, op) {
		from := fromStates(op)
		if op == OpComplete && parent {
			from = []State{StateReady, StateInProgress}
		}
		add(GateState, "%s requires state %s, issue is %s", op, joinStates(from), s)
	}
	switch op {
	case OpDefine:
		if (s == StateDefined || s == StateReady) && !stale {
			add(GateNotStale, "already defined and the requirement is unchanged since the newest baseline")
		}
		if strings.TrimSpace(i.Title) == "" {
			add(GateTitle, "title is empty")
		}
		if !i.Kind.Valid() {
			add(GateKind, "kind %q is not one of %s", i.Kind, joinKinds())
		}
		if strings.TrimSpace(i.Prose) == "" {
			add(GateRequirement, "requirement prose in issue.md is empty")
		}
		if strings.TrimSpace(i.OpenQuestions) != "" {
			add(GateOpenQuestions, "the Open questions section in issue.md is not empty")
		}
	case OpAck:
		if !stale {
			add(GateStale, "nothing to acknowledge: the requirement and kind match the newest baseline")
		}
		if !i.Kind.Valid() {
			add(GateKind, "kind %q is not one of %s", i.Kind, joinKinds())
		}
		if strings.TrimSpace(i.OpenQuestions) != "" {
			add(GateOpenQuestions, "the Open questions section in issue.md is not empty")
		}
	case OpReady:
		if stale {
			add(GateStale, "the requirement changed since the newest baseline; run prep define (re-enrich) or prep ack (trivial change)")
		}
		if len(i.Criteria) == 0 {
			add(GateCriteria, "acceptance.md has no checkable criteria (- [ ] ...)")
		}
		if !parent && i.Kind != KindManual && strings.TrimSpace(i.Context) == "" {
			add(GateContext, "context.md is empty; %s issues require enrichment context", i.Kind)
		}
	case OpClaim:
		if stale {
			add(GateStale, "the issue is stale; resolve with prep ack or prep define")
		}
		if parent {
			add(GateParent, "parents are not implemented themselves; claim a child instead")
		}
		if deps := t.UnresolvedDeps(id); len(deps) > 0 {
			add(GateDeps, "dependencies not done: %s", strings.Join(deps, ", "))
		}
	case OpRelease:
	case OpComplete:
		if stale {
			add(GateStale, "the issue is stale; resolve with prep ack or prep define before completing")
		}
		if !CheckedAll(i.Criteria) {
			n := 0
			for _, c := range i.Criteria {
				if !c.Checked {
					n++
				}
			}
			add(GateUnchecked, "%d acceptance criteria unchecked in acceptance.md", n)
		}
		if parent {
			var open []string
			for _, c := range t.Children(id) {
				if !t.State(c).Terminal() {
					open = append(open, c)
				}
			}
			if len(open) > 0 {
				add(GateChildren, "children not done or dropped: %s", strings.Join(open, ", "))
			}
		} else {
			switch i.Kind {
			case KindCode:
				if in == nil {
					needs = append(needs, "--commit <ref>")
				} else {
					if in.Commit == "" {
						add(GateEvidence, "code issues need commit evidence: --commit <ref>")
					} else if !commitPattern.MatchString(in.Commit) {
						add(GateEvidence, "commit reference %q is not a hex commit hash", in.Commit)
					}
					if strings.HasPrefix(in.Actor, "human:") {
						add(GateActor, "code issues are completed by an agent, with commit evidence")
					}
				}
			case KindResearch:
				if i.Findings == nil || strings.TrimSpace(*i.Findings) == "" {
					add(GateFindings, "research issues need findings in findings.md")
				}
			case KindDecision:
				if !HasOutcome(i.Decisions) {
					add(GateOutcome, "decision issues need an outcome entry in decisions.md (outcome: true)")
				}
			}
		}
		if in == nil {
			needs = append(needs, "--docs <entry>... or --no-impact <reason>")
		} else {
			switch {
			case len(in.Docs) == 0 && strings.TrimSpace(in.NoImpact) == "":
				add(GateDocs, "a documentation decision is required: --docs <entry> (repeatable) or --no-impact <reason>")
			case len(in.Docs) > 0 && strings.TrimSpace(in.NoImpact) != "":
				add(GateDocs, "use either --docs or --no-impact, not both")
			}
			for _, d := range in.Docs {
				if t.Knowledge[NormalizeEntryPath(d)] == nil {
					add(GateDocEntry, "knowledge entry %s does not exist", NormalizeEntryPath(d))
				}
			}
		}
	case OpDrop:
		if in == nil {
			needs = append(needs, "--reason <text>")
		} else if strings.TrimSpace(in.Reason) == "" {
			add(GateReason, "dropping requires --reason")
		}
	}
	return unmet, needs
}

var commitPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// NormalizeEntryPath turns a knowledge reference into a bundle-relative path.
func NormalizeEntryPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, ".prep/knowledge")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, ".md") {
		p += ".md"
	}
	return p
}

func joinStates(ss []State) string {
	parts := make([]string, len(ss))
	for k, s := range ss {
		parts[k] = string(s)
	}
	return strings.Join(parts, " or ")
}

func joinKinds() string {
	parts := make([]string, len(Kinds))
	for k, s := range Kinds {
		parts[k] = string(s)
	}
	return strings.Join(parts, ", ")
}

// Stamp formats a time as an ID-style timestamp.
func Stamp(t time.Time) string { return t.UTC().Format("20060102-150405") }

// nextStamp returns a timestamp after `after`, starting from now.
func nextStamp(now time.Time, taken func(string) bool) string {
	ts := now.UTC().Truncate(time.Second)
	for taken(Stamp(ts)) {
		ts = ts.Add(time.Second)
	}
	return Stamp(ts)
}

// Plan checks the gates for a transition and returns the records to write.
func (t *Tree) Plan(id string, op Op, in Input) (*Change, error) {
	if unmet, _ := t.Gates(id, op, &in); len(unmet) > 0 {
		return nil, &Error{Code: ErrGate, Message: fmt.Sprintf("cannot %s %s", op, id), Unmet: unmet}
	}
	i := t.Issues[id]
	now := in.Now.UTC().Truncate(time.Second)
	c := &Change{Op: op, IssueID: id}
	switch op {
	case OpDefine, OpAck:
		latest := ""
		if b := i.LatestBaseline(); b != nil {
			latest = b.Name
		}
		name := nextStamp(now, func(s string) bool { return s <= latest })
		c.Baseline = &Baseline{Name: name, By: in.Actor, At: now, Kind: i.Kind, Ack: op == OpAck, Requirement: i.Body}
		if op == OpAck && t.ReadyValid(i) {
			r := *i.Ready
			r.Baseline = name
			c.Ready = &r
		}
	case OpReady:
		c.Ready = &Ready{By: in.Actor, At: now, Baseline: i.LatestBaseline().Name, Note: in.Note}
	case OpClaim:
		c.Claim = &Claim{By: in.Actor, At: now, Note: in.Note}
	case OpRelease:
		c.RemoveClaim = true
		line := fmt.Sprintf("- %s released by %s (claimed by %s at %s)", now.Format(time.RFC3339), in.Actor, i.Claim.By, i.Claim.At.UTC().Format(time.RFC3339))
		if r := strings.TrimSpace(in.Reason); r != "" {
			line += ": " + r
		}
		c.History = line
	case OpComplete:
		dod, opt := t.EffectiveDoD(id)
		res := &Resolution{Outcome: OutcomeDone, By: in.Actor, At: now, DoD: dod, DoDOptOuts: opt, Note: in.Note}
		if !t.IsParent(id) && i.Kind == KindCode {
			res.Evidence = strings.ToLower(in.Commit)
		}
		doc := &Documentation{NoImpact: strings.TrimSpace(in.NoImpact)}
		for _, d := range in.Docs {
			doc.Entries = append(doc.Entries, NormalizeEntryPath(d))
		}
		res.Documentation = doc
		c.Resolution = res
	case OpDrop:
		c.Resolution = &Resolution{Outcome: OutcomeDropped, By: in.Actor, At: now, Reason: strings.TrimSpace(in.Reason), Note: in.Note}
	}
	return c, nil
}

// NewIssueInput holds the fields for creating an issue.
type NewIssueInput struct {
	Title     string
	Kind      Kind
	Parent    string
	DependsOn []string
	Tags      []string
	Priority  string // a level name; empty or medium leaves it unset
	Body      string // requirement prose
}

// PlanNew validates a new issue and returns its creation change. The ID is
// generated from the clock; agents never invent IDs.
func (t *Tree) PlanNew(in NewIssueInput, now time.Time) (*Change, error) {
	var unmet []Unmet
	if strings.TrimSpace(in.Title) == "" {
		unmet = append(unmet, Unmet{GateTitle, "--title is required"})
	}
	if !in.Kind.Valid() {
		unmet = append(unmet, Unmet{GateKind, fmt.Sprintf("--kind must be one of %s", joinKinds())})
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrUsage, Message: "cannot create issue", Unmet: unmet}
	}
	id := nextStamp(now, func(s string) bool { return t.Issues[s] != nil })
	tags, err := NormalizeTags(in.Tags)
	if err != nil {
		return nil, &Error{Code: ErrUsage, Message: "cannot create issue", Unmet: []Unmet{{GateTags, err.Error()}}}
	}
	prio, err := ParsePriority(in.Priority)
	if err != nil {
		return nil, &Error{Code: ErrUsage, Message: "cannot create issue", Unmet: []Unmet{{GatePriority, err.Error()}}}
	}
	issue := &Issue{ID: id, Title: strings.TrimSpace(in.Title), Kind: in.Kind, Parent: in.Parent, DependsOn: in.DependsOn, Tags: tags, Priority: prio, Prose: strings.TrimSpace(in.Body)}
	return &Change{Op: "new", IssueID: id, NewIssue: issue}, nil
}

// EditInput holds the fields prep edit changes. Nil means unchanged.
type EditInput struct {
	Actor     string
	Now       time.Time
	Title     *string
	Kind      *Kind
	Parent    *string
	DependsOn *[]string
	Tags      *[]string
	Priority  *string // a level name; medium or empty unsets it
	Body      *string // requirement text, including its Open questions section
}

// PlanEdit validates an edit of an unresolved issue's title, kind, parent,
// dependencies or requirement and returns the change. Tags and priority are
// metadata: an edit of only those also works on resolved issues. Existence, cycles and
// self-dependencies are left to CheckWrite, which validates the edited tree.
func (t *Tree) PlanEdit(id string, in EditInput) (*Change, error) {
	i := t.Issues[id]
	metaOnly := (in.Tags != nil || in.Priority != nil) && in.Title == nil && in.Kind == nil && in.Parent == nil && in.DependsOn == nil && in.Body == nil
	if s := t.State(id); s.Terminal() && !metaOnly {
		return nil, &Error{Code: ErrGate, Message: fmt.Sprintf("cannot edit %s", id),
			Unmet: []Unmet{{GateState, fmt.Sprintf("edit requires an unresolved issue, issue is %s; only --tag and --priority work on resolved issues", s)}}}
	}
	e := &IssueEdit{Title: i.Title, Kind: i.Kind, Parent: i.Parent, DependsOn: i.DependsOn, Tags: i.Tags, Priority: i.Priority}
	var unmet []Unmet
	if in.Title != nil {
		e.Title = strings.TrimSpace(*in.Title)
		e.Fields = append(e.Fields, "title")
		if e.Title == "" {
			unmet = append(unmet, Unmet{GateTitle, "--title must not be empty"})
		}
	}
	if in.Kind != nil {
		e.Kind = *in.Kind
		e.Fields = append(e.Fields, "kind")
		if !e.Kind.Valid() {
			unmet = append(unmet, Unmet{GateKind, fmt.Sprintf("--kind must be one of %s", joinKinds())})
		}
	}
	if in.Parent != nil {
		e.Parent = *in.Parent
		e.Fields = append(e.Fields, "parent")
	}
	if in.DependsOn != nil {
		e.DependsOn = nil
		for _, d := range *in.DependsOn {
			if contains(e.DependsOn, d) {
				continue
			}
			if t.State(d) == StateDropped {
				unmet = append(unmet, Unmet{GateDeps, fmt.Sprintf("cannot depend on dropped issue %s", d)})
			}
			e.DependsOn = append(e.DependsOn, d)
		}
		e.Fields = append(e.Fields, "depends_on")
	}
	if in.Tags != nil {
		tags, err := NormalizeTags(*in.Tags)
		if err != nil {
			unmet = append(unmet, Unmet{GateTags, err.Error()})
		}
		e.Tags = tags
		e.Fields = append(e.Fields, "tags")
	}
	if in.Priority != nil {
		p, err := ParsePriority(*in.Priority)
		if err != nil {
			unmet = append(unmet, Unmet{GatePriority, err.Error()})
		}
		e.Priority = p
		e.Fields = append(e.Fields, "priority")
	}
	if in.Body != nil {
		b := *in.Body
		e.Body = &b
		e.Fields = append(e.Fields, "requirement")
	}
	if len(e.Fields) == 0 {
		unmet = append(unmet, Unmet{GateRequirement, "nothing to edit: pass --title, --kind, --parent, --depends-on, --tag, --priority, --body or --body-file"})
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrUsage, Message: fmt.Sprintf("cannot edit %s", id), Unmet: unmet}
	}
	now := in.Now.UTC().Truncate(time.Second)
	return &Change{Op: "edit", IssueID: id, Edit: e,
		History: fmt.Sprintf("- %s edited by %s: %s", now.Format(time.RFC3339), in.Actor, strings.Join(e.Fields, ", "))}, nil
}

// Apply returns a new tree with the change applied in memory, used to
// validate a write before it reaches storage.
func (t *Tree) Apply(c *Change) *Tree {
	issues := make([]*Issue, 0, len(t.Issues)+1)
	for _, id := range t.IDs() {
		i := t.Issues[id]
		if id == c.IssueID {
			cp := *i
			cp.Baselines = append([]Baseline(nil), i.Baselines...)
			if c.Baseline != nil {
				cp.Baselines = append(cp.Baselines, *c.Baseline)
				cp.SortBaselines()
			}
			if c.Ready != nil {
				r := *c.Ready
				cp.Ready = &r
			}
			if c.Claim != nil {
				cl := *c.Claim
				cp.Claim = &cl
			}
			if c.RemoveClaim {
				cp.Claim = nil
			}
			if c.Resolution != nil {
				r := *c.Resolution
				cp.Resolution = &r
			}
			if e := c.Edit; e != nil {
				cp.Title, cp.Kind, cp.Parent, cp.DependsOn, cp.Tags, cp.Priority = e.Title, e.Kind, e.Parent, e.DependsOn, e.Tags, e.Priority
				if e.Body != nil {
					cp.Body = *e.Body
				}
			}
			if c.Context != nil {
				cp.Context = *c.Context
			}
			if c.Findings != nil {
				f := *c.Findings
				cp.Findings = &f
				if !cp.HasFile("findings.md") {
					cp.Files = append(append([]string(nil), cp.Files...), "findings.md")
				}
			}
			if c.Decision != nil {
				cp.Decisions = append(append([]Decision(nil), i.Decisions...), *c.Decision)
			}
			if len(c.Acceptance) > 0 {
				cp.Criteria, cp.DoDAdd, cp.DoDOptOuts = ApplyAcceptance(i, c.Acceptance)
			}
			i = &cp
		}
		issues = append(issues, i)
	}
	if c.NewIssue != nil {
		issues = append(issues, c.NewIssue)
	}
	entries := make([]*Entry, 0, len(t.Knowledge)+1)
	for p, e := range t.Knowledge {
		if c.KnowledgeEntry != nil && p == c.KnowledgeEntry.Path {
			continue
		}
		entries = append(entries, e)
	}
	if c.KnowledgeEntry != nil {
		entries = append(entries, c.KnowledgeEntry)
	}
	p := t.Project
	if c.Config != nil {
		p.Config = *c.Config
	}
	return NewTree(p, issues, entries, nil)
}

// CheckWrite validates the tree after a change and rejects the write when it
// introduces new errors.
func (t *Tree) CheckWrite(c *Change) error {
	before := map[string]bool{}
	for _, d := range Validate(t) {
		if d.Severity == SevError {
			before[d.Code+"|"+d.Issue+"|"+d.Message] = true
		}
	}
	var introduced []Diagnostic
	for _, d := range Validate(t.Apply(c)) {
		if d.Severity == SevError && !before[d.Code+"|"+d.Issue+"|"+d.Message] {
			introduced = append(introduced, d)
		}
	}
	if len(introduced) > 0 {
		return &Error{Code: ErrInvalid, Message: "the write would produce invalid state", Diags: introduced}
	}
	return nil
}

var tagRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*$`)

// ValidTag reports whether a tag has the allowed form: lowercase letters,
// digits and . _ - /, starting with a letter or digit.
func ValidTag(tag string) bool { return tagRe.MatchString(tag) }

// NormalizeTags lowercases and deduplicates tags, keeping their order, and
// rejects malformed ones.
func NormalizeTags(in []string) ([]string, error) {
	var out []string
	for _, t := range in {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if !ValidTag(t) {
			return nil, fmt.Errorf("tag %q: use lowercase letters, digits and . _ - /", t)
		}
		if !contains(out, t) {
			out = append(out, t)
		}
	}
	return out, nil
}
