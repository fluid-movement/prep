package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/results"
	"github.com/fluid-movement/prep/internal/eval/run"
)

func evalCell(c, v string, k, f int) *matrix.CellResult {
	cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: v, Model: "haiku"}}
	for i := 0; i < k+f; i++ {
		cr.Runs = append(cr.Runs, matrix.RunResult{Index: i, Status: run.Completed, Cost: matrix.Cost{USD: 0.01, Calls: 4},
			Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: i < k}}})
	}
	matrix.Reduce(cr)
	return cr
}

func TestEvalReportCompareAndPrune(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	st := results.Store{Root: filepath.Join(dir, "results")}
	for id, cells := range map[string][]*matrix.CellResult{
		"20261012-100000-aaaa": {evalCell("a", "base", 9, 1), evalCell("a", "mutant", 1, 9)},
		"20261013-100000-bbbb": {evalCell("a", "base", 2, 8)},
	} {
		if err := st.Save(&results.Run{ID: id, Created: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Definition: results.Definition{N: 10}, Cells: cells}); err != nil {
			t.Fatal(err)
		}
	}
	out := h.ok("eval", "report", "--dir", dir, "20261012")
	for _, want := range []string{"run 20261012-100000-aaaa", "base / haiku", "9/10 90%", "mutant / haiku"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if out := h.ok("eval", "report", "--dir", dir, "--failures"); !strings.Contains(out, "run 20261013-100000-bbbb") || !strings.Contains(out, "failed 8x: ran guide") {
		t.Errorf("latest with failures:\n%s", out)
	}
	var j struct {
		ID    string `json:"id"`
		Cells []any  `json:"cells"`
	}
	h.jsonOf(&j, "eval", "report", "--dir", dir, "20261012")
	if j.ID != "20261012-100000-aaaa" || len(j.Cells) != 2 {
		t.Errorf("json: %+v", j)
	}
	cmp := h.ok("eval", "compare", "--dir", dir, "20261012", "20261013")
	if !strings.Contains(cmp, "REGRESSED") || !strings.Contains(cmp, "a / base / haiku") || !strings.Contains(cmp, "ONLY IN A") {
		t.Errorf("compare:\n%s", cmp)
	}
	vc := h.ok("eval", "compare", "--dir", dir, "20261012", "--variants", "base,mutant")
	if !strings.Contains(vc, "REGRESSED (1)") {
		t.Errorf("variant compare:\n%s", vc)
	}
	h.fails("usage", "eval", "compare", "--dir", dir, "20261012")
	h.fails("no stored run", "eval", "report", "--dir", dir, "zzz")
	h.fails("unknown eval command", "eval", "bogus")
	if out := h.ok("eval", "prune", "--dir", dir, "--keep", "1"); !strings.Contains(out, "removed 1") {
		t.Errorf("prune:\n%s", out)
	}
	h.fails("no stored run", "eval", "report", "--dir", dir, "20261012")
}
