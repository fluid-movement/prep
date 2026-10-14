package mdstore

import (
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestVerificationRoundTripAndHistoryIntegrity(t *testing.T) {
	e := domain.VerificationEvidence{Contract: "contract", Requirement: "requirement", Result: "pass",
		Method: "manual", Observation: "Behavior observed", Provenance: "reported", By: "agent/test",
		At: time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC), Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	e.ID = e.ContentID()
	next := e
	next.Previous, next.Result = e.ID, "fail"
	next.ID = next.ContentID()
	raw := renderVerification([]domain.VerificationEvidence{e, next})
	parsed, err := parseVerification(raw)
	if err != nil || len(parsed) != 2 || renderVerification(parsed) != raw {
		t.Fatalf("roundtrip: %v", err)
	}
	if _, err := parseVerification(strings.Replace(raw, "Behavior observed", "Unrecorded change", 1)); err == nil {
		t.Fatal("accepted changed report contents")
	}
	if _, err := parseVerification(renderVerification([]domain.VerificationEvidence{next, e})); err == nil {
		t.Fatal("accepted reordered report history")
	}
	for name, mutate := range map[string]func(*domain.VerificationEvidence){
		"missing observation":    func(e *domain.VerificationEvidence) { e.Observation = "" },
		"invalid method":         func(e *domain.VerificationEvidence) { e.Method = "invented" },
		"missing command":        func(e *domain.VerificationEvidence) { e.Method = "automated" },
		"unsupported provenance": func(e *domain.VerificationEvidence) { e.Provenance = "verified" },
		"invalid snapshot":       func(e *domain.VerificationEvidence) { e.Snapshot = strings.Repeat("z", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			bad := e
			mutate(&bad)
			bad.ID = bad.ContentID()
			if _, err := parseVerification(renderVerification([]domain.VerificationEvidence{bad})); err == nil {
				t.Fatal("accepted malformed report with a valid content ID")
			}
		})
	}
}

func TestResolutionPreservesVerificationAndLegacyRecords(t *testing.T) {
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	c := &domain.CompletionContract{Verification: "Inspect", Documentation: "Review", Requirements: []domain.ContractRequirement{{ID: "requirement", Kind: "acceptance", Text: "Works"}}}
	c.ID = c.ContentID()
	e := domain.VerificationEvidence{Contract: c.ID, Requirement: "requirement", Result: "pass", Method: "manual",
		Observation: "Works", Provenance: "reported", By: "agent/test", At: now, Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	e.ID = e.ContentID()
	r := &domain.Resolution{Outcome: domain.OutcomeDone, By: "agent/test", At: now, Contract: c, Verification: []domain.VerificationEvidence{e}}
	raw := renderResolution(r)
	loaded, err := parseResolution(raw)
	if err != nil || loaded.Contract == nil || len(loaded.Verification) != 1 || renderResolution(loaded) != raw {
		t.Fatalf("resolution roundtrip: %v", err)
	}
	if _, err := parseResolution(strings.Replace(raw, "observation: Works", "observation: Changed", 1)); err == nil {
		t.Fatal("accepted changed resolution verification")
	}
	legacy := renderResolution(&domain.Resolution{Outcome: domain.OutcomeDone, By: "agent/test", At: now})
	loaded, err = parseResolution(legacy)
	if err != nil || loaded.Contract != nil || len(loaded.Verification) != 0 || renderResolution(loaded) != legacy {
		t.Fatalf("legacy resolution fabricated evidence: %v", err)
	}
}
