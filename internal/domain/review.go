package domain

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	ReviewNone     = "none"
	ReviewPending  = "pending"
	ReviewStale    = "stale"
	ReviewApproved = "approved"
)

// ResolutionReview is an explicit, append-only human approval of one
// completed resolution and its reported work snapshot. Actor attribution is
// a local trust assertion, not authentication.
type ResolutionReview struct {
	ID           string    `json:"id" yaml:"id"`
	Previous     string    `json:"previous,omitempty" yaml:"previous,omitempty"`
	ResolutionID string    `json:"resolution_id" yaml:"resolution_id"`
	Generation   uint64    `json:"generation" yaml:"generation"`
	Reason       string    `json:"reason" yaml:"reason"`
	By           string    `json:"by" yaml:"by"`
	At           time.Time `json:"at" yaml:"at"`
	Scopes       []string  `json:"scopes" yaml:"scopes"`
	ConfigScopes []string  `json:"config_scopes,omitempty" yaml:"config_scopes,omitempty"`
	Snapshot     string    `json:"snapshot" yaml:"snapshot"`
}

func (r Resolution) ContentID() string { return contentID(r) }

func reviewGeneration(r *Resolution) uint64 {
	if r == nil || r.Generation == 0 {
		return 1 // legacy resolutions predate explicit generations
	}
	return r.Generation
}

func nextResolutionGeneration(i *Issue) uint64 {
	var next uint64 = 1
	for _, r := range i.ResolutionHistory {
		if r.Generation >= next {
			next = r.Generation + 1
		}
	}
	if i.Resolution != nil && i.Resolution.Generation >= next {
		next = i.Resolution.Generation + 1
	}
	return next
}

func latestReviewID(i *Issue) string {
	if len(i.ResolutionReviews) == 0 {
		return ""
	}
	return i.ResolutionReviews[len(i.ResolutionReviews)-1].ID
}

func validFingerprint(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == 32
}

func (r ResolutionReview) ContentID() string { r.ID = ""; return contentID(r) }

func (r ResolutionReview) Validate() error {
	validHash := func(s string) bool {
		b, err := hex.DecodeString(s)
		return err == nil && len(b) == 32
	}
	if !validHash(r.ResolutionID) || r.Generation == 0 {
		return fmt.Errorf("resolution review needs a resolution ID and generation")
	}
	if !humanActor(r.By) || strings.TrimSpace(r.Reason) == "" || r.At.IsZero() {
		return fmt.Errorf("resolution review needs an explicit human:<name> actor, reason and timestamp")
	}
	if len(r.Scopes) == 0 || !validHash(r.Snapshot) {
		return fmt.Errorf("resolution review needs declared scopes and their content fingerprint")
	}
	for _, p := range append(slices.Clone(r.Scopes), r.ConfigScopes...) {
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("resolution review scopes must not be empty")
		}
	}
	if r.ID != r.ContentID() {
		return fmt.Errorf("resolution review content does not match its ID")
	}
	return nil
}

// ResolutionReviewStatus reports the human review of the active done
// resolution. Historical reviews never approve a later generation.
func (t *Tree) ResolutionReviewStatus(id string) string {
	i := t.Issues[id]
	if i == nil || i.Resolution == nil || i.Resolution.Outcome != OutcomeDone {
		return ReviewNone
	}
	resolutionID := i.Resolution.ContentID()
	for n := len(i.ResolutionReviews) - 1; n >= 0; n-- {
		r := &i.ResolutionReviews[n]
		if r.ResolutionID != resolutionID || r.Generation != reviewGeneration(i.Resolution) {
			continue
		}
		if r.Validate() == nil && r.ID == r.ContentID() && i.ResolutionReviewCurrent[r.ID] {
			return ReviewApproved
		}
		return ReviewStale
	}
	if len(i.ResolutionReviews) > 0 {
		// A previous approval exists, but it belongs to an older resolution
		// identity or generation and therefore cannot approve this one.
		return ReviewStale
	}
	return ReviewPending
}

// ReviewWorkScopes returns all reported work scopes selected by a resolution.
// Adapters additionally include the selected knowledge entry content.
func ReviewWorkScopes(r *Resolution) []string {
	var scopes []string
	if r != nil {
		for _, evidence := range r.Verification {
			scopes = append(scopes, evidence.Scopes...)
		}
		for _, approval := range r.Approvals {
			if approval.Kind == ApprovalException {
				scopes = append(scopes, approval.Scopes...)
			}
		}
		for _, review := range r.KnowledgeReviews {
			scopes = append(scopes, review.Scopes...)
		}
	}
	return sortedUnique(scopes)
}

// ReviewConfigScopes returns all selected check-configuration scopes.
func ReviewConfigScopes(r *Resolution) []string {
	var scopes []string
	if r != nil {
		for _, evidence := range r.Verification {
			scopes = append(scopes, evidence.ConfigScopes...)
		}
		for _, approval := range r.Approvals {
			if approval.Kind == ApprovalException {
				scopes = append(scopes, approval.ConfigScopes...)
			}
		}
		for _, review := range r.KnowledgeReviews {
			scopes = append(scopes, review.ConfigScopes...)
		}
	}
	return sortedUnique(scopes)
}

func sortedUnique(values []string) []string {
	slices.Sort(values)
	values = slices.Compact(values)
	return values
}
