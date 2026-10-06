package cli

import (
	"flag"
	"os"
	"path/filepath"

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
	return tui.Run(tui.Options{
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
