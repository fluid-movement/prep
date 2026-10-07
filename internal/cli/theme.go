package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
)

// cmdTheme lists the TUI themes or starts a custom one in the user's own
// .prep/config.yaml.
func cmdTheme(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("usage: prep theme list | new <name> [--from <theme>]")
	}
	switch args[0] {
	case "list":
		return themeList(a, args[1:])
	case "new":
		return themeNew(a, args[1:])
	}
	return usageErr("prep theme: unknown subcommand %q; use list or new", args[0])
}

func themeList(a *app, args []string) error {
	fs := flag.NewFlagSet("theme list", flag.ContinueOnError)
	if _, err := parse(fs, args); err != nil {
		return err
	}
	var cfg domain.Config
	if a.open() == nil {
		cfg, _ = a.store.LoadConfig()
	}
	active := cfg.Theme
	if active == "" {
		active = palette.Default
	}
	type row struct {
		Name    string `json:"name"`
		Builtin bool   `json:"builtin"`
		Base    string `json:"base,omitempty"`
		Active  bool   `json:"active"`
	}
	var rows []row
	for _, n := range palette.Names(cfg.Themes) {
		rows = append(rows, row{Name: n, Builtin: palette.IsBuiltin(n), Base: cfg.Themes[n].Base, Active: n == active})
	}
	if a.json {
		a.emit(map[string]any{"themes": rows})
		return nil
	}
	for _, r := range rows {
		mark, kind := " ", "built-in"
		if r.Active {
			mark = "*"
		}
		if !r.Builtin {
			kind = "custom"
			if r.Base != "" {
				kind += ", base " + r.Base
			}
		}
		a.printf("%s %-15s %s\n", mark, r.Name, kind)
	}
	return nil
}

func themeNew(a *app, args []string) error {
	fs := flag.NewFlagSet("theme new", flag.ContinueOnError)
	from := fs.String("from", "", "theme whose colors the new one starts with (default: the active theme)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	name, err := one("theme new", pos)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	cfg := t.Project.Config
	if palette.IsBuiltin(name) {
		return usageErr("%q is a built-in theme; pick another name (prep theme list)", name)
	}
	if _, ok := cfg.Themes[name]; ok {
		return usageErr("theme %q exists in config.yaml; edit it there or pick another name", name)
	}
	src := *from
	if src == "" {
		src = cfg.Theme
	}
	p, err := palette.Resolve(src, cfg.Themes)
	if err != nil {
		return usageErr("%v", err)
	}
	themes := map[string]domain.ThemeDef{name: palette.Def(p)}
	for n, d := range cfg.Themes {
		themes[n] = d
	}
	cfg.Themes, cfg.Theme = themes, name
	cfg.ViewOrder = domain.ViewNames(cfg)
	c, err := t.PlanConfig(cfg)
	if err != nil {
		return err
	}
	files, err := a.applyAll(t, []*domain.Change{c})
	if err != nil {
		return err
	}
	if a.json {
		a.reportWrite(writeResult{OK: true, Op: "theme new", Files: files})
		return nil
	}
	a.printf("theme %s: every token from %s, now active\n", name, p.Name)
	for _, f := range files {
		a.printf("  %s\n", f)
	}
	a.printf("Edit the colors under themes.%s; prep tui --gallery shows the result.\n", name)
	return nil
}
