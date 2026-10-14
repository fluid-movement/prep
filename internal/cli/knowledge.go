package cli

import (
	"flag"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/worktree"
)

// cmdKnowledge dispatches the knowledge reads (list, find, show) and writes
// (new, update, confirm, bootstrap).
func cmdKnowledge(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("knowledge needs a subcommand: list, find, show, review, new, update, confirm or bootstrap")
	}
	switch args[0] {
	case "list":
		return knowledgeList(a, args[1:])
	case "find":
		return knowledgeFind(a, args[1:])
	case "show":
		return knowledgeShow(a, args[1:])
	case "review":
		return knowledgeReview(a, args[1:])
	case "new":
		return knowledgeEdit(a, args[1:], true)
	case "update":
		return knowledgeEdit(a, args[1:], false)
	case "confirm":
		return knowledgeConfirm(a, args[1:])
	case "bootstrap":
		return knowledgeBootstrap(a, args[1:])
	}
	return usageErr("unknown knowledge subcommand %q: use list, find, show, review, new, update, confirm or bootstrap", args[0])
}

func knowledgeReview(a *app, args []string) error {
	if len(args) == 0 || args[0] != "record" && args[0] != "show" {
		return usageErr("knowledge review needs show <id> or record <id> --entry <path> --kind updated|unchanged --observation <text> --scope <work-path> [--config-scope <path>] (or --no-impact --reason <text>)")
	}
	fs := flag.NewFlagSet("knowledge review "+args[0], flag.ContinueOnError)
	var r domain.KnowledgeReview
	var scopes, configs multi
	var noImpact bool
	var reason string
	if args[0] == "record" {
		fs.StringVar(&r.Entry, "entry", "", "knowledge entry assessed")
		fs.StringVar(&r.Kind, "kind", "", "updated or unchanged")
		fs.StringVar(&r.Observation, "observation", "", "what was reviewed and why")
		fs.BoolVar(&noImpact, "no-impact", false, "report that knowledge is unaffected")
		fs.StringVar(&reason, "reason", "", "why no knowledge entry applies")
		fs.Var(&scopes, "scope", "project-relative work path reviewed (repeatable)")
		fs.Var(&configs, "config-scope", "project-relative check configuration path (repeatable)")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("knowledge review needs one issue ID")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	issue := t.Issues[id]
	if args[0] == "show" {
		type result struct {
			Review     domain.KnowledgeReview `json:"review"`
			Applicable bool                   `json:"applicable"`
			Problem    string                 `json:"problem,omitempty"`
		}
		rows := make([]result, 0, len(issue.KnowledgeReviews))
		for n := range issue.KnowledgeReviews {
			r := &issue.KnowledgeReviews[n]
			problem := t.KnowledgeReviewProblem(id, r)
			rows = append(rows, result{Review: *r, Applicable: problem == "", Problem: problem})
		}
		if a.json {
			a.emit(map[string]any{"id": id, "reviews": rows})
			return nil
		}
		if len(rows) == 0 {
			a.printf("No knowledge review reports.\n")
		}
		for _, row := range rows {
			a.printf("%s %s entry=%s applicable=%t\n  %s\n", row.Review.ID, row.Review.Kind, row.Review.Entry, row.Applicable, row.Review.Observation)
			if row.Problem != "" {
				a.printf("  %s\n", row.Problem)
			}
		}
		return nil
	}
	if noImpact {
		if r.Entry != "" || r.Kind != "" {
			return usageErr("--no-impact cannot be combined with --entry or --kind")
		}
		r.Kind = domain.KnowledgeNoImpact
		r.Observation = reason
	} else if r.Entry == "" || r.Kind != domain.KnowledgeUpdated && r.Kind != domain.KnowledgeUnchanged {
		return usageErr("entry reviews need --entry and --kind updated|unchanged")
	}
	r.Scopes, r.ConfigScopes = scopes, configs
	r.Snapshot, err = worktree.Fingerprint(a.store.Root, append(append([]string(nil), scopes...), configs...))
	if err != nil {
		return usageErr("cannot fingerprint knowledge review scopes: %v", err)
	}
	r.By, r.At = a.actor, a.now()
	c, err := t.PlanKnowledgeReview(id, r)
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
	a.note(activity.Event{Op: "knowledge review", Verb: "recorded knowledge review", Target: id, Area: activity.AreaKnowledge})
	a.reportWrite(writeResult{OK: true, Op: "knowledge review", ID: id, Files: files})
	return nil
}

func knowledgeEdit(a *app, args []string, isNew bool) error {
	op := "update"
	if isNew {
		op = "new"
	}
	fs := flag.NewFlagSet("knowledge "+op, flag.ContinueOnError)
	typ := fs.String("type", "", "one of "+strings.Join(domain.KnowledgeTypes, ", "))
	title := fs.String("title", "", "entry title")
	desc := fs.String("description", "", "one-line description prep guide lists for triage")
	status := fs.String("status", "", "draft or stable")
	body := fs.String("body", "", "entry body (markdown)")
	bodyFile := fs.String("body-file", "", "read the body from a file (- for stdin)")
	var scope multi
	fs.Var(&scope, "scope", "path or glob the entry applies to (repeatable, replaces the list; '' clears it)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("knowledge "+op, pos)
	if err != nil {
		return err
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if set["body"] && set["body-file"] {
		return usageErr("pass --body or --body-file, not both")
	}
	e := domain.KnowledgeEdit{Path: ref, New: isNew, Actor: a.actor, At: a.now()}
	for name, v := range map[string]*string{"type": typ, "title": title, "description": desc, "status": status} {
		if set[name] {
			val := *v
			switch name {
			case "type":
				e.Type = &val
			case "title":
				e.Title = &val
			case "description":
				e.Description = &val
			case "status":
				e.Status = &val
			}
		}
	}
	if set["scope"] {
		list := []string{}
		if !(len(scope) == 1 && scope[0] == "") {
			for _, s := range scope {
				if s == "" {
					return usageErr("--scope '' clears the scope and cannot be combined with paths")
				}
				list = append(list, s)
			}
		}
		e.Scope = &list
	}
	if set["body"] || set["body-file"] {
		text, err := a.readBody(*body, *bodyFile)
		if err != nil {
			return err
		}
		e.Body = &text
	}
	files, path, err := a.writeKnowledge(e)
	if err != nil {
		return err
	}
	verb := "updated knowledge"
	if isNew {
		verb = "wrote new knowledge"
	}
	a.note(activity.Event{Op: "knowledge " + op, Verb: verb, Target: path, Area: activity.AreaKnowledge})
	a.reportWrite(writeResult{OK: true, Op: "knowledge " + op, Entry: path, Files: files})
	return nil
}

func knowledgeConfirm(a *app, args []string) error {
	fs := flag.NewFlagSet("knowledge confirm", flag.ContinueOnError)
	drifted := fs.Bool("drifted", false, "confirm every entry the drift check flags")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if *drifted == (len(pos) > 0) {
		return usageErr("pass entries to confirm or --drifted, not both")
	}
	t, err := a.load(*drifted)
	if err != nil {
		return err
	}
	if !gitx.IsRepo(a.store.Root) {
		return &domain.Error{Code: domain.ErrInvalid, Message: "confirm needs a git repository: confirmed_commit is a commit"}
	}
	head, err := gitx.Head(a.store.Root)
	if err != nil {
		return err
	}
	paths := pos
	if *drifted {
		paths = nil
		for p, e := range t.Knowledge {
			if len(e.Drifted) > 0 {
				paths = append(paths, p)
			}
		}
		sort.Strings(paths)
	}
	var all []string
	var done []string
	for _, p := range paths {
		e := domain.KnowledgeEdit{Path: p, ConfirmedCommit: &head, Actor: a.actor, At: a.now()}
		files, path, err := a.writeKnowledge(e)
		if err != nil {
			return err
		}
		all = append(all, files...)
		done = append(done, path)
	}
	if len(done) > 0 {
		a.note(activity.Event{Op: "knowledge confirm", Verb: "confirmed knowledge", Target: strings.Join(done, " "), Area: activity.AreaKnowledge})
	}
	if a.json {
		if all == nil {
			all = []string{}
		}
		if done == nil {
			done = []string{}
		}
		a.emit(map[string]any{"ok": true, "op": "knowledge confirm", "commit": head, "entries": done, "files": all})
		return nil
	}
	if len(done) == 0 {
		a.printf("knowledge confirm: nothing drifted\n")
		return nil
	}
	noun := "entries"
	if len(done) == 1 {
		noun = "entry"
	}
	a.printf("knowledge confirm: %d %s at %s\n", len(done), noun, head[:12])
	for _, p := range done {
		a.printf("  %s\n", p)
	}
	return nil
}

// writeKnowledge runs a knowledge write through the pipeline: plan in the
// domain, render through the adapter, validate the tree, write, stage.
func (a *app) writeKnowledge(e domain.KnowledgeEdit) ([]string, string, error) {
	t, err := a.load(false)
	if err != nil {
		return nil, "", err
	}
	c, err := t.PlanKnowledge(e)
	if err != nil {
		return nil, "", err
	}
	entry, err := a.kstore.Render(c.Knowledge)
	if err != nil {
		return nil, "", err
	}
	c.KnowledgeEntry = entry
	if err := t.CheckWrite(c); err != nil {
		return nil, "", err
	}
	files, err := a.kstore.Apply(c.Knowledge)
	if err != nil {
		return nil, "", err
	}
	a.afterWrite(files)
	return files, c.Knowledge.Path, nil
}

// knowledgeBootstrap creates the bootstrap issues when the knowledge base
// is not bootstrapped and none are open.
func knowledgeBootstrap(a *app, args []string) error {
	fs := flag.NewFlagSet("knowledge bootstrap", flag.ContinueOnError)
	if _, err := parse(fs, args); err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	switch {
	case t.Bootstrapped():
		return &domain.Error{Code: domain.ErrInvalid, Message: "the knowledge base is already bootstrapped: " + domain.OverviewEntry + " exists"}
	case t.BootstrapStart() != "":
		return &domain.Error{Code: domain.ErrInvalid, Message: "bootstrap issues already exist: prep guide " + t.BootstrapStart()}
	}
	files, err := a.bootstrap()
	if err != nil {
		return err
	}
	after, err := a.load(false)
	if err != nil {
		return err
	}
	a.reportWrite(writeResult{OK: true, Op: "knowledge bootstrap", ID: after.BootstrapStart(), Files: files})
	return nil
}
