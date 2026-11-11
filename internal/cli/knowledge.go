package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
		var projectWide bool
		fs.Var(&scope, "scope", "code path glob (repeatable)")
		if typ == "decision" {
			fs.BoolVar(&projectWide, "project-wide", false, "decision applies to the whole project (scope **)")
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
		if typ == "decision" {
			if projectWide && set["scope"] {
				return usageErr("pass --project-wide or --scope, not both")
			}
			for _, s := range scope {
				if strings.TrimSpace(s) == "**" {
					return usageErr("use --project-wide instead of --scope '**'")
				}
			}
			if projectWide {
				scope = multi{"**"}
			}
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
		if typ == "decision" && in.Issue == "" && strings.TrimSpace(strings.Join(scope, "")) == "" {
			choice := "an issue (prep decide <id> --title T --description D)"
			if id := a.decisionIssue(t); id != "" {
				choice = fmt.Sprintf("issue %s (prep decide %s --title T --description D)", id, id)
			}
			return usageErr("decide needs %s, --scope <glob>, or --project-wide", choice)
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

// decisionIssue proposes this session's focus, else the single issue in progress.
func (a *app) decisionIssue(t *domain.Tree) string {
	if events, err := activity.Read(a.prepDir(), 0); err == nil {
		key := (activity.Event{Actor: a.actor, Session: os.Getenv("PREP_SESSION")}).Agent()
		if current := activity.AgentKeys(events)[key]; current != "" {
			key = current
		}
		if f, ok := activity.Foci(events)[key]; ok && t.Issues[f.Issue] != nil && !t.State(f.Issue).Terminal() {
			return f.Issue
		}
	}
	id := ""
	for _, candidate := range t.IDs() {
		if t.State(candidate) == domain.StateInProgress {
			if id != "" {
				return ""
			}
			id = candidate
		}
	}
	return id
}
