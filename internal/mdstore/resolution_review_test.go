package mdstore

import (
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestResolutionReviewRecordsRoundTripAndRejectBodyOrBrokenChain(t *testing.T) {
	resolution := domain.Resolution{Outcome: domain.OutcomeDone, Generation: 1, By: "agent/1", At: time.Date(2026, 10, 14, 11, 0, 0, 0, time.UTC), Note: "done"}
	resolutionID := resolution.ContentID()
	review := domain.ResolutionReview{ResolutionID: resolutionID, Generation: 1, Reason: "Reviewed", By: "human:reviewer", At: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC), Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	review.ID = review.ContentID()
	history := renderResolutionHistory([]domain.Resolution{resolution})
	gotHistory, err := parseResolutionHistory(history)
	if err != nil || len(gotHistory) != 1 || gotHistory[0].ContentID() != resolutionID {
		t.Fatalf("history round trip: records=%+v err=%v", gotHistory, err)
	}
	reviews := renderResolutionReviews([]domain.ResolutionReview{review})
	gotReviews, err := parseResolutionReviews(reviews, map[string]uint64{resolutionID: 1})
	if err != nil || len(gotReviews) != 1 || gotReviews[0].ID != review.ID {
		t.Fatalf("reviews round trip: records=%+v err=%v", gotReviews, err)
	}
	if _, err := parseResolutionHistory(history + "unexpected body\n"); err == nil {
		t.Fatal("resolution history body was accepted")
	}
	if _, err := parseResolutionReviews(reviews+"unexpected body\n", map[string]uint64{resolutionID: 1}); err == nil {
		t.Fatal("resolution reviews body was accepted")
	}
	broken := review
	broken.Previous = strings.Repeat("b", 64)
	broken.ID = broken.ContentID()
	if _, err := parseResolutionReviews(renderResolutionReviews([]domain.ResolutionReview{broken}), map[string]uint64{resolutionID: 1}); err == nil {
		t.Fatal("broken review chain was accepted")
	}
	broken = review
	broken.Snapshot = strings.Repeat("b", 64)
	if _, err := parseResolutionReviews(renderResolutionReviews([]domain.ResolutionReview{broken}), map[string]uint64{resolutionID: 1}); err == nil {
		t.Fatal("invalid review content ID was accepted")
	}
}

func TestLegacyResolutionKeepsGenerationAbsent(t *testing.T) {
	raw := "---\noutcome: done\nby: agent/test\nat: 2026-10-14T11:00:00Z\n---\n\nLegacy completion.\n"
	r, err := parseResolution(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Generation != 0 {
		t.Fatalf("legacy generation was fabricated: %d", r.Generation)
	}
	formatted := renderResolution(r)
	if strings.Contains(formatted, "generation:") {
		t.Fatalf("legacy resolution formatting invented a generation:\n%s", formatted)
	}
}
