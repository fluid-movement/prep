package cli

import (
	"flag"
	"sort"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
)

// cmdKnowledge dispatches the knowledge reads (list, find, show) and writes
// (new, update, confirm, bootstrap).
func cmdKnowledge(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("knowledge needs a subcommand: list, find, show, new, update, confirm or bootstrap")
	}
	switch args[0] {
	case "list":
		return knowledgeList(a, args[1:])
	case "find":
		return knowledgeFind(a, args[1:])
	case "show":
		return knowledgeShow(a, args[1:])
	case "new":
		return knowledgeEdit(a, args[1:], true)
	case "update":
		return knowledgeEdit(a, args[1:], false)
	case "confirm":
		return knowledgeConfirm(a, args[1:])
	case "bootstrap":
		return knowledgeBootstrap(a, args[1:])
	}
	return usageErr("unknown knowledge subcommand %q: use list, find, show, new, update, confirm or bootstrap", args[0])
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
