package mutate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/results"
)

// Row is one mutant measured against the base.
type Row struct {
	Variant     string `json:"variant"`
	Description string `json:"description,omitempty"`
	// Cells is how many (case, model) cells both sides have.
	Cells                                 int `json:"cells"`
	Improved, Regressed, Noisy, Unchanged int
	// BaseRate and Rate are the pass rates pooled over those cells.
	BaseRate float64 `json:"base_rate"`
	Rate     float64 `json:"rate"`
	Delta    float64 `json:"delta"`
	// Significant is true when the pooled intervals do not overlap.
	Significant bool `json:"significant"`
	// CostPct is the change in mean cost per run, in percent.
	CostPct float64 `json:"cost_pct"`
	// ByArea is the pooled delta per area.
	ByArea map[string]float64 `json:"by_area,omitempty"`
}

// Rank measures each mutant against the base over the cells they share. The
// rows are sorted by pooled improvement, then by cost.
func Rank(cells []*matrix.CellResult, base string, mutants []string, areas, descriptions map[string]string) []Row {
	baseCells := results.ForVariant(cells, base)
	var rows []Row
	for _, m := range mutants {
		mc := results.ForVariant(cells, m)
		row := Row{Variant: m, Description: descriptions[m], ByArea: map[string]float64{}}
		ds := results.Compare(baseCells, mc, results.CaseModelKey)
		type pool struct{ ap, an, bp, bn int }
		var total pool
		byArea := map[string]*pool{}
		var costA, costB float64
		for _, d := range ds {
			if d.A == nil || d.B == nil {
				continue
			}
			row.Cells++
			switch d.Class {
			case results.Improved:
				row.Improved++
			case results.Regressed:
				row.Regressed++
			case results.Noisy:
				row.Noisy++
			default:
				row.Unchanged++
			}
			area := areas[d.A.Cell.Case]
			if area == "" {
				area = "other"
			}
			if byArea[area] == nil {
				byArea[area] = &pool{}
			}
			for _, p := range []*pool{&total, byArea[area]} {
				p.ap += d.A.Passed
				p.an += d.A.Passed + d.A.Failed
				p.bp += d.B.Passed
				p.bn += d.B.Passed + d.B.Failed
			}
			costA += d.A.Cost.USD.Mean
			costB += d.B.Cost.USD.Mean
		}
		rate := func(k, n int) float64 {
			if n == 0 {
				return 0
			}
			return float64(k) / float64(n)
		}
		row.BaseRate, row.Rate = rate(total.ap, total.an), rate(total.bp, total.bn)
		row.Delta = row.Rate - row.BaseRate
		alo, ahi := matrix.Wilson(total.ap, total.an)
		blo, bhi := matrix.Wilson(total.bp, total.bn)
		row.Significant = total.an > 0 && total.bn > 0 && (blo > ahi || bhi < alo)
		if costA > 0 {
			row.CostPct = (costB - costA) / costA * 100
		}
		for a, p := range byArea {
			row.ByArea[a] = rate(p.bp, p.bn) - rate(p.ap, p.an)
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Delta != rows[j].Delta {
			return rows[i].Delta > rows[j].Delta
		}
		return rows[i].CostPct < rows[j].CostPct
	})
	return rows
}

// Format prints the ranking.
func Format(base string, baseRate float64, rows []Row) string {
	var b strings.Builder
	fmt.Fprintf(&b, "mutants against %s (pass rate %.0f%% over the shared cells)\n\n", base, baseRate*100)
	for _, r := range rows {
		verdict := "within the noise"
		if r.Significant && r.Delta > 0 {
			verdict = "BETTER beyond the noise"
		} else if r.Significant {
			verdict = "WORSE beyond the noise"
		}
		fmt.Fprintf(&b, "%-32s %+5.0f pts  (%.0f%% -> %.0f%%)  cost %+.0f%%  %s\n", r.Variant, r.Delta*100, r.BaseRate*100, r.Rate*100, r.CostPct, verdict)
		if r.Description != "" {
			fmt.Fprintf(&b, "    %s\n", strings.ReplaceAll(r.Description, "\n", "\n    "))
		}
		var areas []string
		for a, d := range r.ByArea {
			if d != 0 {
				areas = append(areas, fmt.Sprintf("%s %+.0f", a, d*100))
			}
		}
		sort.Strings(areas)
		fmt.Fprintf(&b, "    cells: %d better, %d worse, %d noisy, %d same", r.Improved, r.Regressed, r.Noisy, r.Unchanged)
		if len(areas) > 0 {
			fmt.Fprintf(&b, "   by area (pts): %s", strings.Join(areas, ", "))
		}
		b.WriteString("\n")
	}
	return b.String()
}
