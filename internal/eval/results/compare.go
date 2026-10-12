package results

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// Class says how a cell differs between two sides.
type Class string

const (
	Improved  Class = "improved"
	Regressed Class = "regressed"
	Unchanged Class = "unchanged"
	// Noisy means the rates differ but the intervals overlap: the
	// difference may be chance.
	Noisy Class = "noisy"
	OnlyA Class = "only in A"
	OnlyB Class = "only in B"
)

// Delta compares one cell on two sides.
type Delta struct {
	Key   string
	A, B  *matrix.CellResult
	Class Class
	// Rate is B's rate minus A's.
	Rate float64
	// CostPct is the change in mean cost per run, in percent (0 if unknown).
	CostPct float64
	// CallsPct is the same for tool calls.
	CallsPct float64
}

// Compare lines two lists of cells up by key and classifies each. A
// difference is Improved or Regressed only when the intervals do not
// overlap.
func Compare(a, b []*matrix.CellResult, key func(matrix.Cell) string) []Delta {
	am, bm := map[string]*matrix.CellResult{}, map[string]*matrix.CellResult{}
	var keys []string
	seen := map[string]bool{}
	for _, c := range a {
		k := key(c.Cell)
		am[k] = c
		if !seen[k] {
			seen[k], keys = true, append(keys, k)
		}
	}
	for _, c := range b {
		k := key(c.Cell)
		bm[k] = c
		if !seen[k] {
			seen[k], keys = true, append(keys, k)
		}
	}
	sort.Strings(keys)
	var out []Delta
	for _, k := range keys {
		d := Delta{Key: k, A: am[k], B: bm[k]}
		switch {
		case d.A == nil:
			d.Class = OnlyB
		case d.B == nil:
			d.Class = OnlyA
		default:
			d.Rate = d.B.Rate - d.A.Rate
			d.CostPct = change(d.A.Cost.USD.Mean, d.B.Cost.USD.Mean)
			d.CallsPct = change(d.A.Cost.Calls.Mean, d.B.Cost.Calls.Mean)
			switch {
			case d.B.Lo > d.A.Hi:
				d.Class = Improved
			case d.B.Hi < d.A.Lo:
				d.Class = Regressed
			case d.A.Rate == d.B.Rate:
				d.Class = Unchanged
			default:
				d.Class = Noisy
			}
		}
		out = append(out, d)
	}
	return out
}

func change(a, b float64) float64 {
	if a == 0 {
		return 0
	}
	return (b - a) / a * 100
}

// FullKey identifies a cell by case, variant and model.
func FullKey(c matrix.Cell) string { return c.Case + " / " + c.Variant + " / " + c.Model }

// CaseModelKey identifies a cell by case and model, for comparing variants
// inside one run.
func CaseModelKey(c matrix.Cell) string { return c.Case + " / " + c.Model }

// ForVariant returns the cells of one variant.
func ForVariant(cells []*matrix.CellResult, variant string) []*matrix.CellResult {
	var out []*matrix.CellResult
	for _, c := range cells {
		if c.Cell.Variant == variant {
			out = append(out, c)
		}
	}
	return out
}

// FormatCompare prints deltas grouped by class.
func FormatCompare(nameA, nameB string, ds []Delta) string {
	var b strings.Builder
	fmt.Fprintf(&b, "A = %s\nB = %s\n", nameA, nameB)
	groups := map[Class][]Delta{}
	for _, d := range ds {
		groups[d.Class] = append(groups[d.Class], d)
	}
	for _, c := range []Class{Regressed, Improved, Noisy, Unchanged, OnlyA, OnlyB} {
		g := groups[c]
		if len(g) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s (%d)\n", strings.ToUpper(string(c)), len(g))
		for _, d := range g {
			switch c {
			case OnlyA:
				fmt.Fprintf(&b, "  %s: %s\n", d.Key, RateText(d.A))
			case OnlyB:
				fmt.Fprintf(&b, "  %s: %s\n", d.Key, RateText(d.B))
			default:
				fmt.Fprintf(&b, "  %s: %s -> %s  cost %+.0f%%  calls %+.0f%%\n", d.Key, RateText(d.A), RateText(d.B), d.CostPct, d.CallsPct)
			}
		}
	}
	if len(ds) == 0 {
		b.WriteString("\nnothing to compare\n")
	}
	return b.String()
}
