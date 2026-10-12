package results

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// ReportOptions shape a report.
type ReportOptions struct {
	// Failures lists the failing assertions of each cell with the run
	// directories to inspect.
	Failures bool
}

// pct formats a rate as a percentage.
func pct(f float64) string { return fmt.Sprintf("%.0f%%", f*100) }

// RateText is "8/10 80% [49-94]" for a cell, with ~ for a flaky one.
func RateText(c *matrix.CellResult) string {
	n := c.Passed + c.Failed
	if n == 0 {
		return "no runs"
	}
	s := fmt.Sprintf("%d/%d %s [%s-%s]", c.Passed, n, pct(c.Rate), pct(c.Lo), pct(c.Hi))
	if c.Flaky {
		s += " ~"
	}
	return s
}

// Report prints a run as one table per case: a row per variant and model.
func Report(r *Run, o ReportOptions) string {
	var b strings.Builder
	fmt.Fprintf(&b, "run %s  (%s)\n", r.ID, r.Created.Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "N=%d  seed=%d  prep %s", r.Definition.N, r.Definition.Seed, r.Versions.Prep)
	if r.Versions.Harness != "" {
		fmt.Fprintf(&b, "  harness %s", r.Versions.Harness)
	}
	fmt.Fprintf(&b, "  spent $%.2f", r.SpentUSD)
	if r.StoppedBy != "" {
		fmt.Fprintf(&b, "  STOPPED EARLY (%s)", r.StoppedBy)
	}
	b.WriteString("\n")
	byCase := map[string][]*matrix.CellResult{}
	var cases []string
	for _, c := range r.Cells {
		if _, ok := byCase[c.Cell.Case]; !ok {
			cases = append(cases, c.Cell.Case)
		}
		byCase[c.Cell.Case] = append(byCase[c.Cell.Case], c)
	}
	sort.Strings(cases)
	for _, name := range cases {
		fmt.Fprintf(&b, "\n%s\n", name)
		cells := byCase[name]
		sort.SliceStable(cells, func(i, j int) bool {
			if cells[i].Cell.Variant != cells[j].Cell.Variant {
				return cells[i].Cell.Variant < cells[j].Cell.Variant
			}
			return cells[i].Cell.Model < cells[j].Cell.Model
		})
		w := 0
		for _, c := range cells {
			if l := len(c.Cell.Variant) + len(c.Cell.Model) + 3; l > w {
				w = l
			}
		}
		for _, c := range cells {
			label := c.Cell.Variant + " / " + c.Cell.Model
			extra := ""
			if c.Errored > 0 {
				extra += fmt.Sprintf("  %d errored", c.Errored)
			}
			if c.Unjudged > 0 {
				extra += fmt.Sprintf("  %d unjudged", c.Unjudged)
			}
			fmt.Fprintf(&b, "  %-*s  %-22s  $%.3f/run  %.1f calls%s\n", w, label, RateText(c), c.Cost.USD.Mean, c.Cost.Calls.Mean, extra)
			if o.Failures && len(c.FailedBy) > 0 {
				type fb struct {
					desc string
					n    int
				}
				var fs []fb
				for d, n := range c.FailedBy {
					fs = append(fs, fb{d, n})
				}
				sort.Slice(fs, func(i, j int) bool {
					if fs[i].n != fs[j].n {
						return fs[i].n > fs[j].n
					}
					return fs[i].desc < fs[j].desc
				})
				for _, f := range fs {
					fmt.Fprintf(&b, "      failed %dx: %s\n", f.n, f.desc)
				}
				for _, run := range c.Runs {
					if !run.Passed() && run.Dir != "" {
						fmt.Fprintf(&b, "      see %s\n", run.Dir)
					}
				}
			}
		}
	}
	return b.String()
}
