package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Gate codes of the record write commands.
const (
	GateCriterion = "G_CRITERION"
	GateDecision  = "G_DECISION"
	GateDoD       = "G_DOD"
)

// AcceptanceOp is one change to acceptance criteria or Definition of Done
// additions. Index is 1-based and refers to the criteria before the change.
type AcceptanceOp struct {
	Op     string // check, uncheck, remove, add, dod-add, dod-opt-out, dod-remove
	Index  int
	Text   string
	Reason string
}

// RecordInput carries the content of a record write.
type RecordInput struct {
	Actor string
	Now   time.Time
	Text  string

	// decide
	Title      string
	Supersedes string
	Outcome    bool

	// criterion, dod
	Acceptance []AcceptanceOp
}

// Record writes, one command each.
const (
	OpContext   Op = "context"
	OpFindings  Op = "findings"
	OpDecide    Op = "decide"
	OpCriterion Op = "criterion"
	OpDoD       Op = "dod"
	OpLog       Op = "log"
)

var decisionNumRe = regexp.MustCompile(`^D(\d+)$`)

// PlanRecord validates a write to one of an unresolved issue's records and
// returns the change. It never changes the issue's state.
func (t *Tree) PlanRecord(id string, op Op, in RecordInput) (*Change, error) {
	i := t.Issues[id]
	if s := t.State(id); s.Terminal() {
		return nil, &Error{Code: ErrGate, Message: fmt.Sprintf("cannot %s %s", op, id),
			Unmet: []Unmet{{GateState, fmt.Sprintf("%s requires an unresolved issue, issue is %s", op, s)}}}
	}
	var unmet []Unmet
	add := func(code, format string, a ...any) {
		unmet = append(unmet, Unmet{Code: code, Message: fmt.Sprintf(format, a...)})
	}
	now := in.Now.UTC().Truncate(time.Second)
	c := &Change{Op: op, IssueID: id}
	switch op {
	case OpContext:
		text := in.Text
		c.Context = &text
	case OpFindings:
		if i.Kind != KindResearch {
			add(GateKind, "findings are only for research issues, issue is %s", i.Kind)
		}
		text := in.Text
		c.Findings = &text
	case OpDecide:
		if strings.TrimSpace(in.Title) == "" {
			add(GateTitle, "--title is required")
		}
		if in.Outcome && i.Kind != KindDecision {
			add(GateKind, "--outcome is only for decision issues, issue is %s", i.Kind)
		}
		n := 0
		known := map[string]bool{}
		for _, d := range i.Decisions {
			known[d.ID] = true
			if m := decisionNumRe.FindStringSubmatch(d.ID); m != nil {
				if k, _ := strconv.Atoi(m[1]); k > n {
					n = k
				}
			}
		}
		if in.Supersedes != "" && !known[in.Supersedes] {
			add(GateDecision, "--supersedes %s: no such decision", in.Supersedes)
		}
		c.Decision = &Decision{ID: fmt.Sprintf("D%d", n+1), Title: strings.TrimSpace(in.Title), Date: now.Format("2006-01-02"),
			Supersedes: in.Supersedes, Outcome: in.Outcome, Body: strings.TrimSpace(in.Text)}
	case OpCriterion, OpDoD:
		if len(in.Acceptance) == 0 {
			add(GateCriterion, "nothing to change")
		}
		for _, o := range in.Acceptance {
			switch o.Op {
			case "check", "uncheck", "remove":
				if o.Index < 1 || o.Index > len(i.Criteria) {
					add(GateCriterion, "no criterion %d; prep show %s numbers %d criteria", o.Index, id, len(i.Criteria))
				}
			case "add", "dod-add":
				if strings.TrimSpace(o.Text) == "" {
					add(GateCriterion, "cannot add empty text")
				}
			case "dod-opt-out":
				if strings.TrimSpace(o.Text) == "" || strings.TrimSpace(o.Reason) == "" {
					add(GateDoD, "an opt-out needs the item and a --reason")
				}
			case "dod-remove":
				found := contains(i.DoDAdd, o.Text)
				for _, x := range i.DoDOptOuts {
					found = found || x.Item == o.Text
				}
				if !found {
					add(GateDoD, "%q is neither a Definition of Done addition nor an opt-out of this issue", o.Text)
				}
			}
		}
		c.Acceptance = in.Acceptance
	case OpLog:
		text := strings.Join(strings.Fields(in.Text), " ")
		if text == "" {
			add(GateRequirement, "the log entry is empty")
		}
		c.History = fmt.Sprintf("- %s %s: %s", now.Format(time.RFC3339), in.Actor, text)
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrUsage, Message: fmt.Sprintf("cannot %s %s", op, id), Unmet: unmet}
	}
	return c, nil
}

// ApplyAcceptance returns the criteria and Definition of Done fields after
// the operations: index operations refer to the criteria before the change
// and run first, additions follow in order.
func ApplyAcceptance(i *Issue, ops []AcceptanceOp) ([]Criterion, []string, []OptOut) {
	crit := append([]Criterion(nil), i.Criteria...)
	dodAdd := append([]string(nil), i.DoDAdd...)
	optOuts := append([]OptOut(nil), i.DoDOptOuts...)
	removed := map[int]bool{}
	for _, o := range ops {
		switch o.Op {
		case "check", "uncheck":
			if o.Index >= 1 && o.Index <= len(crit) {
				crit[o.Index-1].Checked = o.Op == "check"
			}
		case "remove":
			removed[o.Index] = true
		}
	}
	var idx []int
	for k := range removed {
		idx = append(idx, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(idx)))
	for _, k := range idx {
		if k >= 1 && k <= len(crit) {
			crit = append(crit[:k-1], crit[k:]...)
		}
	}
	for _, o := range ops {
		switch o.Op {
		case "add":
			crit = append(crit, Criterion{Text: strings.TrimSpace(o.Text)})
		case "dod-add":
			dodAdd = append(dodAdd, strings.TrimSpace(o.Text))
		case "dod-opt-out":
			optOuts = append(optOuts, OptOut{Item: strings.TrimSpace(o.Text), Reason: strings.TrimSpace(o.Reason)})
		case "dod-remove":
			var keepAdd []string
			for _, d := range dodAdd {
				if d != o.Text {
					keepAdd = append(keepAdd, d)
				}
			}
			dodAdd = keepAdd
			var keepOpt []OptOut
			for _, x := range optOuts {
				if x.Item != o.Text {
					keepOpt = append(keepOpt, x)
				}
			}
			optOuts = keepOpt
		}
	}
	return crit, dodAdd, optOuts
}
