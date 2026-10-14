package mdstore

import (
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestKnowledgeReviewRecordRoundTripsAndValidatesChain(t *testing.T) {
	r := domain.KnowledgeReview{Contract: "contract", Kind: domain.KnowledgeUnchanged, Entry: "/overview.md", EntryRevision: strings.Repeat("a", 64), Observation: "Reviewed current guidance", Provenance: domain.KnowledgeProvenance, By: "agent/test", At: time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC), Scopes: []string{"internal/domain"}, Snapshot: strings.Repeat("b", 64)}
	r.ID = r.ContentID()
	raw := renderKnowledgeReviews([]domain.KnowledgeReview{r})
	got, err := parseKnowledgeReviews(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != r.ID || renderKnowledgeReviews(got) != raw {
		t.Fatalf("round trip mismatch: %#v", got)
	}
	got[0].Observation = "tampered"
	if _, err := parseKnowledgeReviews(renderKnowledgeReviews(got)); err == nil {
		t.Fatal("accepted tampered review")
	}
}

func TestLegacyResolutionWithoutKnowledgeSnapshotStillReads(t *testing.T) {
	raw := renderResolution(&domain.Resolution{Outcome: domain.OutcomeDone, By: "agent/test", At: time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)})
	if strings.Contains(raw, "knowledge_reviews:") {
		t.Fatalf("empty new field should be omitted: %s", raw)
	}
	if _, err := parseResolution(raw); err != nil {
		t.Fatal(err)
	}
}
