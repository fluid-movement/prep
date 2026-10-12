package cli

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/eval/results"
)

// evalUsage is the one-line help of prep eval; the subcommands are listed in
// evalCommands.
const evalUsage = "eval run|list|doctor|report|compare|models|errors|mutate|check|judge-check|prune  agent behavior tests: run prompts against fake projects and report how agents used prep (see the eval knowledge entry)"

// evalCommand is one prep eval subcommand.
type evalCommand struct {
	name string
	run  func(a *app, dir string, args []string) error
}

var evalCommands []evalCommand

func init() {
	evalCommands = append(evalCommands,
		evalCommand{"report", evalReport},
		evalCommand{"compare", evalCompare},
		evalCommand{"prune", evalPrune},
		evalCommand{"check", evalCheck},
		evalCommand{"errors", evalErrors},
	)
}

// cmdEval dispatches prep eval <subcommand>. It needs no prep project: it
// works on the eval directory (default ./eval) of the prep checkout.
func cmdEval(a *app, args []string) error {
	// --dir may stand before the subcommand or after it.
	dir := "eval"
	for i := 0; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			dir = args[i+1]
			args = append(args[:i:i], args[i+2:]...)
			break
		}
		if strings.HasPrefix(args[i], "--dir=") {
			dir = strings.TrimPrefix(args[i], "--dir=")
			args = append(args[:i:i], args[i+1:]...)
			break
		}
	}
	if len(args) == 0 {
		return usageErr("usage: prep %s", evalUsage)
	}
	for _, c := range evalCommands {
		if c.name == args[0] {
			return c.run(a, dir, args[1:])
		}
	}
	return usageErr("unknown eval command %q; usage: prep %s", args[0], evalUsage)
}

func resultsStore(dir string) results.Store {
	return results.Store{Root: filepath.Join(dir, "results")}
}

// evalReport prints a stored run.
func evalReport(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval report", flag.ContinueOnError)
	failures := fs.Bool("failures", false, "list failing assertions and the run directories to inspect")
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
	if a.json {
		a.emit(run)
		return nil
	}
	a.printf("%s", results.Report(run, results.ReportOptions{Failures: *failures}))
	return nil
}

// evalCompare lines two stored runs up, or two variants of one run.
func evalCompare(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval compare", flag.ContinueOnError)
	variants := fs.String("variants", "", "compare two variants of one run: base,other")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	store := resultsStore(dir)
	if *variants != "" {
		parts := strings.Split(*variants, ",")
		if len(pos) != 1 || len(parts) != 2 {
			return usageErr("usage: prep eval compare <run-id> --variants base,other")
		}
		run, err := store.Load(pos[0])
		if err != nil {
			return err
		}
		ds := results.Compare(results.ForVariant(run.Cells, parts[0]), results.ForVariant(run.Cells, parts[1]), results.CaseModelKey)
		if a.json {
			a.emit(map[string]any{"a": parts[0], "b": parts[1], "deltas": ds})
			return nil
		}
		a.printf("%s", results.FormatCompare(run.ID+" variant "+parts[0], run.ID+" variant "+parts[1], ds))
		return nil
	}
	if len(pos) != 2 {
		return usageErr("usage: prep eval compare <run-a> <run-b> | <run-id> --variants base,other")
	}
	ra, err := store.Load(pos[0])
	if err != nil {
		return err
	}
	rb, err := store.Load(pos[1])
	if err != nil {
		return err
	}
	ds := results.Compare(ra.Cells, rb.Cells, results.FullKey)
	if a.json {
		a.emit(map[string]any{"a": ra.ID, "b": rb.ID, "deltas": ds})
		return nil
	}
	a.printf("%s", results.FormatCompare(ra.ID, rb.ID, ds))
	return nil
}

// evalCheck compares a stored run with the committed baseline and fails on a
// regression beyond the noise; --update rewrites the baseline from the run.
func evalCheck(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval check", flag.ContinueOnError)
	baselinePath := fs.String("baseline", "", "the baseline file (default <dir>/baseline.json)")
	update := fs.Bool("update", false, "write the run's pass rates as the new baseline")
	note := fs.String("note", "", "a note stored in a new baseline")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	id := "latest"
	if len(pos) > 1 {
		return usageErr("usage: prep eval check [<run-id>|latest] [--baseline file] [--update]")
	} else if len(pos) == 1 {
		id = pos[0]
	}
	if *baselinePath == "" {
		*baselinePath = filepath.Join(dir, "baseline.json")
	}
	run, err := resultsStore(dir).Load(id)
	if err != nil {
		return err
	}
	if *update {
		if err := results.SaveBaseline(*baselinePath, results.NewBaseline(run, *note)); err != nil {
			return err
		}
		a.printf("wrote %s from run %s\n", *baselinePath, run.ID)
		return nil
	}
	b, err := results.LoadBaseline(*baselinePath)
	if err != nil {
		return err
	}
	chk := results.CheckBaseline(run, b)
	if a.json {
		a.emit(map[string]any{"run": run.ID, "baseline": b.Run, "checked": chk.Checked, "regressions": chk.Regressions, "missing": chk.Missing})
	} else {
		a.printf("%s", chk.Format())
	}
	if len(chk.Regressions) > 0 {
		return silentErr{&domain.Error{Code: domain.ErrCheck, Message: fmt.Sprintf("%d cell(s) regressed against the baseline", len(chk.Regressions))}}
	}
	return nil
}

// evalErrors lists the prep commands that failed in stored runs.
func evalErrors(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval errors", flag.ContinueOnError)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ids := pos
	if len(ids) == 0 {
		ids = []string{"latest"}
	}
	store := resultsStore(dir)
	var dirs []string
	for _, id := range ids {
		r, err := store.Load(id)
		if err != nil {
			return err
		}
		dirs = append(dirs, store.Dir(r.ID))
	}
	rep, err := results.Errors(dirs...)
	if err != nil {
		return err
	}
	if a.json {
		a.emit(rep)
		return nil
	}
	a.printf("%s", results.FormatErrors(rep))
	return nil
}

// evalPrune removes old stored runs.
func evalPrune(a *app, dir string, args []string) error {
	fs := flag.NewFlagSet("eval prune", flag.ContinueOnError)
	keep := fs.Int("keep", 10, "runs to keep, newest first")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep eval prune [--keep N]")
	}
	removed, err := resultsStore(dir).Prune(*keep)
	if err != nil {
		return err
	}
	if a.json {
		a.emit(map[string]any{"removed": removed})
		return nil
	}
	a.printf("removed %d stored run(s), kept up to %d\n", len(removed), *keep)
	return nil
}
