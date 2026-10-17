package cli

import (
	"flag"
	"fmt"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
)

func cmdKnowledge(a *app, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "list":
			return knowledgeList(a, args[1:])
		case "find":
			return knowledgeFind(a, args[1:])
		case "show":
			return knowledgeShow(a, args[1:])
		}
	}
	return usageErr("knowledge needs a subcommand: list, find or show")
}

func recordTitle(title string, superseded bool) string {
	if superseded {
		return title + " [superseded]"
	}
	return title
}

func knowledgeRecordCmd(typ string) func(*app, []string) error {
	return func(a *app, args []string) error {
		op := typ
		if typ == "decision" {
			op = "decide"
		}
		fs := flag.NewFlagSet(op, flag.ContinueOnError)
		in := domain.KnowledgeRecord{Type: typ}
		fs.StringVar(&in.Title, "title", "", "record title")
		fs.StringVar(&in.Description, "description", "", "required one-line description")
		body := fs.String("body", "", "record body")
		bodyFile := fs.String("body-file", "", "read body from a file (- for stdin)")
		var scope, supersedes multi
		fs.Var(&scope, "scope", "code path glob (repeatable; absent means project-wide)")
		if typ == "decision" {
			fs.Var(&supersedes, "supersedes", "older knowledge record ID or unique suffix (repeatable)")
			fs.BoolVar(&in.Outcome, "outcome", false, "answers the linked decision issue")
		}
		pos, err := parse(fs, args)
		if err != nil {
			return err
		}
		if len(pos) > 1 {
			return usageErr("expected at most one issue ID")
		}
		set := map[string]bool{}
		fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
		if set["body"] && set["body-file"] {
			return usageErr("pass --body or --body-file, not both")
		}
		in.Body, err = a.readBody(*body, *bodyFile)
		if err != nil {
			return err
		}
		unlock, err := a.lockWrite()
		if err != nil {
			return err
		}
		defer unlock()
		t, err := a.load()
		if err != nil {
			return err
		}
		if len(pos) == 1 {
			in.Issue, err = t.Resolve(pos[0])
			if err != nil {
				return err
			}
		}
		in.At = a.now()
		in.ID, err = domain.NewULID(in.At, a.entropy)
		if err != nil {
			return err
		}
		in.Scope, in.Supersedes = scope, supersedes
		c, err := t.PlanKnowledgeRecord(in)
		if err != nil {
			return err
		}
		e, err := a.kstore.Render(c.Knowledge)
		if err != nil {
			return err
		}
		c.KnowledgeEntry = e
		if err = t.CheckWrite(c); err != nil {
			return err
		}
		files, err := a.kstore.Apply(c.Knowledge)
		if err != nil {
			return err
		}
		a.afterWrite(files)
		a.note(activity.Event{Op: op, Issue: in.Issue, Verb: fmt.Sprintf("recorded %s: %s", typ, e.Title), Target: e.Path, Area: activity.AreaKnowledge})
		a.reportWrite(writeResult{OK: true, Op: op, ID: in.Issue, Entry: e.Path, Files: files})
		return nil
	}
}
