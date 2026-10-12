// Package models turns a matrix run over several models into the table the
// user decides by: for each area of the CLI, which is the cheapest model
// that is enough, what the cheaper ones get wrong, and what it costs per
// passing run.
package models

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// Options say what counts as enough.
type Options struct {
	// Variant restricts the analysis to one variant (default: the first in
	// the run).
	Variant string
	// MinRate and MinLower: a model is sufficient in an area when its pooled
	// pass rate is at least MinRate and the lower bound of its 95% interval
	// at least MinLower (defaults 0.9 and 0.8).
	MinRate, MinLower float64
	// MinRuns is the fewest runs in an area that can show anything (default 5).
	MinRuns int
}

func (o *Options) defaults() {
	if o.MinRate == 0 {
		o.MinRate = 0.9
	}
	if o.MinLower == 0 {
		o.MinLower = 0.8
	}
	if o.MinRuns == 0 {
		o.MinRuns = 5
	}
}

// Failure counts one failing assertion.
type Failure struct {
	Desc  string `json:"desc"`
	Kind  string `json:"kind"`
	Group string `json:"group"`
	Count int    `json:"count"`
	// Examples are run directories to inspect.
	Examples []string `json:"examples,omitempty"`
}

// Stat is one model's pooled result in one area (or case).
type Stat struct {
	Model    string   `json:"model"`
	Resolved []string `json:"resolved,omitempty"`
	Runs     int      `json:"runs"`
	Passed   int      `json:"passed"`
	Errored  int      `json:"errored"`
	Rate     float64  `json:"rate"`
	Lo       float64  `json:"lo"`
	Hi       float64  `json:"hi"`
	// USDPerRun is the mean cost of a run; USDPerPass divides the total by the
	// passes (zero when nothing passed).
	USDPerRun  float64 `json:"usd_per_run"`
	USDPerPass float64 `json:"usd_per_pass"`
	// Sufficient is true when the model is enough for this area.
	Sufficient bool      `json:"sufficient"`
	Failures   []Failure `json:"failures,omitempty"`
}

// Status of an area's decision.
type Status string

const (
	Decided Status = "decided"
	// None means no model is sufficient.
	None Status = "none sufficient"
	// NotEnough means the runs are too few to say.
	NotEnough Status = "not enough data"
)

// Area is the decision for one area.
type Area struct {
	Area   string `json:"area"`
	Status Status `json:"status"`
	// Cheapest is the cheapest sufficient model, by cost per passing run.
	Cheapest string `json:"cheapest,omitempty"`
	// NextUp is the model one step more expensive than Cheapest (or than the
	// cheapest overall when none is sufficient).
	NextUp string `json:"next_up,omitempty"`
	Models []Stat `json:"models"`
	// Cases holds each case's stats per model.
	Cases []Case `json:"cases"`
	// Hint says what is missing when the status is not enough data: a model
	// that looks sufficient but whose interval is too wide for the runs.
	Hint string `json:"hint,omitempty"`
	// WrongBelow lists what models cheaper than Cheapest get wrong.
	WrongBelow []Failure `json:"wrong_below,omitempty"`
}

// Case is one case's result per model.
type Case struct {
	Case   string `json:"case"`
	Models []Stat `json:"models"`
}

// Analysis is the whole decision table.
type Analysis struct {
	RunID    string   `json:"run"`
	Variant  string   `json:"variant"`
	Models   []string `json:"models"`
	Harness  string   `json:"harness,omitempty"`
	N        int      `json:"n"`
	Areas    []Area   `json:"areas"`
	Warnings []string `json:"warnings,omitempty"`
	Options  Options  `json:"options"`
}

// Group names the kind of failure an assertion kind is.
func Group(kind string) string {
	switch kind {
	case "rubric":
		return "quality"
	case "limit":
		return "limits"
	case "issue", "unchanged", "knowledge", "check", "claimed", "repo":
		return "outcome"
	}
	return "tool use"
}

// Analyze builds the table from a run's cells. areas maps case names to
// areas (cases without one fall under "other").
func Analyze(runID, harness string, n int, cells []*matrix.CellResult, areas map[string]string, o Options) (*Analysis, error) {
	o.defaults()
	if o.Variant == "" {
		for _, c := range cells {
			if o.Variant == "" || c.Cell.Variant < o.Variant {
				o.Variant = c.Cell.Variant
			}
		}
	}
	a := &Analysis{RunID: runID, Variant: o.Variant, Harness: harness, N: n, Options: o}
	byAreaModel := map[string]map[string][]*matrix.CellResult{}
	modelSet := map[string]bool{}
	resolved := map[string]map[string]bool{} // requested -> resolved ids
	for _, c := range cells {
		if c.Cell.Variant != o.Variant {
			continue
		}
		area := areas[c.Cell.Case]
		if area == "" {
			area = "other"
		}
		if byAreaModel[area] == nil {
			byAreaModel[area] = map[string][]*matrix.CellResult{}
		}
		byAreaModel[area][c.Cell.Model] = append(byAreaModel[area][c.Cell.Model], c)
		modelSet[c.Cell.Model] = true
		if resolved[c.Cell.Model] == nil {
			resolved[c.Cell.Model] = map[string]bool{}
		}
		for _, r := range c.Resolved {
			resolved[c.Cell.Model][r] = true
		}
		if len(c.Resolved) > 1 {
			a.Warnings = append(a.Warnings, fmt.Sprintf("%s ran on more than one model id (%s): the alias moved during the run", c.Cell, strings.Join(c.Resolved, ", ")))
		}
	}
	if len(modelSet) == 0 {
		return nil, fmt.Errorf("the run has no cells for variant %q", o.Variant)
	}
	for m := range modelSet {
		a.Models = append(a.Models, m)
		if len(resolved[m]) > 1 {
			var ids []string
			for id := range resolved[m] {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			a.Warnings = append(a.Warnings, fmt.Sprintf("model %q resolved to several ids across cells: %s", m, strings.Join(ids, ", ")))
		}
	}
	sort.Strings(a.Models)
	// Cost order of the models overall, cheapest first.
	order := costOrder(cells, o.Variant)
	var names []string
	for n := range byAreaModel {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		a.Areas = append(a.Areas, analyzeArea(name, byAreaModel[name], a.Models, order, o))
	}
	sort.Strings(a.Warnings)
	return a, nil
}

// costOrder sorts the models by mean cost per run over all cells.
func costOrder(cells []*matrix.CellResult, variant string) []string {
	sum, n := map[string]float64{}, map[string]int{}
	for _, c := range cells {
		if c.Cell.Variant != variant {
			continue
		}
		for _, r := range c.Runs {
			if r.Status == "error" {
				continue
			}
			sum[c.Cell.Model] += r.Cost.USD
			n[c.Cell.Model]++
		}
	}
	var ms []string
	for m := range sum {
		ms = append(ms, m)
	}
	sort.Slice(ms, func(i, j int) bool {
		ai, aj := sum[ms[i]]/float64(n[ms[i]]), sum[ms[j]]/float64(n[ms[j]])
		if ai != aj {
			return ai < aj
		}
		return ms[i] < ms[j]
	})
	return ms
}

func analyzeArea(name string, byModel map[string][]*matrix.CellResult, models, order []string, o Options) Area {
	area := Area{Area: name}
	for _, m := range models {
		if cs := byModel[m]; len(cs) > 0 {
			area.Models = append(area.Models, pool(m, cs, o))
		}
	}
	caseCells := map[string]map[string][]*matrix.CellResult{}
	for m, cs := range byModel {
		for _, c := range cs {
			if caseCells[c.Cell.Case] == nil {
				caseCells[c.Cell.Case] = map[string][]*matrix.CellResult{}
			}
			caseCells[c.Cell.Case][m] = append(caseCells[c.Cell.Case][m], c)
		}
	}
	var cnames []string
	for c := range caseCells {
		cnames = append(cnames, c)
	}
	sort.Strings(cnames)
	for _, cn := range cnames {
		cs := Case{Case: cn}
		for _, m := range models {
			if cells := caseCells[cn][m]; len(cells) > 0 {
				cs.Models = append(cs.Models, pool(m, cells, o))
			}
		}
		area.Cases = append(area.Cases, cs)
	}
	// The decision.
	best := -1
	enough := true
	for i, s := range area.Models {
		if s.Runs < o.MinRuns {
			enough = false
		}
		if s.Sufficient && (best < 0 || better(s, area.Models[best])) {
			best = i
		}
	}
	switch {
	case best >= 0:
		area.Status, area.Cheapest = Decided, area.Models[best].Model
	case !enough:
		area.Status = NotEnough
	default:
		area.Status = None
	}
	// A model that looks good but is not yet provable makes the status "not
	// enough data", with the number of runs that would settle it.
	if best < 0 {
		for _, s := range area.Models {
			if s.Rate >= o.MinRate {
				area.Status = NotEnough
				area.Hint = fmt.Sprintf("%s passes %s of %d runs, but that many runs cannot show a rate above %s with confidence; about %d runs are needed",
					s.Model, pct(s.Rate), s.Runs, pct(o.MinLower), runsNeeded(s.Rate, o))
				break
			}
		}
	}
	// Next model up in cost order, and what the cheaper models get wrong.
	pos := -1
	if area.Cheapest != "" {
		for i, m := range order {
			if m == area.Cheapest {
				pos = i
			}
		}
	}
	if area.Cheapest != "" {
		if pos >= 0 && pos+1 < len(order) {
			area.NextUp = order[pos+1]
		}
		cheaper := map[string]bool{}
		for i := 0; i < pos; i++ {
			cheaper[order[i]] = true
		}
		agg := map[string]*Failure{}
		for _, s := range area.Models {
			if !cheaper[s.Model] {
				continue
			}
			for _, f := range s.Failures {
				if agg[f.Desc] == nil {
					cp := f
					cp.Examples = nil
					agg[f.Desc] = &cp
				} else {
					agg[f.Desc].Count += f.Count
				}
			}
		}
		for _, f := range agg {
			area.WrongBelow = append(area.WrongBelow, *f)
		}
		sortFailures(area.WrongBelow)
	} else if len(order) > 0 {
		area.NextUp = order[0]
	}
	return area
}

// runsNeeded is the fewest runs at which the given pass rate would clear the
// interval bound (capped at 500).
func runsNeeded(rate float64, o Options) int {
	for n := o.MinRuns; n <= 500; n++ {
		k := int(math.Round(rate * float64(n)))
		if lo, _ := matrix.Wilson(k, n); lo >= o.MinLower && float64(k)/float64(n) >= o.MinRate {
			return n
		}
	}
	return 500
}

// better reports whether a is the cheaper per passing run.
func better(a, b Stat) bool {
	if a.USDPerPass != b.USDPerPass {
		return a.USDPerPass < b.USDPerPass
	}
	return a.Model < b.Model
}

func sortFailures(fs []Failure) {
	sort.Slice(fs, func(i, j int) bool {
		if fs[i].Count != fs[j].Count {
			return fs[i].Count > fs[j].Count
		}
		return fs[i].Desc < fs[j].Desc
	})
}

// pool adds up a model's cells into one stat.
func pool(model string, cells []*matrix.CellResult, o Options) Stat {
	s := Stat{Model: model}
	res := map[string]bool{}
	var usd float64
	fails := map[string]*Failure{}
	for _, c := range cells {
		s.Passed += c.Passed
		s.Runs += c.Passed + c.Failed
		s.Errored += c.Errored
		for _, r := range c.Resolved {
			res[r] = true
		}
		for _, r := range c.Runs {
			if r.Status != "error" {
				usd += r.Cost.USD
			}
		}
		for desc, n := range c.FailedBy {
			f := fails[desc]
			if f == nil {
				kind := c.FailKinds[desc]
				f = &Failure{Desc: desc, Kind: kind, Group: Group(kind)}
				fails[desc] = f
			}
			f.Count += n
		}
	}
	for _, c := range cells {
		for _, r := range c.Runs {
			if r.Dir == "" || r.Passed() || r.Status == "error" {
				continue
			}
			for _, a := range r.Assertions {
				if f := fails[a.Desc]; !a.Pass && f != nil && addable(f, r.Dir) {
					f.Examples = append(f.Examples, r.Dir)
				}
			}
			for _, a := range r.Judged {
				if f := fails[a.Desc]; !a.Pass && !a.Errored && f != nil && addable(f, r.Dir) {
					f.Examples = append(f.Examples, r.Dir)
				}
			}
		}
	}
	for r := range res {
		s.Resolved = append(s.Resolved, r)
	}
	sort.Strings(s.Resolved)
	if s.Runs > 0 {
		s.Rate = float64(s.Passed) / float64(s.Runs)
		s.USDPerRun = usd / float64(s.Runs)
	}
	if s.Passed > 0 {
		s.USDPerPass = usd / float64(s.Passed)
	}
	s.Lo, s.Hi = matrix.Wilson(s.Passed, s.Runs)
	s.Sufficient = s.Runs >= o.MinRuns && s.Rate >= o.MinRate && s.Lo >= o.MinLower
	for _, f := range fails {
		s.Failures = append(s.Failures, *f)
	}
	sortFailures(s.Failures)
	return s
}

// addable reports whether dir can still be listed as an example of f.
func addable(f *Failure, dir string) bool {
	if len(f.Examples) >= 3 {
		return false
	}
	for _, e := range f.Examples {
		if e == dir {
			return false
		}
	}
	return true
}
