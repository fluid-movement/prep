package results

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/eval/matrix"
)

// BaselineCell is what a baseline remembers of a cell: how many of how many
// runs passed.
type BaselineCell struct {
	Case    string `json:"case"`
	Variant string `json:"variant"`
	Model   string `json:"model"`
	Passed  int    `json:"passed"`
	Runs    int    `json:"runs"`
}

// Baseline is a small committed file of pass rates, the reference a CI run
// is checked against.
type Baseline struct {
	Created time.Time      `json:"created"`
	Run     string         `json:"run"`
	Note    string         `json:"note,omitempty"`
	Cells   []BaselineCell `json:"cells"`
}

// NewBaseline records a run's cells.
func NewBaseline(r *Run, note string) *Baseline {
	b := &Baseline{Created: r.Created, Run: r.ID, Note: note}
	for _, c := range r.Cells {
		if n := c.Passed + c.Failed; n > 0 {
			b.Cells = append(b.Cells, BaselineCell{c.Cell.Case, c.Cell.Variant, c.Cell.Model, c.Passed, n})
		}
	}
	sort.Slice(b.Cells, func(i, j int) bool {
		return FullKey(matrix.Cell{Case: b.Cells[i].Case, Variant: b.Cells[i].Variant, Model: b.Cells[i].Model}) <
			FullKey(matrix.Cell{Case: b.Cells[j].Case, Variant: b.Cells[j].Variant, Model: b.Cells[j].Model})
	})
	return b
}

// SaveBaseline writes a baseline file.
func SaveBaseline(path string, b *Baseline) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// LoadBaseline reads a baseline file.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &b, nil
}

// Regression is a cell that fell below its baseline beyond the noise.
type Regression struct {
	Cell     matrix.Cell
	Baseline string // "9/10 90% [60-98]"
	Now      string
}

// BaselineCheck is the result of checking a run against a baseline.
type BaselineCheck struct {
	Regressions []Regression
	// Missing are baseline cells the run did not cover.
	Missing []matrix.Cell
	// Checked counts the cells compared.
	Checked int
}

// CheckBaseline compares a run with a baseline: a cell regressed when the
// upper end of its interval is below the lower end of the baseline's.
func CheckBaseline(r *Run, b *Baseline) BaselineCheck {
	var out BaselineCheck
	have := map[matrix.Cell]*matrix.CellResult{}
	for _, c := range r.Cells {
		have[c.Cell] = c
	}
	for _, bc := range b.Cells {
		cell := matrix.Cell{Case: bc.Case, Variant: bc.Variant, Model: bc.Model}
		cur := have[cell]
		if cur == nil || cur.Passed+cur.Failed == 0 {
			out.Missing = append(out.Missing, cell)
			continue
		}
		out.Checked++
		blo, _ := matrix.Wilson(bc.Passed, bc.Runs)
		_, chi := matrix.Wilson(cur.Passed, cur.Passed+cur.Failed)
		if chi < blo {
			out.Regressions = append(out.Regressions, Regression{Cell: cell, Baseline: rateText(bc.Passed, bc.Runs), Now: RateText(cur)})
		}
	}
	return out
}

func rateText(k, n int) string {
	lo, hi := matrix.Wilson(k, n)
	return fmt.Sprintf("%d/%d %s [%s-%s]", k, n, pct(float64(k)/float64(n)), pct(lo), pct(hi))
}

// Format describes the check.
func (c BaselineCheck) Format() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d cell(s) checked against the baseline: %d regressed", c.Checked, len(c.Regressions))
	if len(c.Missing) > 0 {
		fmt.Fprintf(&b, ", %d not covered by this run", len(c.Missing))
	}
	b.WriteString("\n")
	for _, r := range c.Regressions {
		fmt.Fprintf(&b, "  REGRESSED %s: baseline %s, now %s\n", r.Cell, r.Baseline, r.Now)
	}
	for _, m := range c.Missing {
		fmt.Fprintf(&b, "  not covered: %s\n", m)
	}
	return b.String()
}
