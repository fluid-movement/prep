// Package cli is the agent interface: one command per read or write, every
// read with --json, errors as JSON with stable codes.
package cli

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
)

// Version is set at build time; go install builds fall back to the module
// version Go records in the binary.
var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		Version = bi.Main.Version
	}
}

//go:embed skill.md
var skillText string

// Exit codes.
const (
	exitOK       = 0
	exitFail     = 1 // gate unmet, check errors, not found
	exitUsage    = 2
	exitConflict = 3
)

type app struct {
	out, errw io.Writer
	stdin     io.Reader
	json      bool
	root      string
	actor     string
	now       func() time.Time

	store  *mdstore.Store
	kstore *okf.Store
}

type command struct {
	name  string
	write bool
	usage string
	run   func(a *app, args []string) error
}

var commands []command

func init() {
	commands = []command{
		{"init", true, "init [--no-bootstrap]      create .prep and the knowledge base bootstrap issues", cmdInit},
		{"prime", false, "prime [--max N]            session-start briefing: parents, actionable count, alerts", cmdPrime},
		{"guide", false, "guide <id>                 step contract: state, transitions, unmet gates, inputs, outputs", cmdGuide},
		{"list", false, "list [query flags]         query issues: --state --kind --tag --under --stale --actionable --blocked --parent --leaf --top --text --tree --view", cmdList},
		{"next", false, "next [--under <id>]        actionable issues: ready, not stale, dependencies done, unclaimed", cmdNext},
		{"show", false, "show <id>                  read one issue with its derived state", cmdShow},
		{"check", false, "check [--no-drift]         validate the whole tree", cmdCheck},
		{"watch", false, "watch                      stream a line per change under .prep until killed (for harness integrations)", cmdWatch},
		{"views", false, "views                      saved views from config.yaml", cmdViews},
		{"new", true, "new --title T --kind K     create an issue [--parent id] [--depends-on id]... [--tag t]... [--body text | --body-file path|-]", cmdNew},
		{"edit", true, "edit <id>                  change [--title T] [--kind K] [--parent id|''] [--depends-on id|'']... [--tag t|'']... [--body text | --body-file path|-]", cmdEdit},
		{"context", true, "context <id>               replace the implementation context: --body text | --body-file path|-", recordCmd(domain.OpContext)},
		{"decide", true, "decide <id> --title T      append a decision [--body text | --body-file path|-] [--supersedes D<n>] [--outcome]", recordCmd(domain.OpDecide)},
		{"criterion", true, "criterion <id>             acceptance criteria: --add text, --check n, --uncheck n, --remove n (repeatable)", recordCmd(domain.OpCriterion)},
		{"dod", true, "dod <id>                   Definition of Done: --add item, --opt-out item --reason text, --remove item", recordCmd(domain.OpDoD)},
		{"findings", true, "findings <id>              replace a research issue's findings: --body text | --body-file path|-", recordCmd(domain.OpFindings)},
		{"log", true, "log <id> <text>            append a line to the work log", recordCmd(domain.OpLog)},
		{"knowledge", true, "knowledge new|update|confirm|bootstrap  entries: new <entry> --type --title --description --body..., update <entry> [fields], confirm <entry>... | --drifted, bootstrap", cmdKnowledge},
		{"define", true, "define <id>                write a requirement baseline (open questions must be empty)", opCmd(domain.OpDefine)},
		{"ack", true, "ack <id>                   acknowledge a trivial requirement change: new baseline, enrichment stays valid", opCmd(domain.OpAck)},
		{"ready", true, "ready <id> [--note]        sign off enrichment against the newest baseline", opCmd(domain.OpReady)},
		{"claim", true, "claim <id> [--note]        start work", opCmd(domain.OpClaim)},
		{"release", true, "release <id> [--reason]    stop work without finishing", opCmd(domain.OpRelease)},
		{"complete", true, "complete <id>              finish: [--commit <hash>] --docs <entry>... | --no-impact <reason> [--note]", opCmd(domain.OpComplete)},
		{"drop", true, "drop <id> --reason <text>  drop an issue in any unresolved state", opCmd(domain.OpDrop)},
		{"fmt", true, "fmt [--check]              rewrite files in canonical format (--check: report only, exit 1 on changes)", cmdFmt},
		{"fix", true, "fix                        safe auto-fixes from check", cmdFix},
		{"migrate", true, "migrate                    migrate .prep to this binary's schema", cmdMigrate},
		{"tui", false, "tui [--gallery]            terminal UI next to the harness (--gallery: the design system's components)", cmdTUI},
		{"setup", true, "setup [--harness h,...|--remove h|--refresh]  install harness integrations (skill, hooks, commands)", cmdSetup},
		{"update", true, "update [--check]           replace this binary with the latest release (verifies its checksum)", cmdUpdate},
		{"skill", false, "skill                      print the agent skill (static copy for environments without the binary)", cmdSkill},
		{"version", false, "version                    print the version", cmdVersion},
	}
}

// clock is replaced in tests.
var clock = time.Now

// Main runs the CLI and returns the exit code.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	a := &app{out: stdout, errw: stderr, stdin: stdin, now: clock}
	a.actor = os.Getenv("PREP_ACTOR")
	a.root = os.Getenv("PREP_ROOT")
	rest, err := a.globalFlags(args)
	if err != nil {
		return a.fail(err)
	}
	if len(rest) == 0 || rest[0] == "help" || rest[0] == "-h" || rest[0] == "--help" {
		a.help()
		return exitOK
	}
	name := rest[0]
	for _, c := range commands {
		if c.name == name {
			if a.actor == "" {
				a.actor = "cli/prep-" + Version
			}
			if err := c.run(a, rest[1:]); err != nil {
				return a.fail(err)
			}
			return exitOK
		}
	}
	return a.fail(usageErr("unknown command %q; run prep help", name))
}

// globalFlags strips --json, --root and --by from anywhere in args.
func (a *app) globalFlags(args []string) ([]string, error) {
	var rest []string
	for k := 0; k < len(args); k++ {
		arg := args[k]
		name, val, has := strings.Cut(arg, "=")
		switch name {
		case "--json", "-json":
			a.json = true
		case "--root", "--by":
			if !has {
				if k+1 >= len(args) {
					return nil, usageErr("%s needs a value", name)
				}
				k++
				val = args[k]
			}
			if name == "--root" {
				a.root = val
			} else {
				a.actor = val
			}
		default:
			rest = append(rest, arg)
		}
	}
	return rest, nil
}

func (a *app) help() {
	fmt.Fprintf(a.out, "prep %s — workflow engine and memory for coding agents\n\nUsage: prep <command> [args] [--json] [--root dir] [--by actor]\n\nRead commands:\n", Version)
	for _, c := range commands {
		if !c.write {
			fmt.Fprintf(a.out, "  %s\n", c.usage)
		}
	}
	fmt.Fprintln(a.out, "\nWrite commands:")
	for _, c := range commands {
		if c.write {
			fmt.Fprintf(a.out, "  %s\n", c.usage)
		}
	}
	fmt.Fprintln(a.out, "\nIDs accept any unique suffix. --by (or PREP_ACTOR) names the actor: <producer>/<version> for agents, human:<id> for people.")
}

func usageErr(format string, a ...any) error {
	return &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf(format, a...)}
}

func (a *app) fail(err error) int {
	var se silentErr
	if errors.As(err, &se) {
		return exitFail
	}
	var de *domain.Error
	if !errors.As(err, &de) {
		de = &domain.Error{Code: domain.ErrIO, Message: err.Error()}
	}
	if a.json {
		a.emit(map[string]any{"ok": false, "error": de})
	} else {
		fmt.Fprintf(a.errw, "prep: %s [%s]\n", de.Error(), de.Code)
		for _, d := range de.Diags {
			fmt.Fprintf(a.errw, "  %s\n", formatDiag(d))
		}
	}
	switch de.Code {
	case domain.ErrUsage:
		return exitUsage
	case domain.ErrConflict:
		return exitConflict
	}
	return exitFail
}

func (a *app) emit(v any) {
	enc := json.NewEncoder(a.out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func (a *app) printf(format string, args ...any) { fmt.Fprintf(a.out, format, args...) }

// open locates the project and prepares the stores.
func (a *app) open() error {
	dir := a.root
	if dir == "" {
		dir = "."
	}
	root, err := mdstore.Find(dir)
	if err != nil {
		return err
	}
	a.store = mdstore.Open(root)
	a.kstore = &okf.Store{Root: root}
	return nil
}

func (a *app) load(drift bool) (*domain.Tree, error) {
	if err := a.open(); err != nil {
		return nil, err
	}
	return domain.Load(a.store, a.kstore, drift)
}

// parse runs a flag set allowing flags and positionals in any order.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	fs.SetOutput(io.Discard)
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, usageErr("%s: %v", fs.Name(), err)
		}
		args = fs.Args()
		if len(args) == 0 {
			return pos, nil
		}
		if args[0] == "--" {
			return append(pos, args[1:]...), nil
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
}

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

// one requires exactly one positional argument (the issue reference).
func one(name string, pos []string) (string, error) {
	if len(pos) != 1 {
		return "", usageErr("usage: prep %s <id>", name)
	}
	return pos[0], nil
}

// afterWrite stages or commits the touched paths according to config.
func (a *app) afterWrite(t *domain.Tree, msg string, paths []string) {
	if err := record(a.store.Root, t, msg, paths); err != nil {
		fmt.Fprintf(a.errw, "prep: warning: %v\n", err)
	}
}

// record stages written paths, or commits them in commit mode all. Outside
// a git repository it does nothing.
func record(root string, t *domain.Tree, msg string, paths []string) error {
	if len(paths) == 0 || !gitx.IsRepo(root) {
		return nil
	}
	if t != nil && t.Project.Config.CommitMode == domain.CommitAll {
		return gitx.Commit(root, msg, paths)
	}
	return gitx.Stage(root, paths)
}

func formatDiag(d domain.Diagnostic) string {
	loc := d.File
	if d.Issue != "" {
		loc = d.Issue + "/" + d.File
	}
	return fmt.Sprintf("%-7s %s %s: %s (%s; fix: %s)", d.Severity, d.Code, loc, d.Message, d.Class, d.Fix)
}
