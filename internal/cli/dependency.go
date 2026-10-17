package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
)

func cmdDependency(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" {
		return usageErr("dependency needs show <id>")
	}
	fs := flag.NewFlagSet("dependency show", flag.ContinueOnError)
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("dependency show has invalid arguments")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	i := t.Issues[id]
	type item struct {
		Ref         string        `json:"ref"`
		Observation domain.XIssue `json:"observation"`
	}
	items := []item{}
	for _, ref := range i.DependsOn {
		if _, _, cross := domain.SplitRef(ref); cross {
			items = append(items, item{ref, t.ExternalDependency(id, ref)})
		}
	}
	if a.json {
		a.emit(map[string]any{"id": id, "dependencies": items})
		return nil
	}
	if len(items) == 0 {
		a.printf("No external dependencies.\n")
	}
	for _, it := range items {
		x := it.Observation
		a.printf("%s: %s", it.Ref, x.Outcome)
		if x.Uncertainty != "" {
			a.printf("; unknown: %s", x.Uncertainty)
		}
		a.printf("; source %s (%s), checkout %s, revision %s\n", x.Project, x.SourceProjectID, x.Root, x.Revision)
		if x.ProvenanceWarning != "" {
			a.printf("  provenance warning: %s\n", x.ProvenanceWarning)
		}
	}
	return nil
}
