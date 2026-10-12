package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	_ "github.com/fluid-movement/prep/internal/eval/assert/outcome"
	_ "github.com/fluid-movement/prep/internal/eval/assert/tools"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/judge"
	"github.com/fluid-movement/prep/internal/eval/matrix"
	"github.com/fluid-movement/prep/internal/eval/results"
	"github.com/fluid-movement/prep/internal/eval/variant"
)

func init() {
	evalCommands = append(evalCommands,
		evalCommand{"run", evalRun},
		evalCommand{"list", evalList},
		evalCommand{"doctor", evalDoctor},
		evalCommand{"judge-check", evalJudgeCheck},
	)
}

// evalTools are the flags every command that starts the agent shares.
type evalTools struct {
	plugin, claude, binary string
}

func (t *evalTools) flags(fs *flag.FlagSet) {
	fs.StringVar(&t.plugin, "plugin", "plugins/claude-code", "the shipped plugin directory (the `current` variant)")
	fs.StringVar(&t.claude, "claude", "", "the claude executable (default: found on PATH)")
	fs.StringVar(&t.binary, "binary", "", "the prep binary under test (default: this one)")
}

// resolve fills in the defaults and returns the variant resolver.
func (t *evalTools) resolve(dir string) (variant.Resolver, error) {
	if t.binary == "" {
		self, err := os.Executable()
		if err != nil {
			return variant.Resolver{}, err
		}
		t.binary = self
	}
	return variant.Resolver{Root: filepath.Join(dir, "variants"), PluginDir: t.plugin}, nil
}

// caseFilter holds the flags that select cases.
type caseFilter struct {
	areas, names, tags multi
	smoke              bool
}

func (f *caseFilter) flags(fs *flag.FlagSet) {
	fs.Var(&f.areas, "area", "only cases of this area (repeatable)")
	fs.Var(&f.names, "case", "only cases whose name matches this glob (repeatable)")
	fs.Var(&f.tags, "tag", "only cases with this tag (repeatable)")
	fs.BoolVar(&f.smoke, "smoke", false, "only the cheap smoke subset (tag smoke)")
}

func (f *caseFilter) selectFrom(dir string) ([]*evalcase.Case, error) {
	all, err := evalcase.Load(dir)
	if err != nil {
		return nil, err
	}
	tags := append([]string{}, f.tags...)
	if f.smoke {
		tags = append(tags, "smoke")
	}
	sel := evalcase.Select(all, evalcase.Filter{Areas: splitTags(f.areas), Names: splitTags(f.names), Tags: splitTags(tags)})
	if len(sel) == 0 {
		return nil, usageErr("no case matches (%d cases in %s)", len(all), dir)
	}
	return sel, nil
}

func csv(s string) []string { return splitTags([]string{s}) }

// evalRun runs the matrix and reports it.
func evalRun(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval run", flag.ContinueOnError)
	var tools evalTools
	var filter caseFilter
	tools.flags(fs)
	filter.flags(fs)
	variants := fs.String("variants", "current", "variants to run, comma-separated")
	models := fs.String("models", "sonnet", "models to run, comma-separated (an extended run names several)")
	n := fs.Int("n", 5, "runs per cell")
	parallel := fs.Int("parallel", 2, "runs at once")
	seed := fs.Int64("seed", 1, "seed for the order of runs")
	maxUSD := fs.Float64("max-usd", 0, "stop when this much has been spent (0: no limit)")
	sequential := fs.Bool("sequential", false, "stop a cell early once it clearly always passes or always fails")
	keep := fs.Bool("keep", false, "keep every run's workspace")
	yes := fs.Bool("yes", false, "run although the estimate exceeds --max-usd")
	dry := fs.Bool("dry-run", false, "print the cells and the cost estimate, run nothing")
	note := fs.String("note", "", "a note stored with the run")
	judgeModel := fs.String("judge-model", "sonnet", "the model that judges rubric items")
	noJudge := fs.Bool("no-judge", false, "leave rubric items unjudged")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("unexpected arguments: %s", strings.Join(pos, " "))
	}
	res, err := tools.resolve(dir)
	if err != nil {
		return err
	}
	cases, err := filter.selectFrom(dir)
	if err != nil {
		return err
	}
	opts := matrix.Options{Cases: cases, Variants: csv(*variants), Models: csv(*models), N: *n, Parallel: *parallel, Seed: *seed,
		Sequential: *sequential, MaxUSD: *maxUSD}
	cells := matrix.Cells(opts)
	if len(cells) == 0 {
		return usageErr("no case runs on this combination of variants and models")
	}
	store := resultsStore(dir)
	est := matrix.Estimator(recentCells(store, 5))
	opts.Estimate = est
	total := matrix.EstimateTotal(cells, *n, est)
	a.printf("%d case(s), %d cell(s) x %d run(s); estimated cost about $%.2f\n", len(cases), len(cells), *n, total)
	if *dry {
		for _, c := range cells {
			a.printf("  %s  ~$%.3f/run\n", c, est(c))
		}
		return nil
	}
	if *maxUSD > 0 && total > *maxUSD && !*yes {
		return usageErr("the estimate $%.2f exceeds --max-usd $%.2f; lower N or the cell count, or add --yes to run until the limit stops it", total, *maxUSD)
	}
	auth, err := env.AuthFromEnv(nil)
	if err != nil {
		return &domain.Error{Code: domain.ErrGate, Message: err.Error()}
	}

	parent, err := os.MkdirTemp("", "prep-eval-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(parent)
	id := results.NewID(time.Now())
	live := &matrix.Live{Cases: map[string]*evalcase.Case{}, Variants: res, Binary: tools.binary, Claude: tools.claude, Auth: auth,
		Parent: parent, OutDir: store.Dir(id), Keep: *keep}
	for _, c := range cases {
		live.Cases[c.Name] = c
	}
	if !*noJudge && hasRubric(cases) {
		j, cleanup, err := judge.NewClaude(judge.Config{Model: *judgeModel, Claude: tools.claude, Auth: auth, Parent: parent})
		if err != nil {
			return err
		}
		defer cleanup()
		live.Judge = j.Func()
	}
	opts.Runner = live
	var refs []results.VariantRef
	for _, v := range opts.Variants {
		m, err := res.Materialize(v, filepath.Join(parent, "hash-"+strings.ReplaceAll(v, "/", "_")))
		if err != nil {
			return err
		}
		refs = append(refs, results.VariantRef{Name: v, Hash: m.Hash})
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := 0
	opts.OnResult = func(r matrix.RunResult) {
		done++
		verdict := "FAIL"
		switch {
		case r.Status == "error":
			verdict = "ERROR"
		case r.Passed():
			verdict = "pass"
		}
		fmt.Fprintf(a.errw, "[%d/%d] %s run %d: %s (%s, $%.3f)\n", done, len(cells)**n, r.Cell, r.Index+1, verdict, r.Status, r.Cost.USD)
	}
	started := time.Now()
	out, err := matrix.Run(ctx, opts)
	if err != nil {
		return err
	}
	run := &results.Run{ID: id, Created: started, Note: *note,
		Definition: results.Definition{Cases: caseNames(cases), Variants: refs, Models: opts.Models, N: *n, Seed: *seed, Parallel: *parallel, Sequential: *sequential, MaxUSD: *maxUSD},
		Versions:   results.Versions{Prep: Version, Git: gitRev()},
		Cells:      out.Cells, StoppedBy: out.StoppedBy, SpentUSD: out.SpentUSD}
	for _, c := range out.Cells {
		for _, r := range c.Runs {
			if r.Harness != "" {
				run.Versions.Harness = r.Harness
			}
		}
	}
	if err := store.Save(run); err != nil {
		return err
	}
	if a.json {
		a.emit(run)
		return nil
	}
	a.printf("%s", results.Report(run, results.ReportOptions{Failures: true}))
	a.printf("\nstored as %s (prep eval report %s)\n", id, id)
	return nil
}

func recentCells(store results.Store, n int) []*matrix.CellResult {
	ids, _ := store.IDs()
	if len(ids) > n {
		ids = ids[:n]
	}
	var out []*matrix.CellResult
	for _, id := range ids {
		if r, err := store.Load(id); err == nil {
			out = append(out, r.Cells...)
		}
	}
	return out
}

func hasRubric(cases []*evalcase.Case) bool {
	for _, c := range cases {
		if len(c.Expect.Rubric) > 0 {
			return true
		}
	}
	return false
}

func caseNames(cases []*evalcase.Case) []string {
	var n []string
	for _, c := range cases {
		n = append(n, c.Name)
	}
	sort.Strings(n)
	return n
}

func gitRev() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// evalList lists the cases.
func evalList(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval list", flag.ContinueOnError)
	var filter caseFilter
	filter.flags(fs)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep eval list [--area a] [--case glob] [--tag t] [--smoke]")
	}
	cases, err := filter.selectFrom(dir)
	if err != nil {
		return err
	}
	if a.json {
		type row struct {
			Name        string   `json:"name"`
			Area        string   `json:"area"`
			Tags        []string `json:"tags,omitempty"`
			Models      []string `json:"models,omitempty"`
			Variants    []string `json:"variants,omitempty"`
			Description string   `json:"description,omitempty"`
		}
		var rows []row
		for _, c := range cases {
			rows = append(rows, row{c.Name, c.Area, c.Tags, c.Models, c.Variants, c.Description})
		}
		a.emit(map[string]any{"cases": rows})
		return nil
	}
	for _, c := range cases {
		tags := ""
		if len(c.Tags) > 0 {
			tags = " #" + strings.Join(c.Tags, " #")
		}
		a.printf("%-16s %-34s %s%s\n", c.Area, c.Name, c.Description, tags)
	}
	return nil
}

// evalDoctor runs the isolation leak check and a harness smoke run.
func evalDoctor(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval doctor", flag.ContinueOnError)
	var tools evalTools
	tools.flags(fs)
	variants := fs.String("variants", "current", "variants to check, comma-separated")
	model := fs.String("model", "haiku", "the model of the probe session")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep eval doctor [--variants v,...] [--model m]")
	}
	res, err := tools.resolve(dir)
	if err != nil {
		return err
	}
	auth, err := env.AuthFromEnv(nil)
	if err != nil {
		return &domain.Error{Code: domain.ErrGate, Message: err.Error()}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	bad := 0
	type finding struct {
		Variant string   `json:"variant"`
		OK      bool     `json:"ok"`
		Leaks   []string `json:"leaks,omitempty"`
		Error   string   `json:"error,omitempty"`
	}
	var found []finding
	for _, name := range csv(*variants) {
		f := finding{Variant: name}
		func() {
			tmp, err := os.MkdirTemp("", "prep-eval-doctor-")
			if err != nil {
				f.Error = err.Error()
				return
			}
			defer os.RemoveAll(tmp)
			v, err := res.Materialize(name, filepath.Join(tmp, "plugin"))
			if err != nil {
				f.Error = err.Error()
				return
			}
			e, err := env.New(env.Config{Binary: tools.binary, Claude: tools.claude, Auth: auth, Parent: tmp})
			if err != nil {
				f.Error = err.Error()
				return
			}
			defer e.Cleanup()
			meta, err := e.Probe(ctx, *model, v.PluginDir)
			if err != nil {
				f.Error = err.Error()
				return
			}
			f.Leaks = e.Leaks(meta, env.ExpectedFromPlugin(v.PluginDir))
			f.OK = len(f.Leaks) == 0
		}()
		if !f.OK {
			bad++
		}
		found = append(found, f)
	}
	if a.json {
		a.emit(map[string]any{"ok": bad == 0, "variants": found})
	} else {
		for _, f := range found {
			switch {
			case f.Error != "":
				a.printf("%-20s FAILED  %s\n", f.Variant, f.Error)
			case len(f.Leaks) > 0:
				a.printf("%-20s LEAKS   %s\n", f.Variant, strings.Join(f.Leaks, "; "))
			default:
				a.printf("%-20s ok      isolated: only the variant's plugin and skills are visible\n", f.Variant)
			}
		}
	}
	if bad > 0 {
		return &domain.Error{Code: domain.ErrGate, Message: fmt.Sprintf("%d variant(s) failed the isolation check", bad)}
	}
	return nil
}

// evalJudgeCheck measures the judge against the hand-labelled samples.
func evalJudgeCheck(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval judge-check", flag.ContinueOnError)
	var tools evalTools
	tools.flags(fs)
	samples := fs.String("samples", "", "calibration samples (default <dir>/judge/calibration)")
	repeats := fs.Int("repeats", 3, "judgements per sample")
	model := fs.String("model", "sonnet", "the judge model")
	threshold := fs.Float64("threshold", judge.DefaultThreshold, "agreement below which the judge is untrusted")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep eval judge-check [--samples dir] [--repeats N] [--model m]")
	}
	if *samples == "" {
		*samples = filepath.Join(dir, "judge", "calibration")
	}
	ss, err := judge.LoadSamples(*samples)
	if err != nil {
		return err
	}
	auth, err := env.AuthFromEnv(nil)
	if err != nil {
		return &domain.Error{Code: domain.ErrGate, Message: err.Error()}
	}
	parent, err := os.MkdirTemp("", "prep-eval-judge-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(parent)
	j, cleanup, err := judge.NewClaude(judge.Config{Model: *model, Claude: tools.claude, Auth: auth, Parent: parent})
	if err != nil {
		return err
	}
	defer cleanup()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	rep, err := j.Calibrate(ctx, ss, *repeats, *threshold)
	if err != nil {
		return err
	}
	if a.json {
		a.emit(rep)
	} else {
		verdict := "TRUSTED"
		if !rep.Trusted {
			verdict = "UNTRUSTED"
		}
		a.printf("%s: agreement %.0f%% (threshold %.0f%%), consistency %.0f%%, %d judgements, %d errored\n",
			verdict, rep.Agreement*100, *threshold*100, rep.Consistency*100, rep.Judgements, rep.Errored)
		for _, d := range rep.Disagreements {
			a.printf("  disagreed on %s\n", d)
		}
	}
	if !rep.Trusted {
		return &domain.Error{Code: domain.ErrGate, Message: "the judge is untrusted: its agreement with the labels is below the threshold"}
	}
	return nil
}
