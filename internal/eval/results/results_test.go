package results

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/run"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// cellOf builds a reduced cell with k passes and f failures.
func cellOf(c, v, m string, k, f int) *matrix.CellResult {
	cr := &matrix.CellResult{Cell: matrix.Cell{Case: c, Variant: v, Model: m}, Planned: k + f}
	for i := 0; i < k+f; i++ {
		ok := i < k
		r := matrix.RunResult{Cell: cr.Cell, Index: i, Status: run.Completed, Cost: matrix.Cost{Calls: 5, USD: 0.01},
			Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: ok}}}
		if !ok {
			r.Dir = "/tmp/run-" + c
		}
		cr.Runs = append(cr.Runs, r)
	}
	matrix.Reduce(cr)
	return cr
}

func sample() *Run {
	return &Run{
		ID: "20261012-100000-ab12", Created: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC), Note: "n",
		Definition: Definition{Cases: []string{"a", "b"}, Variants: []VariantRef{{"current", "h1"}}, Models: []string{"haiku"}, N: 10, Seed: 3},
		Versions:   Versions{Prep: "v0.3", Harness: "2.1.296"},
		Cells:      []*matrix.CellResult{cellOf("a", "current", "haiku", 9, 1), cellOf("b", "current", "haiku", 3, 7)},
		SpentUSD:   1.5,
	}
}

func TestStoreRoundTripResolveAndPrune(t *testing.T) {
	s := Store{Root: t.TempDir()}
	if ids, err := s.IDs(); err != nil || len(ids) != 0 {
		t.Fatalf("empty store: %v %v", ids, err)
	}
	if _, err := s.Load("latest"); err == nil {
		t.Error("latest of nothing")
	}
	r := sample()
	if err := s.Save(r); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load(r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Definition.N != 10 || got.Versions.Harness != "2.1.296" || len(got.Cells) != 2 || got.SpentUSD != 1.5 {
		t.Errorf("%+v", got)
	}
	if c := got.Cells[0]; c.Passed != 9 || c.Failed != 1 || c.Rate != 0.9 || len(c.Runs) != 10 {
		t.Errorf("cell not reduced after load: %+v", c)
	}
	for _, id := range []string{"latest", "20261012-1000", r.ID} {
		if _, err := s.Load(id); err != nil {
			t.Errorf("load %q: %v", id, err)
		}
	}
	if _, err := s.Load("nope"); err == nil {
		t.Error("unknown id")
	}
	for i, id := range []string{"20261012-110000-aaaa", "20261012-120000-bbbb", "20261013-090000-cccc"} {
		r2 := sample()
		r2.ID = id
		_ = i
		if err := s.Save(r2); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Load("20261012"); err == nil || !strings.Contains(err.Error(), "matches") {
		t.Errorf("ambiguous prefix: %v", err)
	}
	if l, _ := s.Load("latest"); l.ID != "20261013-090000-cccc" {
		t.Errorf("latest: %s", l.ID)
	}
	removed, err := s.Prune(2)
	if err != nil || len(removed) != 2 {
		t.Fatalf("prune: %v %v", removed, err)
	}
	ids, _ := s.IDs()
	if len(ids) != 2 || ids[0] != "20261013-090000-cccc" {
		t.Errorf("after prune: %v", ids)
	}
	if rm, _ := s.Prune(5); len(rm) != 0 {
		t.Errorf("prune keeps everything when under the limit: %v", rm)
	}
}

func TestNewID(t *testing.T) {
	a, b := NewID(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)), NewID(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if !strings.HasPrefix(a, "20260102-030405-") || a == b {
		t.Errorf("%s %s", a, b)
	}
}

func TestReport(t *testing.T) {
	r := sample()
	r.StoppedBy = "budget"
	out := Report(r, ReportOptions{Failures: true})
	for _, want := range []string{"run 20261012-100000-ab12", "N=10", "prep v0.3", "harness 2.1.296", "spent $1.50", "STOPPED EARLY (budget)",
		"\na\n", "\nb\n", "current / haiku", "9/10 90%", "3/10 30%", "~", "failed 7x: ran guide", "see /tmp/run-b"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	plain := Report(r, ReportOptions{})
	if strings.Contains(plain, "failed 7x") {
		t.Error("failures shown without the option")
	}
}

func TestCompare(t *testing.T) {
	a := []*matrix.CellResult{
		cellOf("same", "v", "m", 10, 0), cellOf("better", "v", "m", 2, 8), cellOf("worse", "v", "m", 10, 0),
		cellOf("noise", "v", "m", 6, 4), cellOf("gone", "v", "m", 5, 5),
	}
	b := []*matrix.CellResult{
		cellOf("same", "v", "m", 10, 0), cellOf("better", "v", "m", 10, 0), cellOf("worse", "v", "m", 2, 8),
		cellOf("noise", "v", "m", 8, 2), cellOf("new", "v", "m", 5, 5),
	}
	got := map[string]Class{}
	for _, d := range Compare(a, b, FullKey) {
		got[d.Key] = d.Class
	}
	want := map[string]Class{"same / v / m": Unchanged, "better / v / m": Improved, "worse / v / m": Regressed, "noise / v / m": Noisy, "gone / v / m": OnlyA, "new / v / m": OnlyB}
	for k, c := range want {
		if got[k] != c {
			t.Errorf("%s: %s, want %s", k, got[k], c)
		}
	}
	out := FormatCompare("run1", "run2", Compare(a, b, FullKey))
	if strings.Index(out, "REGRESSED") > strings.Index(out, "IMPROVED") || !strings.Contains(out, "NOISY (1)") || !strings.Contains(out, "ONLY IN A (1)") {
		t.Errorf("format:\n%s", out)
	}
	if !strings.Contains(FormatCompare("a", "b", nil), "nothing to compare") {
		t.Error("empty compare")
	}
}

func TestCompareVariantsInsideOneRun(t *testing.T) {
	cells := []*matrix.CellResult{cellOf("x", "base", "m", 9, 1), cellOf("x", "mutant", "m", 1, 9), cellOf("y", "base", "m", 5, 5)}
	ds := Compare(ForVariant(cells, "base"), ForVariant(cells, "mutant"), CaseModelKey)
	got := map[string]Class{}
	for _, d := range ds {
		got[d.Key] = d.Class
	}
	if got["x / m"] != Regressed || got["y / m"] != OnlyA {
		t.Errorf("%v", got)
	}
}

func TestBaseline(t *testing.T) {
	run := &Run{ID: "r1", Created: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Cells: []*matrix.CellResult{
		cellOf("fine", "v", "m", 10, 0), cellOf("slipped", "v", "m", 10, 0), cellOf("noise", "v", "m", 10, 0), cellOf("dropped-from-run", "v", "m", 10, 0),
	}}
	path := t.TempDir() + "/baseline.json"
	b := NewBaseline(run, "smoke")
	if err := SaveBaseline(path, b); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBaseline(path)
	if err != nil || len(loaded.Cells) != 4 || loaded.Note != "smoke" || loaded.Cells[0].Case != "dropped-from-run" {
		t.Fatalf("round trip: %v %+v", err, loaded)
	}
	now := &Run{ID: "r2", Cells: []*matrix.CellResult{
		cellOf("fine", "v", "m", 10, 0), cellOf("slipped", "v", "m", 1, 9), cellOf("noise", "v", "m", 7, 3),
	}}
	chk := CheckBaseline(now, loaded)
	if chk.Checked != 3 || len(chk.Regressions) != 1 || chk.Regressions[0].Cell.Case != "slipped" || len(chk.Missing) != 1 {
		t.Errorf("%+v", chk)
	}
	out := chk.Format()
	if !strings.Contains(out, "REGRESSED slipped / v / m: baseline 10/10 100%") || !strings.Contains(out, "not covered: dropped-from-run") {
		t.Errorf("format:\n%s", out)
	}
	if _, err := LoadBaseline(t.TempDir() + "/none.json"); err == nil {
		t.Error("missing file")
	}
}

func TestMerge(t *testing.T) {
	a := &Run{ID: "a", Created: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), SpentUSD: 1, Definition: Definition{N: 5, Models: []string{"haiku"}, Areas: map[string]string{"c1": "x"}},
		Cells: []*matrix.CellResult{cellOf("c1", "v", "haiku", 4, 1), cellOf("c2", "v", "haiku", 5, 0)}}
	b := &Run{ID: "b", Created: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), SpentUSD: 2, Definition: Definition{N: 10, Models: []string{"haiku", "opus"}, Areas: map[string]string{"c2": "y"}},
		Cells: []*matrix.CellResult{cellOf("c1", "v", "haiku", 5, 0), cellOf("c1", "v", "opus", 5, 0)}}
	m := Merge(a, b)
	if m.ID != "a+b" || m.SpentUSD != 3 || m.Definition.N != 10 || !m.Created.Equal(b.Created) || len(m.Definition.Models) != 2 || m.Definition.Areas["c1"] != "x" || m.Definition.Areas["c2"] != "y" {
		t.Fatalf("%+v", m)
	}
	got := map[matrix.Cell]*matrix.CellResult{}
	for _, c := range m.Cells {
		got[c.Cell] = c
	}
	pooled := got[matrix.Cell{Case: "c1", Variant: "v", Model: "haiku"}]
	if pooled == nil || pooled.Passed != 9 || pooled.Failed != 1 || len(pooled.Runs) != 10 || pooled.Runs[9].Index != 9 {
		t.Errorf("pooled cell: %+v", pooled)
	}
	if len(m.Cells) != 3 || got[matrix.Cell{Case: "c1", Variant: "v", Model: "opus"}] == nil {
		t.Errorf("cells: %d", len(m.Cells))
	}
	if Merge(a) != a {
		t.Error("a single run merges to itself")
	}
}

func writeTranscript(t *testing.T, dir, caseName, n string, steps []transcript.Step) {
	t.Helper()
	p := filepath.Join(dir, "runs", caseName, "current", "sonnet", n)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(transcript.Transcript{Steps: steps})
	if err := os.WriteFile(filepath.Join(p, "transcript.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestErrors(t *testing.T) {
	dir := t.TempDir()
	list := transcript.ParseShell("prep list --text x --max 5")
	show := transcript.ParseShell("prep show 01KE6P56E0ABYZR1S1G9JMY5HZ")
	failList := transcript.Step{Tool: "Bash", Commands: list, Error: true, Result: "Exit code 2\nprep: unknown query flag --max [E_USAGE]"}
	writeTranscript(t, dir, "a", "0", []transcript.Step{failList, {Tool: "Bash", Commands: show, Result: "ok"}})
	writeTranscript(t, dir, "a", "1", []transcript.Step{failList})
	writeTranscript(t, dir, "b", "0", []transcript.Step{failList,
		{Tool: "Bash", Commands: show, Error: true, Result: "prep: no issue matches \"01KE6P56E0ABYZR1S1G9JMY5HZ\" [E_NOT_FOUND]"}, {Tool: "Read"}, {Tool: "Read"}})
	rep, err := Errors(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Sessions != 3 || rep.PrepCalls != 5 || rep.Failed != 4 {
		t.Errorf("counts: %+v", rep)
	}
	if len(rep.Groups) != 2 || rep.Groups[0].Count != 3 || rep.Groups[0].Command != "list" || rep.Groups[0].Message != "prep: unknown query flag --max [E_USAGE]" ||
		strings.Join(rep.Groups[0].Cases, ",") != "a,b" {
		t.Errorf("groups: %+v", rep.Groups)
	}
	if !strings.Contains(rep.Groups[1].Message, `no issue matches "<id>"`) {
		t.Errorf("ids are not normalized: %q", rep.Groups[1].Message)
	}
	if rep.ByCommand["list"] != [2]int{3, 3} || rep.ByCommand["show"] != [2]int{2, 1} {
		t.Errorf("by command: %v", rep.ByCommand)
	}
	if rep.MeanCalls[0].Case != "b" || rep.MeanCalls[0].Mean != 4 {
		t.Errorf("mean calls: %+v", rep.MeanCalls)
	}
	out := FormatErrors(rep)
	if !strings.Contains(out, "3x prep list") || !strings.Contains(out, "e.g. prep list --text x --max 5") {
		t.Errorf("format:\n%s", out)
	}
	empty, _ := Errors(t.TempDir())
	if !strings.Contains(FormatErrors(empty), "no prep command failed") {
		t.Error("empty report")
	}
}
