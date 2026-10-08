package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/setup"
	_ "github.com/fluid-movement/prep/internal/setup/claudecode" // registers the Claude Code harness
	"github.com/fluid-movement/prep/internal/tui"
	"github.com/fluid-movement/prep/internal/userconfig"
)

// openTerminal opens the controlling terminal for prompts, which works when
// stdin is a pipe (curl | sh). Tests replace it.
var openTerminal = func() (io.ReadWriteCloser, error) {
	name := "/dev/tty"
	if runtime.GOOS == "windows" {
		name = "CONIN$"
	}
	return os.OpenFile(name, os.O_RDWR, 0)
}

// chooseHarnesses asks the user which harnesses to integrate. Tests replace
// it.
var chooseHarnesses = func(statuses []setup.Status) ([]string, bool, error) {
	tty, err := openTerminal()
	if err != nil {
		return nil, false, err
	}
	defer tty.Close()
	var items []tui.ChecklistItem
	for _, s := range statuses {
		items = append(items, tui.ChecklistItem{Label: s.Title, Detail: describe(s), Checked: s.Chosen || s.Detected || s.Installed != ""})
	}
	out := io.Writer(os.Stdout)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		out = tty
	}
	chosen, ok, err := tui.RunChecklist(tty, out, "Integrate prep with these harnesses", items)
	if err != nil || !ok {
		return nil, ok, err
	}
	var names []string
	for k, it := range chosen {
		if it.Checked {
			names = append(names, statuses[k].Name)
		}
	}
	return names, true, nil
}

func describe(s setup.Status) string {
	var parts []string
	switch {
	case s.Installed != "":
		parts = append(parts, "installed "+s.Installed)
	case s.Detected:
		parts = append(parts, "detected")
	default:
		parts = append(parts, "not found on this machine")
	}
	if s.Err != "" {
		parts = append(parts, s.Err)
	}
	return strings.Join(parts, " · ")
}

func cmdSetup(a *app, args []string) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	var harness, remove multi
	fs.Var(&harness, "harness", "integrate exactly these harnesses (repeatable or comma-separated)")
	fs.Var(&remove, "remove", "remove the integration of a harness (repeatable)")
	refresh := fs.Bool("refresh", false, "bring the chosen harnesses to this binary's version")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	modes := 0
	for _, on := range []bool{len(harness) > 0, len(remove) > 0, *refresh} {
		if on {
			modes++
		}
	}
	if modes > 1 {
		return usageErr("pass one of --harness, --remove or --refresh")
	}
	cfg, err := userconfig.Load()
	if err != nil {
		return err
	}

	var results []setup.Result
	switch {
	case *refresh:
		results = setup.Refresh(cfg.Harnesses, Version)
	case len(remove) > 0:
		names := splitTags(remove)
		if err := setup.Validate(names); err != nil {
			return usageErr("%v", err)
		}
		var keep []string
		for _, n := range cfg.Harnesses {
			if !contains(names, n) {
				keep = append(keep, n)
			}
		}
		for _, n := range names {
			h, _ := setup.Get(n)
			r := setup.Result{Name: n, Action: "removed"}
			if err := h.Remove(); err != nil {
				r.Action, r.Err = "failed", err.Error()
			}
			results = append(results, r)
		}
		cfg.Harnesses = keep
	default:
		if len(setup.All()) == 0 {
			if a.json {
				a.emit(map[string]any{"ok": true, "harnesses": []string{}, "results": []setup.Result{}, "note": "no harness integrations in this version"})
			} else {
				a.printf("setup: this version of prep has no harness integrations yet; nothing to do\n")
			}
			return nil
		}
		names := splitTags(harness)
		if len(harness) == 0 {
			if a.json {
				return usageErr("interactive setup cannot print JSON; pass --harness")
			}
			var ok bool
			names, ok, err = chooseHarnesses(setup.Survey(cfg.Harnesses))
			if err != nil {
				return &domain.Error{Code: domain.ErrUsage, Message: fmt.Sprintf("no terminal to ask on (%v); run prep setup --harness %s", err, strings.Join(setup.Names(), ","))}
			}
			if !ok {
				a.printf("setup cancelled; nothing changed\n")
				return nil
			}
		}
		if err := setup.Validate(names); err != nil {
			return usageErr("%v", err)
		}
		results = setup.Apply(names, Version)
		cfg.Harnesses = names
	}

	path := ""
	if !*refresh {
		if path, err = userconfig.Save(cfg); err != nil {
			return err
		}
	}
	failed := 0
	for _, r := range results {
		if r.Action == "failed" {
			failed++
		}
	}
	if a.json {
		if results == nil {
			results = []setup.Result{}
		}
		a.emit(map[string]any{"ok": failed == 0, "harnesses": cfg.Harnesses, "results": results, "config": path})
	} else {
		if len(results) == 0 {
			a.printf("setup: nothing to do\n")
		}
		for _, r := range results {
			line := fmt.Sprintf("%-12s %s", r.Name, r.Action)
			if r.Version != "" && r.Action != "removed" {
				line += " " + r.Version
			}
			if r.Err != "" {
				line += ": " + r.Err
			}
			a.printf("%s\n", line)
			for _, w := range r.Warnings {
				a.printf("  warning: %s\n", w)
			}
		}
		if path != "" {
			a.printf("choices saved in %s\n", path)
		}
	}
	if failed > 0 {
		return silentErr{&domain.Error{Code: domain.ErrInvalid, Message: fmt.Sprintf("%d harness integration(s) failed", failed)}}
	}
	return nil
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
