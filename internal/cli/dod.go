package cli

import (
	"flag"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
)

// cmdDoD edits project guidance without changing issue completion gates.
func cmdDoD(a *app, args []string) error {
	fs := flag.NewFlagSet("dod", flag.ContinueOnError)
	var adds, removes multi
	fs.Var(&adds, "add", "add a project Definition of Done item (repeatable)")
	fs.Var(&removes, "remove", "remove an item by its exact text (repeatable)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 0 || len(adds)+len(removes) == 0 {
		return usageErr("dod needs --add <item> or --remove <item>, without an issue ID")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	items := append([]string(nil), t.Project.DoD...)
	for _, item := range removes {
		found := false
		var keep []string
		for _, v := range items {
			if v == item {
				found = true
			} else {
				keep = append(keep, v)
			}
		}
		if !found {
			return usageErr("no project Definition of Done item %q", item)
		}
		items = keep
	}
	for _, item := range adds {
		item = strings.TrimSpace(item)
		if item == "" || strings.ContainsAny(item, "\r\n") {
			return usageErr("Definition of Done items must be nonempty single lines")
		}
		found := false
		for _, v := range items {
			found = found || v == item
		}
		if !found {
			items = append(items, item)
		}
	}
	c := &domain.Change{Op: "dod", DoD: &items}
	if err := t.CheckWrite(c); err != nil {
		return err
	}
	files, err := a.store.Apply(c)
	if err != nil {
		return err
	}
	a.afterWrite(files)
	a.note(activity.Event{Op: "dod", Verb: "changed the project Definition of Done", Area: activity.AreaPrep})
	a.reportWrite(writeResult{OK: true, Op: "dod", Files: files, Next: "prep prime"})
	return nil
}
