package matrix

import "strings"

// DefaultCostByModel is the guessed cost of one run when nothing is known
// about a cell: dollars (at API prices) by model family, from the first real
// runs (sonnet averaged $0.043); later runs replace the guess with history.
var DefaultCostByModel = map[string]float64{"haiku": 0.02, "sonnet": 0.05, "opus": 0.15}

// unknownCost is used for a model of no known family.
const unknownCost = 0.3

// Estimator returns the expected cost of one run of a cell from earlier
// results: the cell's own mean, else the mean over earlier runs of the same
// model, else the default for the model's family.
func Estimator(history []*CellResult) func(Cell) float64 {
	own := map[Cell]float64{}
	sum, cnt := map[string]float64{}, map[string]int{}
	for _, c := range history {
		if c.Cost.USD.Mean <= 0 {
			continue
		}
		own[c.Cell] = c.Cost.USD.Mean
		sum[c.Cell.Model] += c.Cost.USD.Mean
		cnt[c.Cell.Model]++
	}
	return func(c Cell) float64 {
		if v, ok := own[c]; ok {
			return v
		}
		if n := cnt[c.Model]; n > 0 {
			return sum[c.Model] / float64(n)
		}
		for family, v := range DefaultCostByModel {
			if strings.Contains(strings.ToLower(c.Model), family) {
				return v
			}
		}
		return unknownCost
	}
}

// EstimateTotal is the expected cost of running every cell n times.
func EstimateTotal(cells []Cell, n int, est func(Cell) float64) float64 {
	var t float64
	for _, c := range cells {
		t += est(c) * float64(n)
	}
	return t
}
