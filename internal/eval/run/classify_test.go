package run

import (
	"context"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/transcript"
)

func TestHarnessErrorWithSuccessSubtypeIsErrored(t *testing.T) {
	const limit = "You've hit your session limit"
	tr := &transcript.Transcript{Steps: []transcript.Step{{Tool: "Bash"}}, Final: transcript.Final{Status: "success", IsError: true, Text: limit}}
	if st, msg := classify(context.Background(), nil, tr, ""); st != Errored || msg != limit {
		t.Errorf("session limit: status %s msg %q, want %s %q", st, msg, Errored, limit)
	}
	tr.Final.Status = ""
	if st, _ := classify(context.Background(), nil, tr, ""); st != Errored {
		t.Errorf("empty subtype with is_error: status %s", st)
	}
	ok := &transcript.Transcript{Steps: []transcript.Step{{Tool: "Bash"}}, Final: transcript.Final{Status: "success", Text: "done"}}
	if st, _ := classify(context.Background(), nil, ok, ""); st != Completed {
		t.Errorf("success without is_error: status %s", st)
	}
}
