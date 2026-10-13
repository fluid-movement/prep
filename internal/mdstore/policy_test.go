package mdstore

import (
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestPolicyAndContractCanonicalRoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	p := domain.Project{Schema: 1, ID: "01M4AAAAAAAAAAAAAAAAAAAAAA", Policy: &domain.CompletionPolicy{Verification: "Tests pass", Documentation: "Review docs", By: "test", At: now}}
	raw := renderProject(p, "# Project\n\n## Definition of Done\n\n- Tests pass")
	parsed, err := parseProject(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Policy == nil || *parsed.Policy != *p.Policy || len(parsed.DoD) != 1 {
		t.Fatalf("project: %+v", parsed)
	}
	if canonicalProject(raw) != raw {
		t.Fatal("project canonicalization lost policy")
	}
	if _, err := parseProject(strings.Replace(raw, "verification:", "unknown:", 1)); err == nil {
		t.Fatal("accepted unknown policy field")
	}
	c := &domain.CompletionContract{Verification: "Tests pass", Documentation: "Review docs", Requirements: []domain.ContractRequirement{{ID: "test", Kind: "dod", Text: "Tests pass"}}}
	c.ID = c.ContentID()
	r := &domain.Ready{By: "test", At: now, Baseline: "20260102-030405", Contract: c}
	stored := renderReady(r)
	loaded, err := parseReady(stored)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Contract == nil || loaded.Contract.ContentID() != c.ID || renderReady(loaded) != stored {
		t.Fatal("ready contract roundtrip")
	}
	legacy := renderReady(&domain.Ready{By: "test", At: now, Baseline: r.Baseline})
	loaded, err = parseReady(legacy)
	if err != nil || loaded.Contract != nil {
		t.Fatal("fabricated legacy contract")
	}
	if renderReady(loaded) != legacy {
		t.Fatal("legacy record changed")
	}
}
