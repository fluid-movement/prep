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
	"github.com/fluid-movement/prep/internal/eval/models"
	"github.com/fluid-movement/prep/internal/eval/results"
	"github.com/fluid-movement/prep/internal/eval/transcript"
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

// matrixFlags are the flags of every command that runs a matrix.
type matrixFlags struct {
	tools                 evalTools
	filter                caseFilter
	variants, models      string
	n, parallel           int
	seed                  int64
	maxUSD                float64
	sequential, keep, yes bool
	dry                   bool
	note, judgeModel      string
	noJudge               bool
}

func (m *matrixFlags) register(fs *flag.FlagSet, defVariants string) {
	m.tools.flags(fs)
	m.filter.flags(fs)
	fs.StringVar(&m.variants, "variants", defVariants, "variants to run, comma-separated")
	fs.StringVar(&m.models, "models", "sonnet", "models to run, comma-separated (an extended run names several)")
	fs.IntVar(&m.n, "n", 5, "runs per cell")
	fs.IntVar(&m.parallel, "parallel", 2, "runs at once")
	fs.Int64Var(&m.seed, "seed", 1, "seed for the order of runs")
	fs.Float64Var(&m.maxUSD, "max-usd", 0, "stop when this much has been spent (0: no limit)")
	fs.BoolVar(&m.sequential, "sequential", false, "stop a cell early once it clearly always passes or always fails")
	fs.BoolVar(&m.keep, "keep", false, "keep every run's workspace")
	fs.BoolVar(&m.yes, "yes", false, "run although the estimate exceeds --max-usd")
	fs.BoolVar(&m.dry, "dry-run", false, "print the cells and the cost estimate, run nothing")
	fs.StringVar(&m.note, "note", "", "a note stored with the run")
	fs.StringVar(&m.judgeModel, "judge-model", "sonnet", "the model that judges rubric items")
	fs.BoolVar(&m.noJudge, "no-judge", false, "leave rubric items unjudged")
}

// execMatrix selects the cases, prints the estimate, runs the matrix for the
// given variants and stores the run. It returns nil when it only estimated.
func execMatrix(a *app, dir string, m *matrixFlags, variants []string) (*results.Run, []*evalcase.Case, error) {
	res, err := m.tools.resolve(dir)
	if err != nil {
		return nil, nil, err
	}
	cases, err := m.filter.selectFrom(dir)
	if err != nil {
		return nil, nil, err
	}
	opts := matrix.Options{Cases: cases, Variants: variants, Models: csv(m.models), N: m.n, Parallel: m.parallel, Seed: m.seed,
		Sequential: m.sequential, MaxUSD: m.maxUSD}
	cells := matrix.Cells(opts)
	if len(cells) == 0 {
		return nil, nil, usageErr("no case runs on this combination of variants and models")
	}
	store := resultsStore(dir)
	est := matrix.Estimator(recentCells(store, 5))
	opts.Estimate = est
	total := matrix.EstimateTotal(cells, m.n, est)
	a.printf("%d case(s), %d cell(s) x %d run(s); estimated cost about $%.2f\n", len(cases), len(cells), m.n, total)
	if m.dry {
		for _, c := range cells {
			a.printf("  %s  ~$%.3f/run\n", c, est(c))
		}
		return nil, cases, nil
	}
	if m.maxUSD > 0 && total > m.maxUSD && !m.yes {
		return nil, nil, usageErr("the estimate $%.2f exceeds --max-usd $%.2f; lower N or the cell count, or add --yes to run until the limit stops it", total, m.maxUSD)
	}
	auth, err := env.AuthFromEnv(nil)
	if err != nil {
		return nil, nil, &domain.Error{Code: domain.ErrGate, Message: err.Error()}
	}
	// scratch holds what is thrown away; work holds the run environments, of
	// which a failing run's is kept for inspection. Both live outside any git
	// repository, so the fixtures do not stage into this checkout.
	scratch, err := os.MkdirTemp("", "prep-eval-scratch-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(scratch)
	work, err := os.MkdirTemp("", "prep-eval-work-")
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(work) // only when empty: nothing was kept
	id := results.NewID(time.Now())
	live := &matrix.Live{Cases: map[string]*evalcase.Case{}, Variants: res, Binary: m.tools.binary, Claude: m.tools.claude, Auth: auth,
		Parent: work, OutDir: store.Dir(id), Keep: m.keep}
	for _, c := range cases {
		live.Cases[c.Name] = c
	}
	if !m.noJudge && hasRubric(cases) {
		j, cleanup, err := judge.NewClaude(judge.Config{Model: m.judgeModel, Claude: m.tools.claude, Auth: auth, Parent: scratch})
		if err != nil {
			return nil, nil, err
		}
		defer cleanup()
		live.Judge = j.Func()
	}
	opts.Runner = live
	var refs []results.VariantRef
	for _, v := range variants {
		mat, err := res.Materialize(v, filepath.Join(scratch, "hash-"+strings.ReplaceAll(v, "/", "_")))
		if err != nil {
			return nil, nil, err
		}
		refs = append(refs, results.VariantRef{Name: v, Hash: mat.Hash})
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
		fmt.Fprintf(a.errw, "[%d/%d] %s run %d: %s (%s, $%.3f)\n", done, len(cells)*m.n, r.Cell, r.Index+1, verdict, r.Status, r.Cost.USD)
		if r.Status == "error" {
			fmt.Fprintf(a.errw, "    %s\n", r.Err)
		}
	}
	started := time.Now()
	out, err := matrix.Run(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	run := &results.Run{ID: id, Created: started, Note: m.note,
		Definition: results.Definition{Cases: caseNames(cases), Areas: caseAreas(cases), Variants: refs, Models: opts.Models, N: m.n, Seed: m.seed, Parallel: m.parallel, Sequential: m.sequential, MaxUSD: m.maxUSD},
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
		return nil, nil, err
	}
	return run, cases, nil
}

// evalRun runs the matrix and reports it.
func evalRun(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval run", flag.ContinueOnError)
	var m matrixFlags
	m.register(fs, "current")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("unexpected arguments: %s", strings.Join(pos, " "))
	}
	run, _, err := execMatrix(a, dir, &m, csv(m.variants))
	if err != nil || run == nil {
		return err
	}
	if a.json {
		a.emit(run)
		return nil
	}
	a.printf("%s", results.Report(run, results.ReportOptions{Failures: true}))
	a.printf("\nstored as %s (prep eval report %s)\n", run.ID, run.ID)
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

func caseAreas(cases []*evalcase.Case) map[string]string {
	m := map[string]string{}
	for _, c := range cases {
		m[c.Name] = c.Area
	}
	return m
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
	// probe starts a session with a variant and says what it could see.
	probe := func(name string, bundled []string) (meta transcript.Meta, leaks []string, err error) {
		tmp, err := os.MkdirTemp("", "prep-eval-doctor-")
		if err != nil {
			return meta, nil, err
		}
		defer os.RemoveAll(tmp)
		v, err := res.Materialize(name, filepath.Join(tmp, "plugin"))
		if err != nil {
			return meta, nil, err
		}
		e, err := env.New(env.Config{Binary: tools.binary, Claude: tools.claude, Auth: auth, Parent: tmp})
		if err != nil {
			return meta, nil, err
		}
		defer e.Cleanup()
		if meta, err = e.Probe(ctx, *model, v.PluginDir); err != nil {
			return meta, nil, err
		}
		want := env.ExpectedFromPlugin(v.PluginDir)
		want.Bundled = bundled
		return meta, e.Leaks(meta, want), nil
	}
	// The skills the harness bundles are what a session with no plugin sees.
	var bundled []string
	if meta, _, err := probe("none", nil); err == nil {
		bundled = meta.Skills
	}
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
		if _, leaks, err := probe(name, bundled); err != nil {
			f.Error = err.Error()
		} else {
			f.Leaks, f.OK = leaks, len(leaks) == 0
		}
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

func init() {
	evalCommands = append(evalCommands, evalCommand{"models", evalModels})
}

// evalModels says which model is enough for which area, from a stored run
// that covered several models.
func evalModels(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval models", flag.ContinueOnError)
	variantName := fs.String("variant", "", "the variant to analyse (default: the first by name)")
	minRate := fs.Float64("min-rate", 0.9, "pass rate a model needs in an area")
	minLower := fs.Float64("min-lower", 0.8, "lower bound of the 95% interval a model needs")
	minRuns := fs.Int("min-runs", 5, "fewest runs in an area that can show anything")
	md := fs.Bool("md", false, "print the decision table as markdown (for eval/models.md)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ids := pos
	if len(ids) == 0 {
		ids = []string{"latest"}
	}
	run, err := resultsStore(dir).LoadAll(ids...)
	if err != nil {
		return err
	}
	an, err := models.Analyze(run.ID, run.Versions.Harness, run.Definition.N, run.Cells, run.Definition.Areas,
		models.Options{Variant: *variantName, MinRate: *minRate, MinLower: *minLower, MinRuns: *minRuns})
	if err != nil {
		return err
	}
	switch {
	case a.json:
		a.emit(an)
	case *md:
		a.printf("%s", models.Markdown(an))
	default:
		a.printf("%s", models.Text(an))
	}
	return nil
}
