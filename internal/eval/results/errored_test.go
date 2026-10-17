package results

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// A harness error (a session limit) is not a pass or a failure: the rate is
// over the 8 finished runs and the report says how many errored.
func TestReportShowsErroredRunsOutsideRate(t *testing.T) {
	r := sample()
	lo, hi := matrix.Wilson(8, 8)
	r.Cells = []*matrix.CellResult{{Cell: matrix.Cell{Case: "a", Variant: "current", Model: "haiku"}, Passed: 8, Errored: 2, Rate: 1, Lo: lo, Hi: hi}}
	out := Report(r, ReportOptions{})
	if !strings.Contains(out, "8/8 100%") || !strings.Contains(out, "2 errored") {
		t.Errorf("report lacks the rate over finished runs or the errored count:\n%s", out)
	}
}
