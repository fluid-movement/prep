package results

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// ErrorGroup is one kind of failure of a prep command, over all sessions.
type ErrorGroup struct {
	Command string `json:"command"`
	Message string `json:"message"`
	Count   int    `json:"count"`
	// Example is a command line that failed this way.
	Example string `json:"example"`
	// Cases lists the cases it happened in.
	Cases []string `json:"cases"`
}

// ErrorReport says where agents wasted calls on prep.
type ErrorReport struct {
	Sessions int `json:"sessions"`
	// PrepCalls and Failed count the prep commands run and the ones that
	// failed (a command in a chained shell call fails when the call did).
	PrepCalls int          `json:"prep_calls"`
	Failed    int          `json:"failed"`
	Groups    []ErrorGroup `json:"groups"`
	// ByCommand maps a command to "calls/failed".
	ByCommand map[string][2]int `json:"by_command"`
	// MeanCalls is the mean number of tool calls per case, largest first.
	MeanCalls []CaseCalls `json:"mean_calls"`
}

// CaseCalls is a case with the mean number of tool calls it took.
type CaseCalls struct {
	Case string  `json:"case"`
	Mean float64 `json:"mean"`
}

var (
	idPattern  = regexp.MustCompile(`\b[0-9A-HJKMNP-TV-Z]{26}\b`)
	numPattern = regexp.MustCompile(`\b\d+\b`)
	exitPrefix = regexp.MustCompile(`^Exit code \d+\s*`)
)

// normalize makes error messages of the same kind equal.
func normalize(msg string) string {
	msg = strings.TrimSpace(exitPrefix.ReplaceAllString(msg, ""))
	if i := strings.Index(msg, "\n"); i >= 0 {
		msg = msg[:i]
	}
	msg = idPattern.ReplaceAllString(msg, "<id>")
	msg = numPattern.ReplaceAllString(msg, "N")
	if len(msg) > 140 {
		msg = msg[:140] + "…"
	}
	return msg
}

// Errors reads the transcripts stored under a run's directory and groups the
// failures of prep commands: every group is a message the CLI could make
// clearer or a flag it could accept.
func Errors(runDirs ...string) (*ErrorReport, error) {
	rep := &ErrorReport{ByCommand: map[string][2]int{}}
	groups := map[string]*ErrorGroup{}
	calls := map[string][]int{}
	for _, dir := range runDirs {
		files, _ := filepath.Glob(filepath.Join(dir, "runs", "*", "*", "*", "*", "transcript.json"))
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			var t transcript.Transcript
			if err := json.Unmarshal(data, &t); err != nil {
				return nil, fmt.Errorf("%s: %w", f, err)
			}
			rel, _ := filepath.Rel(filepath.Join(dir, "runs"), f)
			caseName := strings.Split(rel, string(filepath.Separator))[0]
			rep.Sessions++
			calls[caseName] = append(calls[caseName], len(t.Steps))
			for _, s := range t.Steps {
				for _, c := range s.Commands {
					rep.PrepCalls++
					bc := rep.ByCommand[c.Name]
					bc[0]++
					if s.Error {
						rep.Failed++
						bc[1]++
						msg := normalize(s.Result)
						key := c.Name + "\x00" + msg
						g := groups[key]
						if g == nil {
							g = &ErrorGroup{Command: c.Name, Message: msg, Example: c.Raw}
							groups[key] = g
						}
						g.Count++
						if len(g.Cases) == 0 || g.Cases[len(g.Cases)-1] != caseName {
							g.Cases = append(g.Cases, caseName)
						}
					}
					rep.ByCommand[c.Name] = bc
				}
			}
		}
	}
	for _, g := range groups {
		rep.Groups = append(rep.Groups, *g)
	}
	sort.Slice(rep.Groups, func(i, j int) bool {
		if rep.Groups[i].Count != rep.Groups[j].Count {
			return rep.Groups[i].Count > rep.Groups[j].Count
		}
		return rep.Groups[i].Message < rep.Groups[j].Message
	})
	for c, ns := range calls {
		sum := 0
		for _, n := range ns {
			sum += n
		}
		rep.MeanCalls = append(rep.MeanCalls, CaseCalls{c, float64(sum) / float64(len(ns))})
	}
	sort.Slice(rep.MeanCalls, func(i, j int) bool { return rep.MeanCalls[i].Mean > rep.MeanCalls[j].Mean })
	return rep, nil
}

// FormatErrors prints the report.
func FormatErrors(r *ErrorReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d session(s), %d prep command(s), %d failed\n", r.Sessions, r.PrepCalls, r.Failed)
	if len(r.Groups) == 0 {
		b.WriteString("\nno prep command failed\n")
	} else {
		b.WriteString("\nFAILURES, most frequent first (each one is a message to make clearer or a flag to accept)\n")
		for _, g := range r.Groups {
			fmt.Fprintf(&b, "  %3dx prep %s: %s\n", g.Count, g.Command, g.Message)
			fmt.Fprintf(&b, "        e.g. %s\n", strings.Join(strings.Fields(g.Example), " "))
			fmt.Fprintf(&b, "        in: %s\n", strings.Join(g.Cases, ", "))
		}
	}
	var names []string
	for n := range r.ByCommand {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return r.ByCommand[names[i]][0] > r.ByCommand[names[j]][0] })
	b.WriteString("\nCOMMANDS (calls / failed)\n")
	for i, n := range names {
		if i == 12 {
			break
		}
		fmt.Fprintf(&b, "  %-12s %4d / %d\n", n, r.ByCommand[n][0], r.ByCommand[n][1])
	}
	b.WriteString("\nMOST TOOL CALLS PER SESSION\n")
	for i, c := range r.MeanCalls {
		if i == 6 {
			break
		}
		fmt.Fprintf(&b, "  %-36s %.1f\n", c.Case, c.Mean)
	}
	return b.String()
}
