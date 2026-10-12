package matrix

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	_ "github.com/fluid-movement/prep/internal/eval/assert/outcome"
	_ "github.com/fluid-movement/prep/internal/eval/assert/tools"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/run"
	"github.com/fluid-movement/prep/internal/eval/variant"
)

var prepBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "matrix-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	prepBin = filepath.Join(dir, "prep")
	if out, err := exec.Command("go", "build", "-o", prepBin, "../../../cmd/prep").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build prep: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func mkCase(t *testing.T, name, extra string) *evalcase.Case {
	t.Helper()
	c, err := evalcase.Parse(strings.NewReader("name: "+name+"\nprompt: p\nspec: {issues: []}\n"+extra), name+".yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// script returns results from a function of the cell and the run index.
type script func(c Cell, i int) RunResult

func (s script) Run(ctx context.Context, c Cell, i int) (RunResult, error) { return s(c, i), nil }

func pass(usd float64) RunResult {
	return RunResult{Status: run.Completed, Assertions: []assert.Result{{Kind: "ran", Desc: "ran guide", Pass: true}}, Cost: Cost{Calls: 4, USD: usd, Duration: time.Second}}
}

func failing(desc string) RunResult {
	return RunResult{Status: run.Completed, Assertions: []assert.Result{{Kind: "ran", Desc: desc, Pass: false, Reason: "x"}}, Cost: Cost{Calls: 8, USD: 0.02}}
}

func cell(t *testing.T, r *Result, c, v, m string) *CellResult {
	t.Helper()
	for _, cr := range r.Cells {
		if cr.Cell == (Cell{c, v, m}) {
			return cr
		}
	}
	t.Fatalf("no cell %s/%s/%s", c, v, m)
	return nil
}

func TestRatesFlakyErrorsAndCost(t *testing.T) {
	cases := []*evalcase.Case{mkCase(t, "a", ""), mkCase(t, "b", "")}
	res, err := Run(context.Background(), Options{Cases: cases, Variants: []string{"current"}, Models: []string{"haiku", "sonnet"}, N: 10, Parallel: 4,
		Runner: script(func(c Cell, i int) RunResult {
			switch {
			case c.Case == "a" && c.Model == "sonnet":
				return pass(0.01)
			case c.Case == "a": // haiku: 6 of 10, 2 harness errors
				if i >= 8 {
					return RunResult{Status: run.Errored, Err: "crash"}
				}
				if i < 6 {
					return pass(0.002)
				}
				return failing("ran guide")
			default: // b always fails
				return failing("ran knowledge find")
			}
		})})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Cells) != 4 {
		t.Fatalf("cells: %d", len(res.Cells))
	}
	s := cell(t, res, "a", "current", "sonnet")
	if s.Passed != 10 || s.Rate != 1 || s.Flaky || s.Lo < 0.7 || s.Hi != 1 {
		t.Errorf("sonnet a: %+v", s)
	}
	h := cell(t, res, "a", "current", "haiku")
	if h.Passed != 6 || h.Failed != 2 || h.Errored != 2 || h.Rate != 0.75 || !h.Flaky {
		t.Errorf("haiku a: %+v", h)
	}
	if h.Lo > 0.75 || h.Hi < 0.75 || h.Lo <= 0.3 || h.Hi >= 0.98 {
		t.Errorf("interval [%v, %v]", h.Lo, h.Hi)
	}
	if h.FailedBy["ran guide"] != 2 {
		t.Errorf("failed by: %v", h.FailedBy)
	}
	if h.Cost.USD.Mean <= 0 || h.Cost.Calls.Std <= 0 || h.Cost.Calls.Mean < 4 {
		t.Errorf("cost: %+v", h.Cost)
	}
	b := cell(t, res, "b", "current", "haiku")
	if b.Passed != 0 || b.Rate != 0 || b.Flaky || b.Hi > 0.4 || b.FailedBy["ran knowledge find"] != 10 {
		t.Errorf("b: %+v", b)
	}
}

func TestJudgedAndUnjudged(t *testing.T) {
	pj := pass(0)
	pj.Judged = []assert.Result{{Kind: "rubric", Desc: "clear", Pass: true}}
	fj := pass(0)
	fj.Judged = []assert.Result{{Kind: "rubric", Desc: "clear", Pass: false}}
	uj := pass(0)
	uj.Judged = []assert.Result{{Kind: "rubric", Desc: "clear", Errored: true}}
	if !pj.Passed() || fj.Passed() || !uj.Passed() || !uj.Unjudged() {
		t.Error("pass rules for judged items")
	}
	res, _ := Run(context.Background(), Options{Cases: []*evalcase.Case{mkCase(t, "a", "")}, Variants: []string{"v"}, Models: []string{"m"}, N: 3, Parallel: 1,
		Runner: script(func(c Cell, i int) RunResult { return [3]RunResult{pj, fj, uj}[i] })})
	cr := res.Cells[0]
	if cr.Passed != 2 || cr.Failed != 1 || cr.Unjudged != 1 || cr.FailedBy["clear"] != 1 {
		t.Errorf("%+v", cr)
	}
}

func TestLimitStatusesAreFailures(t *testing.T) {
	rec := &run.Record{Status: run.Turns, Err: "turn limit reached"}
	c := mkCase(t, "a", "")
	comp, _ := c.Compile(func(s string) (string, error) { return s, nil })
	rec.Compiled = comp
	rs := Evaluate(rec)
	if len(rs) != 1 || rs[0].Pass || rs[0].Kind != "limit" {
		t.Errorf("%+v", rs)
	}
}

func TestShuffleIsSeeded(t *testing.T) {
	order := func(seed int64) string {
		var mu sync.Mutex
		var got []string
		Run(context.Background(), Options{Cases: []*evalcase.Case{mkCase(t, "a", ""), mkCase(t, "b", "")}, Variants: []string{"v"}, Models: []string{"m"}, N: 4, Parallel: 1, Seed: seed,
			Runner: script(func(c Cell, i int) RunResult { return pass(0) }),
			OnResult: func(r RunResult) {
				mu.Lock()
				got = append(got, fmt.Sprintf("%s%d", r.Cell.Case, r.Index))
				mu.Unlock()
			}})
		return strings.Join(got, ",")
	}
	if order(1) != order(1) {
		t.Error("same seed, different order")
	}
	if order(1) == order(2) && order(1) == order(3) {
		t.Error("seed does not shuffle")
	}
}

func TestSequentialStop(t *testing.T) {
	var mu sync.Mutex
	runs := map[string]int{}
	res, _ := Run(context.Background(), Options{Cases: []*evalcase.Case{mkCase(t, "always", ""), mkCase(t, "never", ""), mkCase(t, "mixed", "")}, Variants: []string{"v"}, Models: []string{"m"},
		N: 30, Parallel: 1, Sequential: true,
		Runner: script(func(c Cell, i int) RunResult {
			mu.Lock()
			runs[c.Case]++
			mu.Unlock()
			switch c.Case {
			case "always":
				return pass(0)
			case "never":
				return failing("x")
			}
			if i%2 == 0 {
				return pass(0)
			}
			return failing("x")
		})})
	if runs["always"] >= 30 || runs["never"] >= 30 || runs["always"] < 3 {
		t.Errorf("early stop: %v", runs)
	}
	if runs["mixed"] != 30 {
		t.Errorf("mixed cell stopped early: %v", runs)
	}
	if res.Aborted {
		t.Error("early stop of cells is not an abort")
	}
}

func TestBudgetStopKeepsFinishedRuns(t *testing.T) {
	res, _ := Run(context.Background(), Options{Cases: []*evalcase.Case{mkCase(t, "a", "")}, Variants: []string{"v"}, Models: []string{"m"}, N: 20, Parallel: 1,
		MaxUSD: 0.5, Estimate: func(Cell) float64 { return 0.1 },
		Runner: script(func(c Cell, i int) RunResult { return pass(0.1) })})
	cr := res.Cells[0]
	if !res.Aborted || res.StoppedBy != "budget" || len(cr.Runs) < 4 || len(cr.Runs) > 5 || res.SpentUSD > 0.5001 {
		t.Errorf("stopped %q after %d runs, spent %v", res.StoppedBy, len(cr.Runs), res.SpentUSD)
	}
	if cr.Passed != len(cr.Runs) {
		t.Errorf("finished runs lost: %+v", cr)
	}
}

func TestCancelKeepsFinishedRuns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	done := 0
	res, _ := Run(ctx, Options{Cases: []*evalcase.Case{mkCase(t, "a", "")}, Variants: []string{"v"}, Models: []string{"m"}, N: 50, Parallel: 1,
		Runner: script(func(c Cell, i int) RunResult { return pass(0) }),
		OnResult: func(RunResult) {
			mu.Lock()
			done++
			if done == 3 {
				cancel()
			}
			mu.Unlock()
		}})
	if !res.Aborted || res.StoppedBy != "cancelled" || len(res.Cells[0].Runs) != 3 {
		t.Errorf("%+v runs %d", res, len(res.Cells[0].Runs))
	}
}

func TestCaseRestrictions(t *testing.T) {
	cs := []*evalcase.Case{mkCase(t, "a", "models: [haiku]\n"), mkCase(t, "b", "variants: [x]\n"), mkCase(t, "c", "")}
	got := Cells(Options{Cases: cs, Variants: []string{"current", "x"}, Models: []string{"haiku", "opus"}})
	var names []string
	for _, c := range got {
		names = append(names, c.String())
	}
	want := "a / current / haiku,a / x / haiku,b / x / haiku,b / x / opus,c / current / haiku,c / current / opus,c / x / haiku,c / x / opus"
	if strings.Join(names, ",") != want {
		t.Errorf("cells:\n%s\nwant\n%s", strings.Join(names, ","), want)
	}
	if _, err := Run(context.Background(), Options{Cases: cs[:1], Variants: []string{"v"}, Models: []string{"opus"}, Runner: script(nil)}); err == nil {
		t.Error("empty matrix accepted")
	}
}

func TestLiveRunner(t *testing.T) {
	root := t.TempDir()
	plugin := filepath.Join(root, "plugin")
	os.MkdirAll(plugin, 0o755)
	os.WriteFile(filepath.Join(plugin, "plugin.json"), []byte("{}"), 0o644)
	claude := filepath.Join(root, "claude")
	os.WriteFile(claude, []byte(`#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.9.9"; exit 0; fi
cat >/dev/null
prep new --title "Refund flow" --kind code >/dev/null
cat <<'S'
{"type":"system","subtype":"init","model":"claude-haiku-5-5","session_id":"s"}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"prep new --title Refund --kind code"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"ok"}]}}
{"type":"result","subtype":"success","result":"done","num_turns":2,"total_cost_usd":0.03,"usage":{"input_tokens":10,"output_tokens":5}}
S
`), 0o755)
	c := mkCase(t, "refund", "expect:\n  tools: [{ran: new}]\n  outcome: [{issue: {title: refund}}]\n")
	live := &Live{Cases: map[string]*evalcase.Case{"refund": c}, Variants: variant.Resolver{Root: root, PluginDir: plugin}, Binary: prepBin, Claude: claude,
		Auth: env.Auth{APIKey: "k"}, Parent: t.TempDir(), OutDir: t.TempDir()}
	res, err := Run(context.Background(), Options{Cases: []*evalcase.Case{c}, Variants: []string{"current"}, Models: []string{"haiku"}, N: 2, Parallel: 2, Runner: live})
	if err != nil {
		t.Fatal(err)
	}
	cr := res.Cells[0]
	if cr.Passed != 2 || cr.Resolved[0] != "claude-haiku-5-5" || cr.Cost.USD.Mean != 0.03 || cr.Runs[0].Harness != "9.9.9" {
		t.Errorf("%+v", cr)
	}
	// Both passing runs released their environments.
	if entries, _ := os.ReadDir(live.Parent); len(entries) != 0 {
		t.Errorf("environments left behind: %v", entries)
	}
	if _, err := os.Stat(filepath.Join(live.OutDir, "runs", "refund", "current", "haiku", "0", "transcript.json")); err != nil {
		t.Errorf("run files: %v", err)
	}
}
