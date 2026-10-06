package cli

import (
	"flag"
	"os/exec"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/update"
)

// newUpdater is replaced in tests.
var newUpdater = update.New

func cmdUpdate(a *app, args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	check := fs.Bool("check", false, "only report whether a newer release exists")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	u, err := newUpdater(Version)
	if err != nil {
		return err
	}
	if !update.IsRelease(Version) {
		return &domain.Error{Code: domain.ErrInvalid, Message: "this is a development build (" + Version + "); update a checkout with just install, or install a release with " + update.GoInstall}
	}
	if cmd := update.Managed(u.Exe); cmd != "" && !*check {
		return &domain.Error{Code: domain.ErrInvalid, Message: "prep was installed by another tool; update it with: " + cmd}
	}
	r, err := u.Latest()
	if err != nil {
		return err
	}
	newer, err := update.Newer(Version, r.Tag)
	if err != nil {
		return err
	}
	result := map[string]any{"ok": true, "current": Version, "latest": r.Tag, "newer": newer, "updated": false}
	switch {
	case !newer:
		if !a.json {
			a.printf("prep %s is the latest release\n", Version)
		}
	case *check:
		if !a.json {
			a.printf("prep %s is available (installed: %s); run prep update\n", r.Tag, Version)
		}
	default:
		if err := u.Install(r); err != nil {
			return err
		}
		result["updated"] = true
		if !a.json {
			a.printf("updated prep %s → %s (%s)\n", Version, r.Tag, u.Exe)
		}
		// The new binary knows the current integrations, so it refreshes them.
		refresh := exec.Command(u.Exe, "setup", "--refresh")
		if !a.json {
			refresh.Stdout, refresh.Stderr = a.out, a.errw
		}
		if err := refresh.Run(); err != nil {
			result["refresh_error"] = err.Error()
			if !a.json {
				a.printf("prep: refreshing harness integrations failed (%v); run prep setup --refresh\n", err)
			}
		}
	}
	if a.json {
		a.emit(result)
	}
	return nil
}
