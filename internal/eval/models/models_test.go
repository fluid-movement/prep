package models

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/run"
)

// cell builds a reduced cell with k passes, f failures of failKind/failDesc
// and a cost per run.
func cell(c, v, m, resolved string, k, f int, usd float64, failKind, failDesc string) *matrix.CellResult {
	cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: v, Model: m}}
	for i := 0; i < k+f; i++ {
		r := matrix.RunResult{Index: i, Status: run.Completed, ModelResolved: resolved, Cost: matrix.Cost{USD: usd, Calls: 5}, Dir: "/runs/" + c + "/" + m}
		if i < k {
			r.Assertions = []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: true}}
		} else {
			r.Assertions = []assert.Result{{Kind: failKind, Desc: failDesc, Pass: false}}
		}
		cr.Runs = append(cr.Runs, r)
	}
	matrix.Reduce(cr)
	return cr
}

var areas = map[string]string{"q1": "queries", "q2": "queries", "d1": "define", "d2": "define", "t1": "tiny"}

func sampleCells() []*matrix.CellResult {
	return []*matrix.CellResult{
		// queries: haiku is enough and cheapest
		cell("q1", "current", "haiku", "claude-haiku-5-5", 10, 0, 0.01, "", ""),
		cell("q2", "current", "haiku", "claude-haiku-5-5", 10, 0, 0.01, "", ""),
		cell("q1", "current", "sonnet", "claude-sonnet-5-5", 10, 0, 0.05, "", ""),
		cell("q2", "current", "sonnet", "claude-sonnet-5-5", 10, 0, 0.05, "", ""),
		// define: haiku reads .prep files and fails the outcome; sonnet is enough
		cell("d1", "current", "haiku", "claude-haiku-5-5", 4, 6, 0.012, "no_prep_files", "never touched .prep files directly"),
		cell("d2", "current", "haiku", "claude-haiku-5-5", 5, 5, 0.012, "issue", "issue titled like gift is defined"),
		cell("d1", "current", "sonnet", "claude-sonnet-5-5", 10, 0, 0.06, "", ""),
		cell("d2", "current", "sonnet", "claude-sonnet-5-5", 10, 0, 0.06, "rubric", "the requirement says why"),
		// tiny: too few runs to say
		cell("t1", "current", "haiku", "claude-haiku-5-5", 2, 0, 0.01, "", ""),
		cell("t1", "current", "sonnet", "claude-sonnet-5-5", 2, 0, 0.05, "", ""),
		// another variant, ignored
		cell("q1", "other", "haiku", "claude-haiku-5-5", 0, 10, 0.01, "ran", "x"),
	}
}

func area(t *testing.T, a *Analysis, name string) Area {
	t.Helper()
	for _, ar := range a.Areas {
		if ar.Area == name {
			return ar
		}
	}
	t.Fatalf("no area %s", name)
	return Area{}
}

func TestDecisions(t *testing.T) {
	a, err := Analyze("r1", "2.1.296", 10, sampleCells(), areas, Options{Variant: "current"})
	if err != nil {
		t.Fatal(err)
	}
	q := area(t, a, "queries")
	if q.Status != Decided || q.Cheapest != "haiku" || q.NextUp != "sonnet" || len(q.WrongBelow) != 0 {
		t.Errorf("queries: %+v", q)
	}
	d := area(t, a, "define")
	if d.Status != Decided || d.Cheapest != "sonnet" || d.NextUp != "" {
		t.Errorf("define: %+v", d)
	}
	if len(d.WrongBelow) != 2 || d.WrongBelow[0].Count != 6 || d.WrongBelow[0].Group != "tool use" || d.WrongBelow[1].Group != "outcome" {
		t.Errorf("define wrong below: %+v", d.WrongBelow)
	}
	var haiku Stat
	for _, s := range d.Models {
		if s.Model == "haiku" {
			haiku = s
		}
	}
	if haiku.Passed != 9 || haiku.Runs != 20 || haiku.Sufficient || len(haiku.Failures[0].Examples) != 1 || haiku.Failures[0].Examples[0] != "/runs/d1/haiku" {
		t.Errorf("haiku define: %+v", haiku)
	}
	if tiny := area(t, a, "tiny"); tiny.Status != NotEnough || !strings.Contains(tiny.Hint, "runs are needed") {
		t.Errorf("tiny: %+v", tiny)
	}
	if len(d.Cases) != 2 || len(d.Cases[0].Models) != 2 {
		t.Errorf("cases: %+v", d.Cases)
	}
}

func TestCostPerPassNotCostPerRun(t *testing.T) {
	// haiku is cheaper per run but sonnet is cheaper per pass is impossible
	// when both are sufficient only if haiku fails; construct both sufficient
	// with haiku costlier per pass because of retries.
	cells := []*matrix.CellResult{
		cell("q1", "v", "haiku", "h", 9, 1, 0.10, "ran", "x"), // 9/10, $0.10 each: $0.111 per pass
		cell("q1", "v", "sonnet", "s", 10, 0, 0.08, "", ""),   // 10/10, $0.08 per pass
		cell("q1", "v", "haiku", "h", 9, 1, 0.10, "ran", "x"),
		cell("q1", "v", "sonnet", "s", 10, 0, 0.08, "", ""),
	}
	// duplicate cells collapse into one area pool; the point is the ranking
	a, err := Analyze("r", "", 10, cells, map[string]string{"q1": "queries"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	q := area(t, a, "queries")
	if q.Cheapest != "sonnet" {
		t.Errorf("cheapest by cost per pass: %+v", q.Models)
	}
}

func TestWarningsAndErrors(t *testing.T) {
	cells := sampleCells()
	cells[0].Resolved = []string{"claude-haiku-5-5", "claude-haiku-5-6"}
	cells[2].Resolved = []string{"claude-sonnet-5-5"}
	a, err := Analyze("r", "", 10, cells, areas, Options{Variant: "current"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Warnings) < 2 || !strings.Contains(strings.Join(a.Warnings, "|"), "claude-haiku-5-6") {
		t.Errorf("warnings: %v", a.Warnings)
	}
	if _, err := Analyze("r", "", 10, cells, areas, Options{Variant: "nope"}); err == nil {
		t.Error("unknown variant accepted")
	}
	// default variant: the first by name
	b, _ := Analyze("r", "", 10, cells, areas, Options{})
	if b.Variant != "current" {
		t.Errorf("default variant %q", b.Variant)
	}
}

func TestNoneSufficient(t *testing.T) {
	cells := []*matrix.CellResult{
		cell("q1", "v", "haiku", "h", 3, 7, 0.01, "ran", "ran guide"),
		cell("q1", "v", "sonnet", "s", 6, 4, 0.05, "ran", "ran guide"),
	}
	a, _ := Analyze("r", "", 10, cells, map[string]string{"q1": "queries"}, Options{})
	q := area(t, a, "queries")
	if q.Status != None || q.Cheapest != "" || q.NextUp != "haiku" || q.Hint != "" {
		t.Errorf("%+v", q)
	}
}

func TestFormats(t *testing.T) {
	a, _ := Analyze("r1", "2.1.296", 10, sampleCells(), areas, Options{Variant: "current"})
	txt := Text(a)
	for _, want := range []string{"model comparison of run r1", "DECISION", "queries", "haiku", "define", "sonnet", "not enough data",
		"cheaper models get wrong: never touched .prep files directly (tool use, 6x)", "BY MODEL AND AREA", "see /runs/d1/haiku", "BY CASE", "q1"} {
		if !strings.Contains(txt, want) {
			t.Errorf("text lacks %q:\n%s", want, txt)
		}
	}
	md := Markdown(a)
	for _, want := range []string{"| Area | Cheapest enough |", "| queries | haiku |", "| define | sonnet |", "`haiku=claude-haiku-5-5`", "Claude Code 2.1.296"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown lacks %q:\n%s", want, md)
		}
	}
}

func TestRunsNeeded(t *testing.T) {
	o := Options{MinRate: 0.9, MinLower: 0.8, MinRuns: 5}
	if n := runsNeeded(1, o); n < 14 || n > 17 {
		t.Errorf("100%% needs %d runs", n)
	}
	if n := runsNeeded(0.95, o); n <= runsNeeded(1, o) {
		t.Errorf("95%% needs %d runs", n)
	}
}
