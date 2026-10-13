package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// CompletionContract is the immutable content agreed when work becomes ready.
// Presentation configuration and attribution are intentionally absent: changing
// a theme, recording the same policy again, or committing does not change work.
type CompletionContract struct {
	ID            string                `json:"id" yaml:"id"`
	Verification  string                `json:"verification" yaml:"verification"`
	Documentation string                `json:"documentation" yaml:"documentation"`
	Ancestors     []string              `json:"ancestors,omitempty" yaml:"ancestors,omitempty"`
	Requirements  []ContractRequirement `json:"requirements" yaml:"requirements"`
	OptOuts       []OptOut              `json:"opt_outs,omitempty" yaml:"opt_outs,omitempty"`
}

type ContractRequirement struct {
	ID   string `json:"id" yaml:"id"`
	Kind string `json:"kind" yaml:"kind"`
	Text string `json:"text" yaml:"text"`
}

func contentID(v any) string {
	b, _ := json.Marshal(v) // only concrete strings, slices and structs enter here
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (c CompletionContract) ContentID() string {
	c.ID = ""
	return contentID(c)
}

// Contract computes effective requirements including ancestor contributions.
// Checked criterion state is evidence, not part of the requirement identity.
func (t *Tree) Contract(id string) *CompletionContract {
	p := t.Project.Policy
	if !p.Valid() {
		return nil
	}
	i := t.Issues[id]
	if i == nil {
		return nil
	}
	c := &CompletionContract{Verification: p.Verification, Documentation: p.Documentation, Ancestors: t.Ancestors(id)}
	add := func(kind, text string) {
		r := ContractRequirement{Kind: kind, Text: strings.TrimSpace(text)}
		r.ID = contentID(struct{ Kind, Text string }{r.Kind, r.Text})
		for _, existing := range c.Requirements {
			if existing.ID == r.ID {
				return
			}
		}
		c.Requirements = append(c.Requirements, r)
	}
	dod, opt := t.EffectiveDoD(id)
	c.OptOuts = opt
	add("verification", p.Verification)
	add("documentation", p.Documentation)
	for _, text := range dod {
		add("dod", text)
	}
	for _, criterion := range i.Criteria {
		add("acceptance", criterion.Text)
	}
	c.ID = c.ContentID()
	return c
}

func (t *Tree) PolicyChanged(id string) bool {
	i := t.Issues[id]
	if i == nil || i.Ready == nil || i.Ready.Contract == nil {
		return false
	}
	current := t.Contract(id)
	return current == nil || current.ID != i.Ready.Contract.ID
}

// PlanAdoptPolicy explicitly binds an existing sign-off to the current policy
// without changing its requirement baseline or claiming work. A subsequent
// evidence implementation must key evidence by this contract ID.
func (t *Tree) PlanAdoptPolicy(id string, in Input) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	if i.Ready == nil || t.State(id).Terminal() || !t.ReadyValid(i) || t.Stale(id) {
		return nil, &Error{Code: ErrGate, Message: "policy adoption requires a non-stale unresolved issue with a valid ready record"}
	}
	contract := t.Contract(id)
	if contract == nil {
		return nil, &Error{Code: ErrGate, Message: "set the project completion policy before adopting it", Unmet: []Unmet{policyRequired()}}
	}
	r := *i.Ready
	r.Contract = contract
	return &Change{Op: OpPolicy, IssueID: id, Ready: &r, History: "- " + in.Now.UTC().Format("2006-01-02T15:04:05Z") + " " + in.Actor + ": adopted completion contract " + contract.ID}, nil
}
