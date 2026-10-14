package domain

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	OpApproval        Op = "approval"
	GateApproval         = "G_APPROVAL"
	ApprovalException    = "exception"
	ApprovalRetention    = "retain-policy"
	ApprovalOptOut       = "dod-opt-out"
)

// CompletionApproval records an explicit approval action in the local trust
// model. The declared human actor is attribution, not authenticated identity.
type CompletionApproval struct {
	ID              string              `json:"id" yaml:"id"`
	Previous        string              `json:"previous,omitempty" yaml:"previous,omitempty"`
	Kind            string              `json:"kind" yaml:"kind"`
	Contract        string              `json:"contract,omitempty" yaml:"contract,omitempty"`
	PolicyRevision  string              `json:"policy_revision,omitempty" yaml:"policy_revision,omitempty"`
	RequestRevision string              `json:"request_revision,omitempty" yaml:"request_revision,omitempty"`
	Reviewed        *CompletionContract `json:"reviewed,omitempty" yaml:"reviewed,omitempty"`
	Requirement     string              `json:"requirement,omitempty" yaml:"requirement,omitempty"`
	Evidence        string              `json:"evidence,omitempty" yaml:"evidence,omitempty"`
	Item            string              `json:"item,omitempty" yaml:"item,omitempty"`
	OptOutReason    string              `json:"opt_out_reason,omitempty" yaml:"opt_out_reason,omitempty"`
	Reason          string              `json:"reason" yaml:"reason"`
	By              string              `json:"by" yaml:"by"`
	At              time.Time           `json:"at" yaml:"at"`
	Scopes          []string            `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	ConfigScopes    []string            `json:"config_scopes,omitempty" yaml:"config_scopes,omitempty"`
	Snapshot        string              `json:"snapshot,omitempty" yaml:"snapshot,omitempty"`
}

func (a CompletionApproval) ContentID() string { a.ID = ""; return contentID(a) }

func humanActor(actor string) bool {
	return strings.HasPrefix(actor, "human:") && strings.TrimSpace(strings.TrimPrefix(actor, "human:")) != ""
}

func (a CompletionApproval) Validate() error {
	if !humanActor(a.By) || strings.TrimSpace(a.Reason) == "" || a.At.IsZero() {
		return fmt.Errorf("approval needs an explicit human:<name> actor, reason and timestamp")
	}
	switch a.Kind {
	case ApprovalException:
		fp, err := hex.DecodeString(a.Snapshot)
		if a.Contract == "" || a.Requirement == "" || len(a.Scopes) == 0 || err != nil || len(fp) != 32 {
			return fmt.Errorf("exception approval needs a contract, requirement and scoped work fingerprint")
		}
		if a.Reviewed != nil || a.Item != "" || a.OptOutReason != "" || a.PolicyRevision != "" || a.RequestRevision != "" {
			return fmt.Errorf("unexpected exception approval fields")
		}
	case ApprovalRetention:
		if a.Contract == "" || a.PolicyRevision == "" || a.Reviewed == nil || a.Reviewed.ID != a.Reviewed.ContentID() {
			return fmt.Errorf("policy retention approval needs the bound and reviewed contracts")
		}
		if a.Requirement != "" || a.Evidence != "" || a.Item != "" || a.OptOutReason != "" || len(a.Scopes) != 0 || len(a.ConfigScopes) != 0 || a.Snapshot != "" || a.RequestRevision != "" {
			return fmt.Errorf("unexpected policy retention fields")
		}
	case ApprovalOptOut:
		if strings.TrimSpace(a.Item) == "" || strings.TrimSpace(a.OptOutReason) == "" || a.Requirement != optOutID(a.Item, a.OptOutReason) {
			return fmt.Errorf("opt-out approval needs the exact item and requested reason")
		}
		if a.Contract != "" || a.Reviewed != nil || a.Evidence != "" || len(a.Scopes) != 0 || len(a.ConfigScopes) != 0 || a.Snapshot != "" || a.PolicyRevision != "" {
			return fmt.Errorf("unexpected opt-out approval fields")
		}
	default:
		return fmt.Errorf("unknown approval kind %q", a.Kind)
	}
	return nil
}

func optOutID(item, reason string) string {
	return contentID(struct{ Item, Reason string }{item, reason})
}

func latestApproval(i *Issue, kind, requirement string) *CompletionApproval {
	for n := len(i.Approvals) - 1; n >= 0; n-- {
		a := &i.Approvals[n]
		if a.Kind == kind && a.Requirement == requirement {
			return a
		}
	}
	return nil
}

func validApproval(a *CompletionApproval) bool {
	return a != nil && a.Validate() == nil && a.ID == a.ContentID()
}

func approvedOptOut(i *Issue, o OptOut) string {
	a := latestApproval(i, ApprovalOptOut, optOutID(o.Item, o.Reason))
	if validApproval(a) && a.RequestRevision == i.OptOutRevisions[o.Item] {
		return a.ID
	}
	return ""
}

func (t *Tree) unapprovedOptOuts(id string) []OptOut {
	_, opts := t.EffectiveDoD(id)
	var missing []OptOut
	for _, o := range opts {
		if o.ApprovalID == "" {
			missing = append(missing, o)
		}
	}
	return missing
}

// Retaining an older policy cannot retain superseded issue acceptance text,
// ancestry, or opt-out choices. Those need explicit contract adoption.
func retentionCompatible(bound, current *CompletionContract) bool {
	if bound == nil || current == nil || !slices.Equal(bound.Ancestors, current.Ancestors) || !slices.Equal(bound.OptOuts, current.OptOuts) {
		return false
	}
	acceptance := func(c *CompletionContract) []string {
		var ids []string
		for _, r := range c.Requirements {
			if r.Kind == "acceptance" {
				ids = append(ids, r.ID)
			}
		}
		return ids
	}
	return slices.Equal(acceptance(bound), acceptance(current))
}

func (t *Tree) RetentionApproval(id string) *CompletionApproval {
	i := t.Issues[id]
	if i == nil || i.Ready == nil || i.Ready.Contract == nil || t.Stale(id) {
		return nil
	}
	a := latestApproval(i, ApprovalRetention, "")
	current := t.Contract(id)
	if !validApproval(a) || current == nil || a.Contract != i.Ready.Contract.ID || a.Reviewed.ID != current.ID || a.PolicyRevision != t.policyRevision(id) || !retentionCompatible(i.Ready.Contract, current) {
		return nil
	}
	return a
}

func (t *Tree) PolicyReviewRequired(id string) bool {
	return t.PolicyChanged(id) && t.RetentionApproval(id) == nil
}

// ExceptionApproval binds a waiver to the report state the human reviewed as
// well as scoped work. A later failing report requires renewed approval.
func (t *Tree) ExceptionApproval(id, requirement string) *CompletionApproval {
	i := t.Issues[id]
	if i == nil || i.Ready == nil || i.Ready.Contract == nil || t.PolicyReviewRequired(id) || t.Stale(id) {
		return nil
	}
	a := latestApproval(i, ApprovalException, requirement)
	if !validApproval(a) || a.Contract != i.Ready.Contract.ID || !i.ApprovalCurrent[a.ID] {
		return nil
	}
	evidence := ""
	if e := LatestEvidence(i, requirement); e != nil {
		evidence = e.ID
	}
	if a.Evidence != evidence {
		return nil
	}
	return a
}

func (t *Tree) appendApproval(id string, a CompletionApproval) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	if t.State(id).Terminal() {
		return nil, &Error{Code: ErrGate, Message: "approvals require an unresolved issue"}
	}
	a.Reason, a.By = strings.TrimSpace(a.Reason), strings.TrimSpace(a.By)
	a.Scopes = append([]string(nil), a.Scopes...)
	a.ConfigScopes = append([]string(nil), a.ConfigScopes...)
	a.At = a.At.UTC().Truncate(time.Second)
	if err := a.Validate(); err != nil {
		return nil, &Error{Code: ErrGate, Message: err.Error(), Unmet: []Unmet{{GateApproval, err.Error()}}}
	}
	a.Previous = ""
	if n := len(i.Approvals); n > 0 {
		a.Previous = i.Approvals[n-1].ID
	}
	a.ID = a.ContentID()
	records := append(append([]CompletionApproval(nil), i.Approvals...), a)
	return &Change{Op: OpApproval, IssueID: id, Approvals: &records, History: fmt.Sprintf("- %s %s: approved %s %s — %s", a.At.Format(time.RFC3339), a.By, a.Kind, a.ID, a.Reason)}, nil
}

func (t *Tree) PlanException(id string, a CompletionApproval) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	if i.Ready == nil || i.Ready.Contract == nil || !t.ReadyValid(i) || t.PolicyReviewRequired(id) || t.Stale(id) {
		return nil, &Error{Code: ErrGate, Message: "exception approval needs a valid reconciled bound contract"}
	}
	found := false
	for _, r := range i.Ready.Contract.Requirements {
		if r.ID == a.Requirement {
			found = true
		}
	}
	if !found {
		return nil, &Error{Code: ErrGate, Message: "unknown requirement; inspect prep policy show <id>"}
	}
	a.Kind, a.Contract, a.Evidence = ApprovalException, i.Ready.Contract.ID, ""
	if e := LatestEvidence(i, a.Requirement); e != nil {
		a.Evidence = e.ID
	}
	return t.appendApproval(id, a)
}

func (t *Tree) PlanRetainPolicy(id string, in Input) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	if i.Ready == nil || i.Ready.Contract == nil || !t.ReadyValid(i) || t.Stale(id) || !t.PolicyChanged(id) {
		return nil, &Error{Code: ErrGate, Message: "policy retention needs a non-stale unresolved issue with a changed bound policy"}
	}
	current := t.Contract(id)
	if !retentionCompatible(i.Ready.Contract, current) {
		return nil, &Error{Code: ErrGate, Message: "acceptance, ancestry or opt-out choices changed; adopt the current contract instead"}
	}
	return t.appendApproval(id, CompletionApproval{Kind: ApprovalRetention, Contract: i.Ready.Contract.ID, Reviewed: current, PolicyRevision: t.policyRevision(id), Reason: in.Reason, By: in.Actor, At: in.Now})
}

func (t *Tree) PlanApproveOptOut(id, item string, in Input) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	var requested *OptOut
	for _, o := range i.DoDOptOuts {
		if o.Item != item {
			continue
		}
		if requested != nil {
			return nil, &Error{Code: ErrGate, Message: "duplicate opt-outs for this item; remove the duplicate before approving"}
		}
		copy := o
		requested = &copy
	}
	if requested == nil {
		return nil, &Error{Code: ErrGate, Message: "no requested opt-out for this item on this issue"}
	}
	inherited := t.Project.DoD
	if i.Parent != "" {
		inherited, _ = t.EffectiveDoD(i.Parent)
	}
	if !contains(inherited, item) {
		return nil, &Error{Code: ErrGate, Message: "requested opt-out matches no effective inherited requirement"}
	}
	return t.appendApproval(id, CompletionApproval{Kind: ApprovalOptOut, Requirement: optOutID(requested.Item, requested.Reason), RequestRevision: i.OptOutRevisions[item], Item: requested.Item, OptOutReason: requested.Reason, Reason: in.Reason, By: in.Actor, At: in.Now})
}

// policyRevision identifies the particular reviewed change, not just its final
// text. CLI policy-method and issue DoD reverts cannot revive an approval.
// Project DoD is manually edited; unobserved manual reverts have no history head.
func (t *Tree) policyRevision(id string) string {
	type source struct{ ID, Revision string }
	state := struct {
		Verification, Documentation string
		History                     []CompletionPolicy
		DoD                         []string
		Sources                     []source
	}{Verification: t.Project.Policy.Verification, Documentation: t.Project.Policy.Documentation, History: t.Project.PolicyHistory, DoD: t.Project.DoD}
	for _, id := range append(t.Ancestors(id), id) {
		if i := t.Issues[id]; i != nil {
			state.Sources = append(state.Sources, source{i.ID, i.DoDRevision})
		}
	}
	return contentID(state)
}

// DoD revision heads are chained on actual record changes. Keep removed item
// heads too: removing and recreating a request needs renewed human approval.
func (i *Issue) planDoDRevisions(add []string, opts []OptOut) (string, map[string]string, bool) {
	type policy struct {
		Add     []string
		OptOuts []OptOut
	}
	if contentID(policy{i.DoDAdd, i.DoDOptOuts}) == contentID(policy{add, opts}) {
		return "", nil, false
	}
	revision := contentID(struct {
		Previous string
		Policy   policy
	}{i.DoDRevision, policy{add, opts}})
	requests := map[string]string{}
	for item, version := range i.OptOutRevisions {
		requests[item] = version
	}
	items := map[string]bool{}
	for _, o := range append(append([]OptOut(nil), i.DoDOptOuts...), opts...) {
		items[o.Item] = true
	}
	for item := range items {
		var before, after []OptOut
		for _, o := range i.DoDOptOuts {
			if o.Item == item {
				before = append(before, o)
			}
		}
		for _, o := range opts {
			if o.Item == item {
				after = append(after, o)
			}
		}
		if contentID(before) != contentID(after) {
			requests[item] = contentID(struct {
				Previous string
				Request  []OptOut
			}{requests[item], after})
		}
	}
	return revision, requests, true
}

func previousContracts(old *Ready, next *CompletionContract) []CompletionContract {
	if old == nil {
		return nil
	}
	history := append([]CompletionContract(nil), old.PreviousContracts...)
	if old.Contract != nil && (next == nil || old.Contract.ID != next.ID) {
		for _, c := range history {
			if c.ID == old.Contract.ID {
				return history
			}
		}
		history = append(history, *old.Contract)
	}
	return history
}

// CompletionApprovals selects the approvals actually used by a resolution,
// including inherited subtree opt-outs, without copying unrelated history.
func (t *Tree) CompletionApprovals(id string) []CompletionApproval {
	var out []CompletionApproval
	seen := map[string]bool{}
	add := func(a *CompletionApproval) {
		if a != nil && !seen[a.ID] {
			seen[a.ID] = true
			out = append(out, *a)
		}
	}
	add(t.RetentionApproval(id))
	i := t.Issues[id]
	if i.Ready != nil && i.Ready.Contract != nil {
		for _, r := range i.Ready.Contract.Requirements {
			e := LatestEvidence(i, r.ID)
			unchecked := false
			if r.Kind == "acceptance" {
				for _, criterion := range i.Criteria {
					if strings.TrimSpace(criterion.Text) == r.Text && !criterion.Checked {
						unchecked = true
					}
				}
			}
			if e == nil || e.Result != "pass" || !i.EvidenceCurrent[e.ID] || unchecked {
				add(t.ExceptionApproval(id, r.ID))
			}
		}
		for _, source := range append(t.Ancestors(id), id) {
			if t.Issues[source] == nil {
				continue
			}
			for n := range t.Issues[source].Approvals {
				a := &t.Issues[source].Approvals[n]
				for _, o := range i.Ready.Contract.OptOuts {
					if a.ID == o.ApprovalID {
						add(a)
					}
				}
			}
		}
	}
	return out
}
