package matrix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/run"
	"github.com/fluid-movement/prep/internal/eval/variant"
)

// JudgeFunc scores a run's rubric; see package judge.
type JudgeFunc func(ctx context.Context, rec *run.Record, rubric []string) []assert.Result

// Live runs cells for real: it lays out the variant, runs the agent in a
// fresh workspace and checks the assertions.
type Live struct {
	Cases    map[string]*evalcase.Case
	Variants variant.Resolver
	Binary   string
	Claude   string
	Auth     env.Auth
	// Parent is where run environments are created.
	Parent string
	// OutDir receives <case>/<variant>/<model>/<n>/ per run; "" stores nothing.
	OutDir string
	// Keep keeps every environment.
	Keep bool
	// Judge scores rubrics; nil leaves rubrics unjudged.
	Judge JudgeFunc
}

// Run implements Runner.
func (l *Live) Run(ctx context.Context, cell Cell, index int) (RunResult, error) {
	c := l.Cases[cell.Case]
	if c == nil {
		return RunResult{}, fmt.Errorf("unknown case %q", cell.Case)
	}
	tmp, err := os.MkdirTemp(l.Parent, "variant-")
	if err != nil {
		return RunResult{}, err
	}
	defer os.RemoveAll(tmp)
	v, err := l.Variants.Materialize(cell.Variant, filepath.Join(tmp, "plugin"))
	if err != nil {
		return RunResult{}, err
	}
	var outDir string
	if l.OutDir != "" {
		outDir = filepath.Join(l.OutDir, "runs", c.Name, strings.ReplaceAll(cell.Variant, "/", "_"), cell.Model, fmt.Sprint(index))
	}
	rec, err := run.Run(ctx, run.Request{Case: c, Variant: v, Model: cell.Model, Binary: l.Binary, Claude: l.Claude,
		Auth: l.Auth, Parent: l.Parent, OutDir: outDir, Keep: l.Keep})
	if err != nil {
		return RunResult{}, err
	}
	rr := RunResult{Status: rec.Status, Err: rec.Err, ModelResolved: rec.ModelResolved, Harness: rec.HarnessVersion, Dir: outDir}
	if rec.Transcript != nil {
		t := rec.Transcript
		rr.Cost = Cost{Calls: t.Counts.ToolCalls, Errors: t.Counts.Errors, TokensIn: t.Counts.InputTokens + t.Counts.CacheRead + t.Counts.CacheWrite,
			TokensOut: t.Counts.OutputTokens, USD: t.Counts.CostUSD, Duration: rec.Duration}
	}
	if rec.Status != run.Errored && rec.Input != nil {
		rr.Assertions = Evaluate(rec)
		if len(rec.Compiled.Rubric) > 0 && l.Judge != nil {
			rr.Judged = l.Judge(ctx, rec, rec.Compiled.Rubric)
		}
	}
	pass := rr.Passed()
	if kept := rec.Release(pass); kept != "" {
		rr.Dir = kept
	}
	return rr, nil
}

// Evaluate runs every deterministic assertion of the run's case, and adds a
// failure when the session was cut off by a limit.
func Evaluate(rec *run.Record) []assert.Result {
	var out []assert.Result
	for _, a := range append(append([]assert.Assertion{}, rec.Compiled.Tools...), rec.Compiled.Outcome...) {
		out = append(out, a.Check(rec.Input))
	}
	switch rec.Status {
	case run.Timeout, run.Budget, run.Turns:
		out = append(out, assert.Result{Kind: "limit", Desc: "the session ended within its limits", Pass: false, Reason: rec.Err, Step: -1})
	}
	return out
}
