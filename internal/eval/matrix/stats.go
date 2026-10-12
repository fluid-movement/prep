package matrix

import (
	"math"
	"time"
)

// wilson returns the 95% Wilson score interval for k successes in n trials.
// With no trials it is the whole range.
func wilson(k, n int) (lo, hi float64) {
	if n == 0 {
		return 0, 1
	}
	const z = 1.96
	p := float64(k) / float64(n)
	nn := float64(n)
	den := 1 + z*z/nn
	center := (p + z*z/(2*nn)) / den
	half := z * math.Sqrt(p*(1-p)/nn+z*z/(4*nn*nn)) / den
	lo, hi = center-half, center+half
	if lo < 0 {
		lo = 0
	}
	if hi > 1 {
		hi = 1
	}
	return lo, hi
}

// Stat is a mean and a spread.
type Stat struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
}

func stat(xs []float64) Stat {
	if len(xs) == 0 {
		return Stat{}
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	if len(xs) == 1 {
		return Stat{Mean: mean}
	}
	var ss float64
	for _, x := range xs {
		ss += (x - mean) * (x - mean)
	}
	return Stat{Mean: mean, Std: math.Sqrt(ss / float64(len(xs)-1))}
}

// Cost is what one run spent.
type Cost struct {
	Calls     int           `json:"calls"`
	Errors    int           `json:"errors"`
	TokensIn  int           `json:"tokens_in"`
	TokensOut int           `json:"tokens_out"`
	USD       float64       `json:"usd"`
	Duration  time.Duration `json:"duration"`
}

// CostStats is the mean and spread of a cell's runs.
type CostStats struct {
	Calls     Stat `json:"calls"`
	Errors    Stat `json:"errors"`
	TokensIn  Stat `json:"tokens_in"`
	TokensOut Stat `json:"tokens_out"`
	USD       Stat `json:"usd"`
	Seconds   Stat `json:"seconds"`
}

func costStats(cs []Cost) CostStats {
	col := func(f func(Cost) float64) Stat {
		xs := make([]float64, len(cs))
		for i, c := range cs {
			xs[i] = f(c)
		}
		return stat(xs)
	}
	return CostStats{
		Calls:     col(func(c Cost) float64 { return float64(c.Calls) }),
		Errors:    col(func(c Cost) float64 { return float64(c.Errors) }),
		TokensIn:  col(func(c Cost) float64 { return float64(c.TokensIn) }),
		TokensOut: col(func(c Cost) float64 { return float64(c.TokensOut) }),
		USD:       col(func(c Cost) float64 { return c.USD }),
		Seconds:   col(func(c Cost) float64 { return c.Duration.Seconds() }),
	}
}
