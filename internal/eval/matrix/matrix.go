// Package matrix runs the cases across variants and models, several times
// each, and reduces the runs to a result per cell: a pass rate with an
// interval, what failed, and what it cost.
package matrix

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/run"
)

// Cell is one combination to measure.
type Cell struct {
	Case    string `json:"case"`
	Variant string `json:"variant"`
	Model   string `json:"model"`
}

func (c Cell) String() string { return c.Case + " / " + c.Variant + " / " + c.Model }

// RunResult is one run reduced to what matters.
type RunResult struct {
	Cell  Cell `json:"cell"`
	Index int  `json:"index"`
	// Status is how the session ended; Errored runs are excluded from the
	// pass rate.
	Status        run.Status `json:"status"`
	Err           string     `json:"error,omitempty"`
	ModelResolved string     `json:"model_resolved,omitempty"`
	Harness       string     `json:"harness,omitempty"`
	// Assertions are the deterministic results, Judged the rubric ones.
	Assertions []assert.Result `json:"assertions,omitempty"`
	Judged     []assert.Result `json:"judged,omitempty"`
	Cost       Cost            `json:"cost"`
	// Dir is where the run's files are, when kept.
	Dir string `json:"dir,omitempty"`
}

// Passed reports whether the run counts as a pass: not errored, every
// deterministic assertion passed and no rubric item failed.
func (r RunResult) Passed() bool {
	if r.Status == run.Errored {
		return false
	}
	for _, a := range r.Assertions {
		if !a.Pass {
			return false
		}
	}
	for _, a := range r.Judged {
		if !a.Pass && !a.Errored {
			return false
		}
	}
	return true
}

// Unjudged reports whether a rubric item could not be judged.
func (r RunResult) Unjudged() bool {
	for _, a := range r.Judged {
		if a.Errored {
			return true
		}
	}
	return false
}

// Runner runs one cell once.
type Runner interface {
	Run(ctx context.Context, cell Cell, index int) (RunResult, error)
}

// Options configure a matrix run.
type Options struct {
	Cases    []*evalcase.Case
	Variants []string
	Models   []string
	// N is the number of runs per cell.
	N int
	// Parallel is how many runs go at once (default 2).
	Parallel int
	// Seed shuffles the order of runs, so a time-of-day effect does not fall
	// on one cell.
	Seed   int64
	Runner Runner
	// Sequential stops a cell early once its interval is clearly high or low
	// (DecideAbove, DecideBelow) after at least MinRuns (default 3) runs.
	Sequential bool
	MinRuns    int
	// MaxUSD stops the matrix when the total cost of finished runs plus the
	// estimate for those in flight reaches it; 0 means no limit.
	MaxUSD float64
	// Estimate is the expected cost of a run of a cell, used with MaxUSD.
	Estimate func(Cell) float64
	// OnResult is called after every run, from one goroutine at a time.
	OnResult func(RunResult)
}

// CellResult reduces a cell's runs.
type CellResult struct {
	Cell     Cell        `json:"cell"`
	Runs     []RunResult `json:"runs"`
	Planned  int         `json:"planned"`
	Passed   int         `json:"passed"`
	Failed   int         `json:"failed"`
	Errored  int         `json:"errored"`
	Unjudged int         `json:"unjudged"`
	// Rate is passes over runs that were not errored; Lo and Hi are the 95%
	// Wilson interval of it.
	Rate  float64 `json:"rate"`
	Lo    float64 `json:"lo"`
	Hi    float64 `json:"hi"`
	Flaky bool    `json:"flaky"`
	// FailedBy counts the failing assertions by their description.
	FailedBy map[string]int `json:"failed_by,omitempty"`
	Cost     CostStats      `json:"cost"`
	// Resolved lists the model ids the harness reported.
	Resolved []string `json:"resolved,omitempty"`
}

// Result is a whole matrix run.
type Result struct {
	Cells   []*CellResult `json:"cells"`
	Aborted bool          `json:"aborted,omitempty"`
	// StoppedBy says why it ended early: "cancelled" or "budget".
	StoppedBy string    `json:"stopped_by,omitempty"`
	SpentUSD  float64   `json:"spent_usd"`
	Started   time.Time `json:"started"`
}

// Cells lists the cells the options describe: every case with every variant
// and model, restricted by the case's own lists.
func Cells(o Options) []Cell {
	var out []Cell
	for _, c := range o.Cases {
		for _, v := range o.Variants {
			if len(c.Variants) > 0 && !contains(c.Variants, v) {
				continue
			}
			for _, m := range o.Models {
				if len(c.Models) > 0 && !contains(c.Models, m) {
					continue
				}
				out = append(out, Cell{c.Name, v, m})
			}
		}
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

type job struct {
	cell  int
	index int
}

// Run executes the matrix. Cancelling ctx stops it cleanly and keeps the
// runs that finished.
func Run(ctx context.Context, o Options) (*Result, error) {
	if o.Runner == nil {
		return nil, fmt.Errorf("matrix: no runner")
	}
	if o.N < 1 {
		o.N = 1
	}
	if o.Parallel < 1 {
		o.Parallel = 2
	}
	if o.MinRuns < 1 {
		o.MinRuns = 3
	}
	cells := Cells(o)
	if len(cells) == 0 {
		return nil, fmt.Errorf("matrix: nothing to run (no case, variant and model combination)")
	}
	res := &Result{Started: time.Now()}
	results := make([]*CellResult, len(cells))
	var jobs []job
	for i, c := range cells {
		results[i] = &CellResult{Cell: c, Planned: o.N}
		for n := 0; n < o.N; n++ {
			jobs = append(jobs, job{i, n})
		}
	}
	rng := rand.New(rand.NewSource(o.Seed))
	rng.Shuffle(len(jobs), func(a, b int) { jobs[a], jobs[b] = jobs[b], jobs[a] })

	var (
		mu       sync.Mutex
		next     int
		inflight float64
		stopped  string
		spent    float64
	)
	est := func(c Cell) float64 {
		if o.Estimate == nil {
			return 0
		}
		return o.Estimate(c)
	}
	// take returns the next job worth running, or false when none is.
	take := func() (job, bool) {
		mu.Lock()
		defer mu.Unlock()
		for next < len(jobs) {
			if ctx.Err() != nil {
				stopped = "cancelled"
				return job{}, false
			}
			j := jobs[next]
			next++
			cr := results[j.cell]
			if o.Sequential && decided(cr, o.MinRuns) {
				continue
			}
			if o.MaxUSD > 0 && (spent >= o.MaxUSD || spent+inflight+est(cr.Cell) > o.MaxUSD) {
				stopped = "budget"
				next = len(jobs)
				return job{}, false
			}
			inflight += est(cr.Cell)
			return j, true
		}
		return job{}, false
	}
	var wg sync.WaitGroup
	for w := 0; w < o.Parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				j, ok := take()
				if !ok {
					return
				}
				cell := results[j.cell].Cell
				rr, err := o.Runner.Run(ctx, cell, j.index)
				if err != nil {
					rr = RunResult{Status: run.Errored, Err: err.Error()}
				}
				rr.Cell, rr.Index = cell, j.index
				mu.Lock()
				inflight -= est(cell)
				spent += rr.Cost.USD
				if ctx.Err() != nil && rr.Status == run.Errored {
					// a run cut off by the cancellation says nothing; drop it
					mu.Unlock()
					continue
				}
				results[j.cell].Runs = append(results[j.cell].Runs, rr)
				if o.OnResult != nil {
					o.OnResult(rr)
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if stopped == "" && ctx.Err() != nil {
		stopped = "cancelled"
	}
	for _, cr := range results {
		reduce(cr)
		res.Cells = append(res.Cells, cr)
	}
	sort.SliceStable(res.Cells, func(a, b int) bool { return res.Cells[a].Cell.String() < res.Cells[b].Cell.String() })
	res.SpentUSD, res.StoppedBy, res.Aborted = spent, stopped, stopped != ""
	return res, nil
}

// The sequential stop ends a cell once the interval of its pass rate lies
// wholly above DecideAbove or wholly below DecideBelow.
const (
	DecideAbove = 0.8
	DecideBelow = 0.2
)

// decided reports whether a cell's interval is clearly high or clearly low.
func decided(cr *CellResult, minRuns int) bool {
	k, n := 0, 0
	for _, r := range cr.Runs {
		if r.Status == run.Errored {
			continue
		}
		n++
		if r.Passed() {
			k++
		}
	}
	if n < minRuns {
		return false
	}
	lo, hi := wilson(k, n)
	return lo >= DecideAbove || hi <= DecideBelow
}

// Reduce computes a cell result's figures from its runs; Run calls it, and
// the results store calls it on loaded cells.
func Reduce(cr *CellResult) { reduce(cr) }

func reduce(cr *CellResult) {
	sort.SliceStable(cr.Runs, func(a, b int) bool { return cr.Runs[a].Index < cr.Runs[b].Index })
	cr.Passed, cr.Failed, cr.Errored, cr.Unjudged = 0, 0, 0, 0
	cr.FailedBy = map[string]int{}
	var costs []Cost
	resolved := map[string]bool{}
	for _, r := range cr.Runs {
		if r.ModelResolved != "" {
			resolved[r.ModelResolved] = true
		}
		if r.Status == run.Errored {
			cr.Errored++
			continue
		}
		costs = append(costs, r.Cost)
		if r.Unjudged() {
			cr.Unjudged++
		}
		if r.Passed() {
			cr.Passed++
			continue
		}
		cr.Failed++
		for _, a := range append(append([]assert.Result{}, r.Assertions...), r.Judged...) {
			if !a.Pass && !a.Errored {
				cr.FailedBy[a.Desc]++
			}
		}
	}
	n := cr.Passed + cr.Failed
	if n > 0 {
		cr.Rate = float64(cr.Passed) / float64(n)
	}
	cr.Lo, cr.Hi = wilson(cr.Passed, n)
	cr.Flaky = cr.Passed > 0 && cr.Failed > 0
	cr.Cost = costStats(costs)
	cr.Resolved = cr.Resolved[:0]
	for m := range resolved {
		cr.Resolved = append(cr.Resolved, m)
	}
	sort.Strings(cr.Resolved)
}
