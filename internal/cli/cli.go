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
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/projects"
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
	// project is --project (or PREP_PROJECT): the registered project the
	// command runs against instead of the one at root. target is what it
	// resolved to; home is the project the command was started in, when
	// that is another one, whose activity stream also notes the command.
	project     string
	rootFlag    bool
	projectFlag bool
	target      *projects.Project
	home        string
	actor       string
	now         func() time.Time
	entropy     io.Reader // random bits of new issue IDs

	store  *mdstore.Store
	kstore *okf.Store

	// event is what the command records in the activity stream when it
	// succeeds; eventReads adds the size of its output.
	event      *activity.Event
	eventReads bool
	named      bool // the actor came from --by or PREP_ACTOR, not the default
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
		{"hook", false, "hook session-start|prompt|finish|before-tool|after-tool [--harness claude-code|codex] [--plugin path]  shared lifecycle hooks (event JSON on stdin)", cmdHook},
		{"guide", false, "guide <id>                 step contract: state, transitions, unmet gates, inputs, outputs", cmdGuide},
		{"list", false, "list [query flags]         query issues: --state --kind --tag --priority --under --stale --actionable --blocked --parent --leaf --top --text --tree --view --max N; not-<value> negates (--state not-open)", cmdList},
		{"next", false, "next [--under <id>]        actionable issues: ready, not stale, dependencies done, unclaimed", cmdNext},
		{"show", false, "show <id>                  read one issue with its derived state", cmdShow},
		{"check", false, "check [--no-drift]         validate the whole tree", cmdCheck},
		{"activity", false, "activity [--max N] [--actor a]  what agents did with prep, newest last (local, per machine); activity add: harness events as JSON lines on stdin", cmdActivity},
		{"flags", false, "flags                      every query flag with the values it accepts now and their issue counts", cmdFlags},
		{"knowledge", false, "knowledge list|find|show|review  read entries; review show <id> reports review applicability", cmdKnowledge},
		{"evidence", false, "evidence show <id>         inspect reported verification and current applicability", cmdEvidence},
		{"evidence", true, "evidence record <id>       record scoped manual or automated results (never executes checks)", cmdEvidence},
		{"exception", false, "exception show <id>        inspect verification waivers and applicability", cmdException},
		{"exception", true, "exception approve <id> --requirement R --reason T --scope P --by human:N  approve a scoped waiver", cmdException},
		{"dependency", false, "dependency show <id>         show external dependency observations and override applicability", cmdDependency},
		{"dependency", true, "dependency override <id> <project:id> --reason T --by human:N  override one unknown external dependency observation", cmdDependency},
		{"policy", false, "policy show [id]          show project policy or an issue's bound/current contracts", cmdPolicy},
		{"policy", true, "policy set --verification V --documentation D | adopt <id> | retain <id> --reason T --by human:N  configure or reconcile completion methods", cmdPolicy},
		{"views", false, "views                      saved views from config.yaml (else config.yaml.dist)", cmdViews},
		{"projects", false, "projects [add <dir> | remove <dir|name|id>]  prep projects on this machine, most recently used first (recorded by every command run in one): name, ID suffix, last use, root, remote", cmdProjects},
		{"links", false, "links                      this project's links to other projects and where each is checked out here", cmdLinks},
		{"new", true, "new --title T --kind K     create an issue [--parent id] [--depends-on id|project:id]... [--tag t]... [--priority p] [--body text | --body-file path|-]", cmdNew},
		{"edit", true, "edit <id>                  change [--title T] [--kind K] [--parent id|''] [--depends-on id|project:id|'']... [--tag t|'']... [--priority p] [--body text | --body-file path|-]", cmdEdit},
		{"context", true, "context <id>               replace the implementation context: --body text | --body-file path|-", recordCmd(domain.OpContext)},
		{"decide", true, "decide <id> --title T      append a decision [--body text | --body-file path|-] [--supersedes D<n>] [--outcome]", recordCmd(domain.OpDecide)},
		{"criterion", true, "criterion <id>             acceptance criteria: --add text, --check n, --uncheck n, --remove n (repeatable)", recordCmd(domain.OpCriterion)},
		{"dod", true, "dod <id> --add|--opt-out|--remove | approve <id> --item T --reason R --by human:N  Definition of Done", recordCmd(domain.OpDoD)},
		{"findings", true, "findings <id>              replace a research issue's findings: --body text | --body-file path|-", recordCmd(domain.OpFindings)},
		{"log", true, "log <id> <text>            append a line to the work log", recordCmd(domain.OpLog)},
		{"focus", true, "focus <id> | --clear        set or clear the issue you work on, for prep tui's Agent view (local; writes no records)", cmdFocus},
		{"import", true, "import                     create the issue that guides importing existing work items (TODOs, GitHub issues, other trackers)", cmdImport},
		{"knowledge", true, "knowledge review|new|update|confirm|bootstrap  review record <id> ...; update knowledge entries with new/update; confirm <entry>... | --drifted", cmdKnowledge},
		{"link", true, "link add <project> [--as name] | rename <old> <new> | remove <name>  links to other projects (rename rewrites every reference)", cmdLink},
		{"define", true, "define <id>                write a requirement baseline (open questions must be empty)", opCmd(domain.OpDefine)},
		{"ack", true, "ack <id>                   acknowledge a trivial requirement change: new baseline, enrichment stays valid", opCmd(domain.OpAck)},
		{"ready", true, "ready <id> [--note]        sign off enrichment against the newest baseline", opCmd(domain.OpReady)},
		{"claim", true, "claim <id> [--note]        start work", opCmd(domain.OpClaim)},
		{"release", true, "release <id> [--reason]    stop work without finishing", opCmd(domain.OpRelease)},
		{"complete", true, "complete <id>              finish: [--commit <hash>] --docs <entry>... | --no-impact <reason> [--note]", opCmd(domain.OpComplete)},
		{"review", true, "review show <id> | approve <id> --reason <text> --by human:<name> [--scope <path>...]", cmdReview},
		{"artifact", true, "artifact add|remove <id> --source <identity> --location <reference> [--title <text>]  record an external reference only", cmdArtifact},
		{"reopen", true, "reopen <id> --reason <text> explicitly resume a done issue through its existing ready sign-off", opCmd(domain.OpReopen)},
		{"drop", true, "drop <id> --reason <text>  drop an issue in any unresolved state", opCmd(domain.OpDrop)},
		{"fmt", true, "fmt [--check]              rewrite files in canonical format (--check: report only, exit 1 on changes)", cmdFmt},
		{"fix", true, "fix                        safe auto-fixes from check", cmdFix},
		{"migrate", true, "migrate                    migrate .prep to this binary's schema", cmdMigrate},
		{"tui", false, "tui [--agent|--gallery]    terminal UI next to the harness (--agent: open on what the agent is doing; --gallery: every component, t/T switch themes)", cmdTUI},
		{"setup", true, "setup [--harness h,...|--remove h|--refresh]  install harness integrations (skill, hooks, commands)", cmdSetup},
		{"update", true, "update [--check]           replace this binary with the latest release (verifies its checksum)", cmdUpdate},
		{"eval", false, evalUsage, cmdEval},
		{"skill", false, "skill                      print the agent skill (static copy for environments without the binary)", cmdSkill},
		{"version", false, "version                    print the version", cmdVersion},
	}
}

// clock and entropy are replaced in tests; nil entropy means crypto/rand.
var (
	clock   = time.Now
	entropy io.Reader
)

// Main runs the CLI and returns the exit code.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	out := &counter{w: stdout}
	a := &app{out: out, errw: stderr, stdin: stdin, now: clock, entropy: entropy}
	a.testSeams()
	a.actor = os.Getenv("PREP_ACTOR")
	a.root = os.Getenv("PREP_ROOT")
	a.project = os.Getenv("PREP_PROJECT")
	rest, err := a.globalFlags(args)
	if err != nil {
		return a.fail(err)
	}
	// Hooks must be able to return before any project lookup (prompt off),
	// and use the event's cwd rather than a registered PREP_PROJECT.
	if len(rest) > 0 && rest[0] == "hook" {
		if err := cmdHook(a, rest[1:]); err != nil {
			return a.fail(err)
		}
		return exitOK
	}
	if err := a.resolveProject(); err != nil {
		return a.fail(err)
	}
	if len(rest) == 0 || rest[0] == "help" || rest[0] == "-h" || rest[0] == "--help" {
		a.help()
		return exitOK
	}
	name := rest[0]
	for _, c := range commands {
		if c.name == name {
			a.named = a.actor != ""
			if !a.named {
				a.actor = "cli/prep-" + Version
			}
			if err := c.run(a, rest[1:]); err != nil {
				return a.fail(err)
			}
			a.recordEvent(out.n)
			a.recordProject()
			return exitOK
		}
	}
	return a.fail(usageErr("unknown command %q; run prep help", name))
}

// testSeams reads the test-only environment variables the eval fixture
// builder sets: PREP_TEST_NOW (RFC3339) replaces the clock and PREP_TEST_SEED
// (an integer) seeds the random bits of new issue IDs. Unset, nothing changes.
func (a *app) testSeams() {
	if v := os.Getenv("PREP_TEST_NOW"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			a.now = func() time.Time { return t }
		}
	}
	if v := os.Getenv("PREP_TEST_SEED"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			a.entropy = rand.New(rand.NewSource(n))
		}
	}
}

// resolveProject points the command at the project --project names. Flags
// beat environment variables: --root overrides PREP_PROJECT and --project
// overrides PREP_ROOT; both flags together are a usage error.
func (a *app) resolveProject() error {
	switch {
	case a.project == "":
		return nil
	case a.rootFlag && a.projectFlag:
		return usageErr("--project and --root both choose the project; give one")
	case a.rootFlag:
		return nil
	}
	start := a.root
	if start == "" {
		start = "."
	}
	home := ""
	if h, err := mdstore.Find(start); err == nil {
		home, _ = filepath.Abs(h)
	}
	p, err := a.projectNamed(home, a.project)
	if err != nil {
		return err
	}
	if home != p.Root {
		a.home = home
	}
	a.root, a.target = p.Root, &p
	return nil
}

// projectNamed resolves --project: a link name of the home project first
// (so a reference's name works as written), else the registry's rules.
func (a *app) projectNamed(home, query string) (projects.Project, error) {
	if home != "" {
		if hp, err := mdstore.ReadProject(home); err == nil {
			if l, ok := hp.Link(query); ok {
				p, found, err := projects.ByID(l.ID)
				if err != nil {
					return p, err
				}
				if !found {
					hint := ""
					if l.Remote != "" {
						hint = "; clone " + l.Remote + " and run prep there once (or prep projects add <dir>)"
					}
					return p, &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("linked project %s is not checked out here%s", query, hint)}
				}
				return p, nil
			}
		}
	}
	return projects.Resolve(query)
}

// globalFlags strips --json, --root, --project and --by from anywhere in args before
// a "--"; from "--" on, args pass through as written (so "--" lets a text
// argument be "--json"), and the command's own flag parsing stops there.
func (a *app) globalFlags(args []string) ([]string, error) {
	var rest []string
	for k := 0; k < len(args); k++ {
		arg := args[k]
		if arg == "--" {
			return append(rest, args[k:]...), nil
		}
		name, val, has := strings.Cut(arg, "=")
		switch name {
		case "--json", "-json":
			a.json = true
			if has {
				on, err := strconv.ParseBool(val)
				if err != nil {
					return nil, usageErr("%s takes true or false, not %q", name, val)
				}
				a.json = on
			}
		case "--root", "--by", "--project":
			if !has {
				if k+1 >= len(args) {
					return nil, usageErr("%s needs a value", name)
				}
				k++
				val = args[k]
			}
			switch name {
			case "--root":
				a.root, a.rootFlag = val, true
			case "--project":
				a.project, a.projectFlag = val, true
			default:
				a.actor = val
			}
		default:
			rest = append(rest, arg)
		}
	}
	return rest, nil
}

func (a *app) help() {
	fmt.Fprintf(a.out, "prep %s — workflow engine and memory for coding agents\n\nUsage: prep <command> [args] [--json] [--root dir | --project name] [--by actor]\n\nRead commands:\n", Version)
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
	fmt.Fprintln(a.out, "\n--project (or PREP_PROJECT) runs a command in another prep project on this machine: a name, ID, remote or path from prep projects.")
	fmt.Fprintln(a.out, "\nIDs accept any unique suffix. --by (or PREP_ACTOR) names the actor: <producer>/<version> for agents, human:<id> for people.")
}

func usageErr(format string, a ...any) error {
	return &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf(format, a...)}
}

func withNext(err error, next string) error {
	var de *domain.Error
	if !errors.As(err, &de) || de.Next != "" {
		return err
	}
	copy := *de
	copy.Next = next
	return &copy
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
		if de.Next != "" {
			fmt.Fprintf(a.errw, "Next: %s\n", de.Next)
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
	t, err := domain.Load(a.store, a.kstore, drift)
	if err == nil {
		projects.ResolveExternal(t)
	}
	return t, err
}

// parse runs a flag set allowing flags and positionals in any order.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	fs.SetOutput(io.Discard)
	var pos []string
	for {
		// flag.Parse drops a leading "--" without saying so; everything
		// after it is positional.
		if len(args) > 0 && args[0] == "--" {
			return append(pos, args[1:]...), nil
		}
		if err := fs.Parse(args); err != nil {
			return nil, usageErr("%s: %v", fs.Name(), err)
		}
		if n := len(fs.Args()); n < len(args) && args[len(args)-n-1] == "--" {
			return append(pos, fs.Args()...), nil
		}
		args = fs.Args()
		if len(args) == 0 {
			return pos, nil
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

// afterWrite stages the touched paths.
func (a *app) afterWrite(paths []string) {
	if err := record(a.store.Root, paths); err != nil {
		fmt.Fprintf(a.errw, "prep: warning: %v\n", err)
	}
}

// record stages written paths; the user or agent commits them with the
// code. Ignored paths (the user's own config) stay unstaged. Outside a git
// repository it does nothing.
func record(root string, paths []string) error {
	if len(paths) == 0 || !gitx.IsRepo(root) {
		return nil
	}
	return gitx.Stage(root, gitx.Tracked(root, paths))
}

func formatDiag(d domain.Diagnostic) string {
	loc := d.File
	if d.Issue != "" {
		loc = d.Issue + "/" + d.File
	}
	return fmt.Sprintf("%-7s %s %s: %s (%s; fix: %s)", d.Severity, d.Code, loc, d.Message, d.Class, d.Fix)
}
