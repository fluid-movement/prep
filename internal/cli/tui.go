package cli

import (
	"flag"
	"os"

	"golang.org/x/term"

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
	return usageErr("the issue views are not built yet (issue 20261005-152617); try prep tui --gallery")
}
