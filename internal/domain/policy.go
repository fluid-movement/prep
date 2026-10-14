package domain

import (
	"fmt"
	"strings"
	"time"
)

// CompletionPolicy records how a project verifies work and reviews knowledge
// impact. It describes the contract, never commands for prep to execute.
// By is attribution, not proof of human approval.
type CompletionPolicy struct {
	Verification  string    `json:"verification" yaml:"verification"`
	Documentation string    `json:"documentation" yaml:"documentation"`
	By            string    `json:"by" yaml:"by"`
	At            time.Time `json:"at" yaml:"at"`
}

const OpPolicy Op = "policy"
const GatePolicy = "G_POLICY"

func (p *CompletionPolicy) Valid() bool {
	return p != nil && strings.TrimSpace(p.Verification) != "" && strings.TrimSpace(p.Documentation) != "" && strings.TrimSpace(p.By) != "" && !p.At.IsZero()
}

// PlanPolicy replaces the complete project policy; callers must supply both
// dimensions so changing one cannot silently discard the other.
func (t *Tree) PlanPolicy(p CompletionPolicy) (*Change, error) {
	p.Verification = strings.TrimSpace(p.Verification)
	p.Documentation = strings.TrimSpace(p.Documentation)
	p.By = strings.TrimSpace(p.By)
	p.At = p.At.UTC().Truncate(time.Second)
	if !p.Valid() {
		return nil, &Error{Code: ErrUsage, Message: "completion policy requires verification, documentation, actor and timestamp", Unmet: []Unmet{{GatePolicy, "describe how work is verified and how documentation impact is reviewed"}}}
	}
	history := append([]CompletionPolicy(nil), t.Project.PolicyHistory...)
	if old := t.Project.Policy; old.Valid() && (old.Verification != p.Verification || old.Documentation != p.Documentation) {
		history = append(history, *old)
	}
	return &Change{Op: OpPolicy, Policy: &p, PolicyHistory: &history}, nil
}

func policyRequired() Unmet {
	return Unmet{GatePolicy, fmt.Sprint("set the project completion policy with prep policy set --verification <method> --documentation <method> before marking work ready")}
}
