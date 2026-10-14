package domain

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const OpDependencyOverride Op = "dependency-override"

// DependencyOverride is a human decision to proceed despite one unresolved
// cross-project dependency observation. It is valid only while both the
// source observation and the dependent's requirement snapshot still match.
type DependencyOverride struct {
	ID           string    `json:"id" yaml:"id"`
	Previous     string    `json:"previous,omitempty" yaml:"previous,omitempty"`
	Ref          string    `json:"ref" yaml:"ref"`
	SourceDigest string    `json:"source_digest" yaml:"source_digest"`
	LocalDigest  string    `json:"local_digest" yaml:"local_digest"`
	Reason       string    `json:"reason" yaml:"reason"`
	By           string    `json:"by" yaml:"by"`
	At           time.Time `json:"at" yaml:"at"`
}

func (o DependencyOverride) ContentID() string { o.ID = ""; return contentID(o) }

func (o DependencyOverride) Validate() error {
	_, id, cross := SplitRef(o.Ref)
	source, sourceErr := hex.DecodeString(o.SourceDigest)
	local, localErr := hex.DecodeString(o.LocalDigest)
	if !cross || !ValidID(id) || len(source) != 32 || sourceErr != nil || len(local) != 32 || localErr != nil {
		return fmt.Errorf("dependency override needs a cross-project reference and source/local snapshots")
	}
	if !humanActor(o.By) || strings.TrimSpace(o.Reason) == "" || o.At.IsZero() {
		return fmt.Errorf("dependency override needs a human:<name> actor, reason and timestamp")
	}
	return nil
}

// ValidateDependencyOverrides checks the append-only identity chain.
func ValidateDependencyOverrides(overrides []DependencyOverride) error {
	previous := ""
	seen := map[string]bool{}
	for n, o := range overrides {
		if err := o.Validate(); err != nil {
			return fmt.Errorf("dependency override %d: %w", n+1, err)
		}
		if o.ID != o.ContentID() || o.Previous != previous || seen[o.ID] {
			return fmt.Errorf("dependency override %d has an invalid identity chain", n+1)
		}
		seen[o.ID] = true
		previous = o.ID
	}
	return nil
}

type DependencyOverrideInput struct {
	Ref, Reason, Actor string
	Now                time.Time
}

func (t *Tree) PlanDependencyOverride(id string, in DependencyOverrideInput) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: fmt.Sprintf("issue %s does not exist", id)}
	}
	if !contains(i.DependsOn, in.Ref) {
		return nil, &Error{Code: ErrUsage, Message: "override reference must be an existing dependency"}
	}
	if _, _, cross := SplitRef(in.Ref); !cross {
		return nil, &Error{Code: ErrUsage, Message: "only unknown cross-project dependencies can be overridden"}
	}
	x := t.XIssueOf(in.Ref)
	if t.DependencyOutcome(in.Ref) != DependencyUnknown {
		return nil, &Error{Code: ErrUsage, Message: "dependency is known; overrides apply only while its source is unknown"}
	}
	if !humanActor(in.Actor) {
		return nil, &Error{Code: ErrUsage, Message: "unknown dependency override requires an explicit human:<name> actor"}
	}
	if strings.TrimSpace(in.Reason) == "" {
		return nil, &Error{Code: ErrUsage, Message: "unknown dependency override requires a reason"}
	}
	if err := ValidateDependencyOverrides(i.DependencyOverrides); err != nil {
		return nil, &Error{Code: ErrInvalid, Message: "invalid dependency override history: " + err.Error()}
	}
	if t.DependencyOverrideStatus(id, in.Ref) == "current" {
		return nil, &Error{Code: ErrUsage, Message: "this dependency already has a current human override"}
	}
	o := DependencyOverride{Ref: in.Ref, SourceDigest: x.ObservationDigest(), LocalDigest: DependencyLocalDigest(i, in.Ref),
		Reason: strings.TrimSpace(in.Reason), By: in.Actor, At: in.Now.UTC().Truncate(time.Second)}
	if n := len(i.DependencyOverrides); n > 0 {
		o.Previous = i.DependencyOverrides[n-1].ID
	}
	o.ID = o.ContentID()
	if err := o.Validate(); err != nil {
		return nil, &Error{Code: ErrUsage, Message: "cannot override unknown dependency: " + err.Error()}
	}
	updated := append(append([]DependencyOverride(nil), i.DependencyOverrides...), o)
	return &Change{Op: OpDependencyOverride, IssueID: id, DependencyOverrides: &updated,
		History: fmt.Sprintf("- %s dependency override by %s: %s (%s)", o.At.Format(time.RFC3339), o.By, o.Ref, o.Reason)}, nil
}

// DependencyLocalDigest covers the local requirement and exact dependency
// reference to which a human's unknown-source decision applies.
func DependencyLocalDigest(i *Issue, ref string) string {
	if i == nil {
		return ""
	}
	baseline := ""
	if b := i.LatestBaseline(); b != nil {
		baseline = b.Name + ":" + b.Requirement + ":" + string(b.Kind)
	}
	return contentID(struct {
		ID, Ref, Kind, Body, Baseline string
		DependsOn, DependsOnApproval  []string
	}{i.ID, ref, string(i.Kind), i.Body, baseline, i.DependsOn, i.DependsOnApproval})
}

func (t *Tree) DependencyOutcome(ref string) DependencyOutcome {
	if _, _, cross := SplitRef(ref); !cross {
		if t.Issues[ref] == nil {
			return DependencyUnknown
		}
		if t.State(ref) == StateDone {
			return DependencySatisfied
		}
		return DependencyUnsatisfied
	}
	x := t.XIssueOf(ref)
	if x.Outcome != "" {
		return x.Outcome
	}
	if x.Status != XFound {
		return DependencyUnknown
	}
	if x.State == StateDone {
		return DependencySatisfied
	}
	return DependencyUnsatisfied
}

func (t *Tree) DependencyApprovalStatus(ref string) string {
	if _, _, cross := SplitRef(ref); !cross {
		return t.ResolutionReviewStatus(ref)
	}
	x := t.XIssueOf(ref)
	if x.Status != XFound {
		return ReviewNone
	}
	return x.ReviewStatus
}

// DependencyOverrideStatus reports whether the newest override for ref is
// current, stale, or absent. A stale newest report cannot fall back to an
// older matching report.
func (t *Tree) DependencyOverrideStatus(issueID, ref string) string {
	i := t.Issues[issueID]
	if i == nil {
		return "none"
	}
	var latest *DependencyOverride
	for n := len(i.DependencyOverrides) - 1; n >= 0; n-- {
		if i.DependencyOverrides[n].Ref == ref {
			latest = &i.DependencyOverrides[n]
			break
		}
	}
	if latest == nil {
		return "none"
	}
	if !contains(i.DependsOn, ref) || t.DependencyOutcome(ref) != DependencyUnknown || ValidateDependencyOverrides(i.DependencyOverrides) != nil ||
		latest.ID != latest.ContentID() || latest.SourceDigest != t.XIssueOf(ref).ObservationDigest() || latest.LocalDigest != DependencyLocalDigest(i, ref) {
		return "stale"
	}
	return "current"
}

func (t *Tree) ExternalDependency(issueID, ref string) XIssue {
	x := t.XIssueOf(ref)
	x.OverrideStatus = t.DependencyOverrideStatus(issueID, ref)
	i := t.Issues[issueID]
	if i != nil {
		x.RequiresApproval = contains(i.DependsOnApproval, ref)
		if x.RequiresApproval && x.Outcome == DependencySatisfied && x.ReviewStatus != ReviewApproved {
			x.Outcome = DependencyUnsatisfied
		}
		for n := len(i.DependencyOverrides) - 1; n >= 0; n-- {
			o := i.DependencyOverrides[n]
			if o.Ref == ref {
				x.OverrideReason, x.OverrideBy = o.Reason, o.By
				break
			}
		}
	}
	return x
}
