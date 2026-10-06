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

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/tui"
)

func cmdTUI(a *app, args []string) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	gallery := fs.Bool("gallery", false, "show the design system's components")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		return usageErr("tui needs a terminal; agents use the other commands")
	}
	if *gallery {
		return tui.RunGallery(os.Stdin, os.Stdout)
	}
	if err := a.open(); err != nil {
		return err
	}
	// Loads run in background goroutines, sometimes two at once (reload and
	// check), so each builds its own stores instead of sharing the app's.
	root := a.store.Root
	load := func(drift bool) (*domain.Tree, error) {
		return domain.Load(mdstore.Open(root), &okf.Store{Root: root}, drift)
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
		if err := record(root, t, strings.Join(strings.Fields(fmt.Sprintf("prep: %s %s (tui)", c.Op, c.IssueID)), " "), files); err != nil {
			return c.IssueID, fmt.Errorf("written, but git failed: %v", err)
		}
		return c.IssueID, nil
	}
	return tui.Run(tui.Options{
		Actor: humanActor(),
		Write: write,
		Load:  func() (*domain.Tree, error) { return load(false) },
		Watch: filepath.Join(root, mdstore.Dir),
		Check: func() ([]domain.Diagnostic, error) {
			t, err := load(true)
			if err != nil {
				return nil, err
			}
			return domain.Validate(t), nil
		},
	}, os.Stdin, os.Stdout)
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
