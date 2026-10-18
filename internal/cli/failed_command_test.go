package cli

import (
	"testing"

	"github.com/fluid-movement/prep/internal/activity"
)

func TestFailedCommandsAreRecorded(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "Failed commands", "--kind", "code", "--by", "agent/1")
	h.fails("no issue matches", "guide", "NOPE", "--by", "agent/1")
	h.fails("E_USAGE", "context", id, "--by", "agent/1")
	var failed []activity.Event
	for _, e := range h.activity().Events {
		if e.Failed {
			failed = append(failed, e)
		}
	}
	if len(failed) != 2 {
		t.Fatalf("failed events: %+v", failed)
	}
	if e := failed[0]; e.Kind != activity.KindPrep || e.Op != "guide" || e.Issue != "" || e.Verb == "" || e.MovesFocus() {
		t.Fatalf("unknown issue: %+v", e)
	}
	if e := failed[1]; e.Kind != activity.KindPrep || e.Op != "context" || e.Issue != id || e.Verb != "E_USAGE" || e.Actor != "agent/1" || e.MovesFocus() {
		t.Fatalf("usage error: %+v", e)
	}
}
