package domain

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// VerificationEvidence is a report, not an independently authenticated result.
// Work and config scopes bind it to filesystem content, without commit IDs.
type VerificationEvidence struct {
	ID           string    `json:"id" yaml:"id"`
	Previous     string    `json:"previous,omitempty" yaml:"previous,omitempty"`
	Contract     string    `json:"contract" yaml:"contract"`
	Requirement  string    `json:"requirement" yaml:"requirement"`
	Result       string    `json:"result" yaml:"result"`
	Method       string    `json:"method" yaml:"method"`
	Command      string    `json:"command,omitempty" yaml:"command,omitempty"`
	Observation  string    `json:"observation" yaml:"observation"`
	Source       string    `json:"source,omitempty" yaml:"source,omitempty"`
	Provenance   string    `json:"provenance" yaml:"provenance"`
	By           string    `json:"by" yaml:"by"`
	At           time.Time `json:"at" yaml:"at"`
	Scopes       []string  `json:"scopes" yaml:"scopes"`
	ConfigScopes []string  `json:"config_scopes,omitempty" yaml:"config_scopes,omitempty"`
	Snapshot     string    `json:"snapshot" yaml:"snapshot"`
}

const OpEvidence Op = "evidence"

// LatestEvidence selects the latest report for the bound contract, including
// failures: a previous pass cannot mask a later failed run.
func LatestEvidence(i *Issue, requirement string) *VerificationEvidence {
	if i.Ready == nil || i.Ready.Contract == nil {
		return nil
	}
	for n := len(i.Verification) - 1; n >= 0; n-- {
		e := &i.Verification[n]
		if e.Requirement == requirement && e.Contract == i.Ready.Contract.ID {
			return e
		}
	}
	return nil
}

func (e VerificationEvidence) ContentID() string { e.ID = ""; return contentID(e) }

// ValidateReport checks persisted report fields without claiming that the
// reported observation is true or the supplied scopes cover all affected work.
func (e VerificationEvidence) ValidateReport() error {
	if e.Result != "pass" && e.Result != "fail" {
		return fmt.Errorf("evidence result must be pass or fail")
	}
	if e.Method != "manual" && e.Method != "automated" {
		return fmt.Errorf("evidence method must be manual or automated")
	}
	if strings.TrimSpace(e.Observation) == "" || strings.TrimSpace(e.By) == "" || e.At.IsZero() {
		return fmt.Errorf("evidence needs an observation, actor and timestamp")
	}
	if e.Method == "automated" && strings.TrimSpace(e.Command) == "" {
		return fmt.Errorf("automated evidence needs the command that was run; prep never runs it")
	}
	snapshot, err := hex.DecodeString(e.Snapshot)
	if len(e.Scopes) == 0 || err != nil || len(snapshot) != 32 {
		return fmt.Errorf("evidence needs work scopes and their content fingerprint")
	}
	if e.Provenance != "reported" {
		return fmt.Errorf("submitted evidence is reported; independently verified provenance requires an authenticated source adapter")
	}
	if e.Contract == "" || e.Requirement == "" {
		return fmt.Errorf("evidence needs contract and requirement identities")
	}
	return nil
}

func (t *Tree) PlanEvidence(id string, e VerificationEvidence) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	reject := func(message string) (*Change, error) { return nil, &Error{Code: ErrGate, Message: message} }
	if t.State(id).Terminal() || i.Ready == nil || i.Ready.Contract == nil {
		return reject("evidence requires an unresolved issue with an explicitly bound completion contract; use prep policy adopt <id>")
	}
	if t.PolicyChanged(id) || t.Stale(id) {
		return reject("review and reconcile changed requirements or completion policy before recording evidence")
	}
	found := false
	for _, r := range i.Ready.Contract.Requirements {
		if r.ID == e.Requirement {
			found = true
		}
	}
	if !found {
		return reject("unknown requirement ID; use prep policy show <id>")
	}
	e.Observation = strings.TrimSpace(e.Observation)
	e.Command = strings.TrimSpace(e.Command)
	if e.Provenance != "" && e.Provenance != "reported" {
		return reject("submitted evidence is reported; independently verified provenance requires an authenticated source adapter")
	}
	e.Contract = i.Ready.Contract.ID
	e.Provenance = "reported"
	if err := e.ValidateReport(); err != nil {
		return reject(err.Error())
	}
	e.At = e.At.UTC()
	e.Previous = ""
	if n := len(i.Verification); n > 0 {
		e.Previous = i.Verification[n-1].ID
	}
	e.ID = e.ContentID()
	records := append(append([]VerificationEvidence(nil), i.Verification...), e)
	return &Change{Op: OpEvidence, IssueID: id, Verification: &records, History: fmt.Sprintf("- %s %s: reported %s evidence %s for requirement %s", e.At.Format(time.RFC3339), e.By, e.Result, e.ID, e.Requirement)}, nil
}
