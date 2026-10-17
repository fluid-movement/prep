package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Gate codes of the record write commands.
const (
	GateCriterion = "G_CRITERION"
)

// AcceptanceOp is one change to acceptance criteria. Index is 1-based and
// refers to the criteria before the change.
type AcceptanceOp struct {
	Op    string // check, uncheck, remove, add
	Index int
	Text  string
}

// RecordInput carries the content of a record write.
type RecordInput struct {
	Actor string
	Now   time.Time
	Text  string

	// criterion
	Acceptance []AcceptanceOp
}

// Record writes, one command each.
const (
	OpContext   Op = "context"
	OpCriterion Op = "criterion"
	OpLog       Op = "log"
)

// PlanRecord validates a write to one of an unresolved issue's records and
// returns the change. It never changes the issue's state.
func (t *Tree) PlanRecord(id string, op Op, in RecordInput) (*Change, error) {
	if s := t.State(id); s.Terminal() {
		return nil, &Error{Code: ErrGate, Message: fmt.Sprintf("cannot %s %s", op, id),
			Unmet: []Unmet{{GateState, fmt.Sprintf("%s requires an unresolved issue, issue is %s", op, s)}}}
	}
	i := t.Issues[id]
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
	case OpCriterion:
		if len(in.Acceptance) == 0 {
			add(GateCriterion, "nothing to change")
		}
		for _, o := range in.Acceptance {
			switch o.Op {
			case "check", "uncheck", "remove":
				if o.Index < 1 || o.Index > len(i.Criteria) {
					add(GateCriterion, "no criterion %d; prep show %s numbers %d criteria", o.Index, id, len(i.Criteria))
				}
			case "add":
				if strings.TrimSpace(o.Text) == "" {
					add(GateCriterion, "cannot add empty text")
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

// ApplyAcceptance returns the criteria after
// the operations: index operations refer to the criteria before the change
// and run first, additions follow in order.
func ApplyAcceptance(i *Issue, ops []AcceptanceOp) []Criterion {
	crit := append([]Criterion(nil), i.Criteria...)
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

		}
	}
	return crit
}
