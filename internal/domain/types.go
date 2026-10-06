// Package domain holds prep's model: issues, their records, derived state,
// the lifecycle state machine with its gates, validation and the query
// engine. It knows nothing about files; storage adapters translate between
// their representation and these types.
package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SchemaVersion is the storage schema this binary reads and writes.
const SchemaVersion = 1

// Kind sets the completion rules for leaf issues.
type Kind string

const (
	KindCode     Kind = "code"
	KindManual   Kind = "manual"
	KindResearch Kind = "research"
	KindDecision Kind = "decision"
)

// Kinds is the closed list of issue kinds.
var Kinds = []Kind{KindCode, KindManual, KindResearch, KindDecision}

// Priority orders work: critical before high before medium before low. The
// empty priority is medium, which storage never writes.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// Priorities is the closed list of priority levels, most important first.
var Priorities = []Priority{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow}

// Effective returns the level, medium for an unset priority.
func (p Priority) Effective() Priority {
	if p == "" {
		return PriorityMedium
	}
	return p
}

// Valid reports whether p is unset or one of the levels.
func (p Priority) Valid() bool {
	_, ok := p.level()
	return ok
}

// Rank orders priorities, 0 for critical. Unset is medium, and an unknown
// level ranks with medium so a typo does not hide an issue.
func (p Priority) Rank() int {
	if k, ok := p.level(); ok {
		return k
	}
	return PriorityMedium.Rank()
}

func (p Priority) level() (int, bool) {
	for k, v := range Priorities {
		if p.Effective() == v {
			return k, true
		}
	}
	return 0, false
}

// ParsePriority reads a level from user input; medium and empty both mean unset.
func ParsePriority(s string) (Priority, error) {
	p := Priority(strings.ToLower(strings.TrimSpace(s)))
	if p == "" || p == PriorityMedium {
		return "", nil
	}
	if !p.Valid() {
		return "", fmt.Errorf("priority must be one of %s", joinPriorities())
	}
	return p, nil
}

func joinPriorities() string {
	parts := make([]string, len(Priorities))
	for k, p := range Priorities {
		parts[k] = string(p)
	}
	return strings.Join(parts, ", ")
}

// Valid reports whether k is one of the built-in kinds.
func (k Kind) Valid() bool {
	for _, v := range Kinds {
		if k == v {
			return true
		}
	}
	return false
}

// State is derived from which records exist; it is never stored.
type State string

const (
	StateOpen       State = "open"
	StateDefined    State = "defined"
	StateReady      State = "ready"
	StateInProgress State = "in_progress"
	StateDone       State = "done"
	StateDropped    State = "dropped"
)

// States lists all states in lifecycle order.
var States = []State{StateOpen, StateDefined, StateReady, StateInProgress, StateDone, StateDropped}

// Terminal reports whether the state is done or dropped.
func (s State) Terminal() bool { return s == StateDone || s == StateDropped }

// Project holds project-level facts from project.md and config.yaml.
type Project struct {
	Schema int
	DoD    []string
	Config Config
}

// Config is the project-level configuration.
type Config struct {
	CommitMode string            `json:"commit_mode"`
	Views      map[string]string `json:"views,omitempty"`
	// ViewOrder lists the view names in the order the config file gives them.
	ViewOrder []string `json:"view_order,omitempty"`
}

// Commit modes.
const (
	CommitOff = "off"
	CommitAll = "all"
)

// Criterion is one checkable acceptance criterion.
type Criterion struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// OptOut removes an inherited Definition of Done item for an issue subtree.
type OptOut struct {
	Item   string `json:"item"`
	Reason string `json:"reason"`
}

// Decision is one append-only entry in decisions.md.
type Decision struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Supersedes string `json:"supersedes,omitempty"`
	Outcome    bool   `json:"outcome,omitempty"`
	Body       string `json:"body,omitempty"`
}

// Baseline is a requirement snapshot written by define or ack.
type Baseline struct {
	Name        string    `json:"name"` // timestamp, also the file name
	By          string    `json:"by"`
	At          time.Time `json:"at"`
	Kind        Kind      `json:"kind"`
	Ack         bool      `json:"ack,omitempty"`
	Requirement string    `json:"requirement"`
}

// Ready is the enrichment sign-off.
type Ready struct {
	By       string    `json:"by"`
	At       time.Time `json:"at"`
	Baseline string    `json:"baseline"`
	Note     string    `json:"note,omitempty"`
}

// Claim marks work in progress.
type Claim struct {
	By   string    `json:"by"`
	At   time.Time `json:"at"`
	Note string    `json:"note,omitempty"`
}

// Outcomes of a resolution.
const (
	OutcomeDone    = "done"
	OutcomeDropped = "dropped"
)

// Documentation is the documentation decision made at completion.
type Documentation struct {
	Entries  []string `json:"entries,omitempty"`
	NoImpact string   `json:"no_impact,omitempty"`
}

// Resolution closes an issue as done or dropped.
type Resolution struct {
	Outcome       string         `json:"outcome"`
	By            string         `json:"by"`
	At            time.Time      `json:"at"`
	Reason        string         `json:"reason,omitempty"`
	Evidence      string         `json:"evidence,omitempty"`
	Documentation *Documentation `json:"documentation,omitempty"`
	DoD           []string       `json:"dod,omitempty"`
	DoDOptOuts    []OptOut       `json:"dod_opt_outs,omitempty"`
	Note          string         `json:"note,omitempty"`
}

// Issue is one issue with all of its records, as parsed by a storage adapter.
type Issue struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Kind      Kind     `json:"kind"`
	Parent    string   `json:"parent,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	// Priority as stored: empty means medium. Output carries Effective().
	Priority Priority `json:"-"`

	// Body is the normalized requirement text that baselines snapshot.
	Body string `json:"-"`
	// Prose is the requirement without the open questions section.
	Prose string `json:"requirement"`
	// OpenQuestions is the content of the Open questions section.
	OpenQuestions string `json:"open_questions,omitempty"`

	Criteria   []Criterion `json:"criteria,omitempty"`
	DoDAdd     []string    `json:"dod_add,omitempty"`
	DoDOptOuts []OptOut    `json:"dod_opt_outs,omitempty"`

	Context      string   `json:"-"`
	ContextLinks []string `json:"context_links,omitempty"` // knowledge entries linked from context.md
	ContextPaths []string `json:"context_paths,omitempty"` // code paths mentioned in context.md

	Decisions []Decision `json:"decisions,omitempty"`
	History   string     `json:"-"` // the work log
	Findings  *string    `json:"-"`

	Baselines  []Baseline  `json:"baselines,omitempty"` // oldest first
	Ready      *Ready      `json:"ready,omitempty"`
	Claim      *Claim      `json:"claim,omitempty"`
	Resolution *Resolution `json:"resolution,omitempty"`

	// Files lists the top-level entries present in the issue directory.
	Files []string `json:"files"`
}

// LatestBaseline returns the newest baseline, or nil.
func (i *Issue) LatestBaseline() *Baseline {
	if len(i.Baselines) == 0 {
		return nil
	}
	return &i.Baselines[len(i.Baselines)-1]
}

// HasFile reports whether a top-level file exists in the issue directory.
func (i *Issue) HasFile(name string) bool {
	for _, f := range i.Files {
		if f == name {
			return true
		}
	}
	return false
}

// SortBaselines orders baselines oldest first by name.
func (i *Issue) SortBaselines() {
	sort.Slice(i.Baselines, func(a, b int) bool { return i.Baselines[a].Name < i.Baselines[b].Name })
}

// Entry is one knowledge base entry (OKF concept), as read by the knowledge adapter.
type Entry struct {
	Path            string   `json:"path"` // bundle-relative, e.g. /components/export.md
	Type            string   `json:"type"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Status          string   `json:"status,omitempty"`
	Scope           []string `json:"scope,omitempty"`
	ConfirmedCommit string   `json:"confirmed_commit,omitempty"`
	Links           []string `json:"-"`
	Size            int      `json:"-"`
	// Drifted lists scoped paths changed since ConfirmedCommit, computed by the adapter.
	Drifted []string `json:"drifted,omitempty"`
	// DriftErr is set when drift could not be computed (e.g. unknown commit).
	DriftErr string `json:"-"`
}

// KnowledgeTypes are the allowed OKF entry types.
var KnowledgeTypes = []string{"feature", "component", "decision", "convention", "pitfall", "overview"}

// OverviewEntry is the project overview entry, the last retrieval fallback.
const OverviewEntry = "/overview.md"
