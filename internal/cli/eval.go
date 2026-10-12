package cli

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/results"
)

// evalUsage is the one-line help of prep eval; the subcommands are listed in
// evalCommands.
const evalUsage = "eval run|list|doctor|report|compare|models|mutate|check|judge-check|prune  agent behavior tests: run prompts against fake projects and report how agents used prep (see the eval knowledge entry)"

// evalCommand is one prep eval subcommand.
type evalCommand struct {
	name string
	run  func(a *app, dir string, args []string) error
}

var evalCommands []evalCommand

func init() {
	evalCommands = []evalCommand{
		{"report", evalReport},
		{"compare", evalCompare},
		{"prune", evalPrune},
	}
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
	id := "latest"
	if len(pos) > 1 {
		return usageErr("usage: prep eval report [<run-id>|latest] [--failures]")
	} else if len(pos) == 1 {
		id = pos[0]
	}
	run, err := resultsStore(dir).Load(id)
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
