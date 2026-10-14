package tui

import "github.com/fluid-movement/prep/internal/domain"

// ReviewSnapshot is the captured content displayed before a human approves.
// The adapter rechecks ResolutionID and Fingerprint immediately before writing.
type ReviewSnapshot struct {
	Issue        *domain.Issue
	Status       string
	ResolutionID string
	WorkScopes   []string
	ConfigScopes []string
	Fingerprint  string
	Evidence     string // informational Git evidence, never part of recorded identity
	Changes      string
	ChangesNote  string
}

type ReviewApprovalRequest struct {
	Snapshot ReviewSnapshot
	Reason   string
}
