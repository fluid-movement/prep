package cli

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/term"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/projects"
	"github.com/fluid-movement/prep/internal/tui"
	"github.com/fluid-movement/prep/internal/userconfig"
)

func cmdTUI(a *app, args []string) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	gallery := fs.Bool("gallery", false, "show the design system's components")
	agent := fs.Bool("agent", false, "open on the Agent screen: what the agent is doing, live")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		return usageErr("tui needs a terminal; agents use the other commands")
	}
	if *gallery {
		// In a prep project the gallery starts
		// with the chosen one; elsewhere the built-ins.
		var cfg domain.Config
		if a.open() == nil {
			cfg, _ = a.store.LoadConfig()
		}
		return tui.RunGallery(os.Stdin, os.Stdout, cfg)
	}
	if err := a.open(); err != nil {
		return err
	}
	// Switching projects ends the program with where to go; it starts
	// over there, remembering the project it came from for backspace.
	root, _ := filepath.Abs(a.store.Root)
	previous := ""
	var to *tui.Switch
	for {
		var err error
		if to, err = runTUI(root, *agent, previous, to); err != nil || to == nil {
			return err
		}
		_ = projects.Touch(to.Root, a.now())
		previous, root = root, to.Root
		*agent = false
	}
}

// runTUI runs the TUI on the project at root until it quits or switches.
// at is where the switch that led here asked to open, if any.
func runTUI(root string, agent bool, previous string, at *tui.Switch) (*tui.Switch, error) {
	sel, entry := "", ""
	if at != nil {
		sel, entry = at.Issue, at.Entry
	}
	// Loads run in background goroutines, sometimes two at once (reload and
	// check), so each builds its own stores instead of sharing the app's.
	load := func(drift bool) (*domain.Tree, error) {
		t, err := domain.Load(mdstore.Open(root), &okf.Store{Root: root}, drift)
		if err == nil {
			projects.ResolveExternal(t)
		}
		return t, err
	}
	var mu sync.Mutex
	write := func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		st := mdstore.Open(root)
		t, err := domain.Load(st, &okf.Store{Root: root}, false)
		if err != nil {
			return "", err
		}
		c, err := plan(t)
		if err != nil {
			return "", err
		}
		if err := t.CheckWrite(c); err != nil {
			return "", err
		}
		files, err := st.Apply(c)
		if err != nil {
			return "", err
		}
		if err := record(root, files); err != nil {
			return c.IssueID, fmt.Errorf("written, but git failed: %v", err)
		}
		return c.IssueID, nil
	}
	// A broken user configuration leaves the mouse on; prep setup reports it.
	ucfg, _ := userconfig.Load()
	return tui.Run(tui.Options{
		Project:    projects.Describe(root).Name,
		Root:       root,
		Projects:   projects.List,
		Previous:   previous,
		Select:     sel,
		Entry:      entry,
		NoMouse:    ucfg.TUI.Mouse != nil && !*ucfg.TUI.Mouse,
		SaveMouse:  saveMouse,
		Single:     ucfg.TUI.Layout == "single",
		SaveLayout: saveLayout,
		Actor:      humanActor(),
		Write:      write,
		Load:       func() (*domain.Tree, error) { return load(false) },
		Watch:      filepath.Join(root, mdstore.Dir),
		// The last 2,000 events at most: the stream prunes itself near that.
		Activity:   func() ([]activity.Event, error) { return activity.Read(filepath.Join(root, mdstore.Dir), 2000) },
		StartAgent: agent,
		Check: func() ([]domain.Diagnostic, error) {
			t, err := load(true)
			if err != nil {
				return nil, err
			}
			return validate(t), nil
		},
	}, os.Stdin, os.Stdout)
}

// saveLayout records the TUI's layout choice in the user configuration,
// keeping its other choices.
func saveLayout(single bool) error {
	c, err := userconfig.Load()
	if err != nil {
		return err
	}
	c.TUI.Layout = ""
	if single {
		c.TUI.Layout = "single"
	}
	_, err = userconfig.Save(c)
	return err
}

// saveMouse records the TUI's mouse capture choice in the user
// configuration, keeping its other choices.
func saveMouse(on bool) error {
	c, err := userconfig.Load()
	if err != nil {
		return err
	}
	c.TUI.Mouse = &on
	_, err = userconfig.Save(c)
	return err
}

// humanActor names the person using the TUI. The domain treats human:
// actors as people, so they cannot complete code issues.
func humanActor() string {
	name := os.Getenv("USER")
	if u, err := user.Current(); err == nil && u.Username != "" {
		name = u.Username
	}
	name = strings.Join(strings.Fields(name), "-")
	if name == "" {
		name = "tui"
	}
	return "human:" + name
}
