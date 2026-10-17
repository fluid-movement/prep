package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
)

func cmdArtifact(a *app, args []string) error {
	if len(args) == 0 || args[0] != "add" && args[0] != "remove" {
		return usageErr("artifact needs add|remove <id> --source <identity> --location <URI-or-path> [--title <text>]; references never count as verification")
	}
	fs := flag.NewFlagSet("artifact "+args[0], flag.ContinueOnError)
	var ref domain.ArtifactReference
	fs.StringVar(&ref.Source, "source", "", "external system or document identity")
	fs.StringVar(&ref.Location, "location", "", "reference-only URI or path; never fetched by prep")
	if args[0] == "add" {
		fs.StringVar(&ref.Title, "title", "", "optional reference label")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("artifact needs one issue ID, --source and --location")
	}
	unlock, err := a.lockWrite()
	if err != nil {
		return err
	}
	defer unlock()
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	c, err := t.PlanArtifact(id, domain.ArtifactInput{Reference: ref, Remove: args[0] == "remove", Actor: a.actor, Now: a.now()})
	if err != nil {
		return err
	}
	if err := t.CheckWrite(c); err != nil {
		return err
	}
	files, err := a.store.Apply(c)
	if err != nil {
		return err
	}
	a.afterWrite(files)
	a.reportWrite(writeResult{OK: true, Op: string(c.Op), ID: id, State: t.State(id), Files: files, Next: "prep show " + id})
	a.noteRead(issueEvent(t, string(c.Op), id, "record an external artifact reference (not verification)"))
	return nil
}
