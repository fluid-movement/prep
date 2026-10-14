package mdstore

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

type approvalsFM struct {
	Approvals       []domain.CompletionApproval `yaml:"approvals"`
	DoDRevision     string                      `yaml:"dod_revision,omitempty"`
	OptOutRevisions map[string]string           `yaml:"opt_out_revisions,omitempty"`
}

func parseApprovalState(raw string) (approvalsFM, error) {
	var f approvalsFM
	fm, body, ok, err := splitFrontmatter(raw)
	if err != nil {
		return f, err
	}
	if !ok || strings.TrimSpace(body) != "" {
		return f, fmt.Errorf("approvals.md requires approval frontmatter and no body")
	}
	if err := decodeStrict(fm, &f); err != nil {
		return f, err
	}
	validRevision := func(s string) bool { b, err := hex.DecodeString(s); return err == nil && len(b) == 32 }
	if f.DoDRevision != "" && !validRevision(f.DoDRevision) {
		return f, fmt.Errorf("invalid DoD revision")
	}
	for item, revision := range f.OptOutRevisions {
		if strings.TrimSpace(item) == "" || !validRevision(revision) {
			return f, fmt.Errorf("invalid opt-out request revision")
		}
	}
	previous := ""
	for _, a := range f.Approvals {
		if a.ID != a.ContentID() || a.Previous != previous {
			return f, fmt.Errorf("invalid approval content hash or history chain")
		}
		if err := a.Validate(); err != nil {
			return f, err
		}
		previous = a.ID
	}
	return f, nil
}

func parseApprovals(raw string) ([]domain.CompletionApproval, error) {
	f, err := parseApprovalState(raw)
	return f.Approvals, err
}

func renderApprovalState(f approvalsFM) string { return withFrontmatter(encodeYAML(f), "") }

func renderApprovals(records []domain.CompletionApproval) string {
	return renderApprovalState(approvalsFM{Approvals: records})
}
