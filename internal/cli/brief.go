package cli

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// handoffPath returns the absolute path of a hand-off file of an issue
// (mdstore.HandoffBrief, HandoffReport and their check variants) under
// .prep/local/handoff/<id>/ of the opened project.
func (a *app) handoffPath(id, name string) string { return a.store.HandoffPath(id, name) }

// briefOut takes --out out of the arguments of prep brief. The flag's value
// is optional, which the flag package cannot express: --out=<file> names the
// file, and --out <file> does when another argument is left to be the issue.
// A bare --out asks for the issue's default place.
func briefOut(args []string) (rest []string, set bool, file string) {
	for k := 0; k < len(args); k++ {
		arg := args[k]
		name, val, has := strings.Cut(arg, "=")
		switch {
		case arg == "--":
			return append(rest, args[k:]...), set, file
		case name != "--out" && name != "-out":
			rest = append(rest, arg)
		case has:
			set, file = true, val
		default:
			set, file = true, ""
			if k+1 < len(args) && !strings.HasPrefix(args[k+1], "-") && positionals(rest)+positionals(args[k+2:]) > 0 {
				k++
				file = args[k]
			}
		}
	}
	return rest, set, file
}

// positionals counts the arguments that are not flags; prep brief has only
// boolean flags besides --out.
func positionals(args []string) int {
	n := 0
	for _, a := range args {
		if a == "--" {
			break
		}
		if !strings.HasPrefix(a, "-") {
			n++
		}
	}
	return n
}

// cmdBrief prints the hand-off text for an implementer, or with --check for
// a checker: everything the agent needs, with no prep command in it.
func cmdBrief(a *app, args []string) error {
	args, toFile, file := briefOut(args)
	fs := flag.NewFlagSet("brief", flag.ContinueOnError)
	check := fs.Bool("check", false, "the checker's brief")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("usage: prep brief <id> [--check] [--out [file]]")
	}
	t, err := a.load()
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	briefName, reportName := mdstore.HandoffFiles(*check)
	b, err := t.BuildBrief(id, *check, a.handoffPath(id, reportName))
	if err != nil {
		return err
	}
	a.noteRead(issueEvent(t, "brief", id, "read the "+b.Role+"'s brief"))
	// The report's directory is made here so the agent can write the file
	// with any tool. Failing to make it only matters when the brief goes there.
	mkErr := os.MkdirAll(filepath.Dir(b.ReportPath), 0o755)
	if !toFile {
		if a.json {
			a.emit(b)
			return nil
		}
		a.printf("%s", b.Text())
		return nil
	}
	if file == "" {
		if mkErr != nil {
			return mkErr
		}
		file = a.handoffPath(id, briefName)
	}
	if err := os.WriteFile(file, []byte(b.Text()), 0o644); err != nil {
		return err
	}
	if a.json {
		a.emit(struct {
			OK   bool   `json:"ok"`
			Path string `json:"path"`
			*domain.Brief
		}{true, file, b})
		return nil
	}
	a.printf("%s\n", file)
	return nil
}
