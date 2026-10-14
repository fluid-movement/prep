package mdstore

import (
	"fmt"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

type verificationFM struct {
	Evidence []domain.VerificationEvidence `yaml:"evidence"`
}

func parseVerification(raw string) ([]domain.VerificationEvidence, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err != nil {
		return nil, err
	}
	if !ok || strings.TrimSpace(body) != "" {
		return nil, fmt.Errorf("verification.md requires evidence frontmatter and no body")
	}
	var f verificationFM
	if err := decodeStrict(fm, &f); err != nil {
		return nil, err
	}
	previous := ""
	for _, e := range f.Evidence {
		if e.ID != e.ContentID() || e.Previous != previous {
			return nil, fmt.Errorf("invalid evidence content hash or history chain")
		}
		if err := e.ValidateReport(); err != nil {
			return nil, err
		}
		previous = e.ID
	}
	return f.Evidence, nil
}
func renderVerification(records []domain.VerificationEvidence) string {
	return withFrontmatter(encodeYAML(verificationFM{Evidence: records}), "")
}
