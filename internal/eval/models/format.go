package models

import (
	"fmt"
	"sort"
	"strings"
)

func pct(f float64) string { return fmt.Sprintf("%.0f%%", f*100) }

func usd(f float64) string {
	if f == 0 {
		return "-"
	}
	return fmt.Sprintf("$%.3f", f)
}

func (s Stat) rate() string {
	if s.Runs == 0 {
		return "no runs"
	}
	return fmt.Sprintf("%d/%d %s [%s-%s]", s.Passed, s.Runs, pct(s.Rate), pct(s.Lo), pct(s.Hi))
}

// wrong summarizes the top failures in a few words.
func wrong(fs []Failure, max int) string {
	if len(fs) == 0 {
		return "-"
	}
	var parts []string
	for i, f := range fs {
		if i == max {
			parts = append(parts, fmt.Sprintf("+%d more", len(fs)-max))
			break
		}
		parts = append(parts, fmt.Sprintf("%s (%s, %dx)", f.Desc, f.Group, f.Count))
	}
	return strings.Join(parts, "; ")
}

func decision(a Area) (cheapest, rate, perPass string) {
	cheapest, rate, perPass = "-", "-", "-"
	switch a.Status {
	case Decided:
		cheapest = a.Cheapest
		for _, s := range a.Models {
			if s.Model == a.Cheapest {
				rate, perPass = s.rate(), usd(s.USDPerPass)
			}
		}
	case NotEnough:
		cheapest = "not enough data"
	default:
		cheapest = "none sufficient"
	}
	return
}

// Text prints the analysis for a terminal.
func Text(a *Analysis) string {
	var b strings.Builder
	fmt.Fprintf(&b, "model comparison of run %s  variant %s  N=%d", a.RunID, a.Variant, a.N)
	if a.Harness != "" {
		fmt.Fprintf(&b, "  harness %s", a.Harness)
	}
	fmt.Fprintf(&b, "\nenough = pass rate >= %s and interval lower bound >= %s, at least %d runs in the area\n", pct(a.Options.MinRate), pct(a.Options.MinLower), a.Options.MinRuns)
	for _, w := range a.Warnings {
		fmt.Fprintf(&b, "warning: %s\n", w)
	}
	b.WriteString("\nDECISION\n")
	for _, ar := range a.Areas {
		cheapest, rate, perPass := decision(ar)
		fmt.Fprintf(&b, "  %-16s %-16s %-24s %-9s next up: %s\n", ar.Area, cheapest, rate, perPass, orDash(ar.NextUp))
		if ar.Hint != "" {
			fmt.Fprintf(&b, "  %-16s %s\n", "", ar.Hint)
		}
		if ar.Cheapest != "" && len(ar.WrongBelow) > 0 {
			fmt.Fprintf(&b, "  %-16s cheaper models get wrong: %s\n", "", wrong(ar.WrongBelow, 3))
		}
	}
	b.WriteString("\nBY MODEL AND AREA\n")
	for _, ar := range a.Areas {
		fmt.Fprintf(&b, "  %s\n", ar.Area)
		for _, s := range ar.Models {
			mark := ""
			if s.Sufficient {
				mark = "  enough"
			}
			ids := ""
			if len(s.Resolved) > 0 {
				ids = "  (" + strings.Join(s.Resolved, ", ") + ")"
			}
			fmt.Fprintf(&b, "    %-10s %-24s %s/run  %s/pass%s%s\n", s.Model, s.rate(), usd(s.USDPerRun), usd(s.USDPerPass), mark, ids)
			for i, f := range s.Failures {
				if i == 4 {
					fmt.Fprintf(&b, "        +%d more failing assertions\n", len(s.Failures)-4)
					break
				}
				fmt.Fprintf(&b, "        fails %dx [%s] %s\n", f.Count, f.Group, f.Desc)
				for _, e := range f.Examples {
					fmt.Fprintf(&b, "            see %s\n", e)
				}
			}
		}
	}
	b.WriteString("\nBY CASE\n")
	for _, ar := range a.Areas {
		for _, c := range ar.Cases {
			var parts []string
			for _, s := range c.Models {
				parts = append(parts, fmt.Sprintf("%s %s", s.Model, pct(s.Rate)))
			}
			fmt.Fprintf(&b, "  %-34s %s\n", c.Case, strings.Join(parts, "   "))
		}
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Markdown prints the decision table for committing (eval/models.md).
func Markdown(a *Analysis) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Which model is enough for which task\n\nRun `%s`, variant `%s`, %d runs per cell", a.RunID, a.Variant, a.N)
	if a.Harness != "" {
		fmt.Fprintf(&b, ", Claude Code %s", a.Harness)
	}
	fmt.Fprintf(&b, ".\nEnough means a pass rate of at least %s with the lower bound of the 95%% interval at least %s.\n\n", pct(a.Options.MinRate), pct(a.Options.MinLower))
	var ids []string
	seen := map[string]bool{}
	for _, ar := range a.Areas {
		for _, s := range ar.Models {
			for _, r := range s.Resolved {
				if k := s.Model + "=" + r; !seen[k] {
					seen[k] = true
					ids = append(ids, "`"+k+"`")
				}
			}
		}
	}
	sort.Strings(ids)
	if len(ids) > 0 {
		fmt.Fprintf(&b, "Model ids: %s.\n\n", strings.Join(ids, ", "))
	}
	for _, w := range a.Warnings {
		fmt.Fprintf(&b, "> Warning: %s\n", w)
	}
	b.WriteString("| Area | Cheapest enough | Pass rate | Cost per pass | Next up | What cheaper models get wrong |\n|---|---|---|---|---|---|\n")
	for _, ar := range a.Areas {
		cheapest, rate, perPass := decision(ar)
		w := "-"
		if ar.Cheapest != "" {
			w = wrong(ar.WrongBelow, 3)
		} else if ar.Hint != "" {
			w = ar.Hint
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", ar.Area, cheapest, rate, perPass, orDash(ar.NextUp), strings.ReplaceAll(w, "|", "\\|"))
	}
	return b.String()
}
