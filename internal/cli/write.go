package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// writeResult is the JSON answer of every write command.
type writeResult struct {
	OK    bool         `json:"ok"`
	Op    string       `json:"op"`
	ID    string       `json:"id,omitempty"`
	State domain.State `json:"state,omitempty"`
	Stale bool         `json:"stale,omitempty"`
	Files []string     `json:"files"`
}

func (a *app) reportWrite(r writeResult) {
	if r.Files == nil {
		r.Files = []string{}
	}
	if a.json {
		a.emit(r)
		return
	}
	switch {
	case r.ID != "" && r.State != "" && r.Stale:
		a.printf("%s %s: now %s, stale: the requirement or kind differs from the newest baseline; run prep define (re-enrich) or prep ack (trivial change)\n", r.Op, r.ID, r.State)
	case r.ID != "" && r.State != "":
		a.printf("%s %s: now %s\n", r.Op, r.ID, r.State)
	case r.ID != "":
		a.printf("%s %s\n", r.Op, r.ID)
	default:
		a.printf("%s: %d files\n", r.Op, len(r.Files))
	}
	for _, f := range r.Files {
		a.printf("  %s\n", f)
	}
}

func cmdInit(a *app, args []string) error {
	dir := a.root
	if dir == "" {
		dir = "."
	}
	a.root = dir
	a.store = nil
	st := openAt(dir)
	files, err := st.Init()
	if err != nil {
		return err
	}
	a.store = st
	a.afterWrite(nil, "prep: init", files)
	a.reportWrite(writeResult{OK: true, Op: "init", Files: files})
	return nil
}

func cmdNew(a *app, args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	title := fs.String("title", "", "issue title")
	kind := fs.String("kind", "", "code, manual, research or decision")
	parent := fs.String("parent", "", "parent issue id")
	body := fs.String("body", "", "requirement prose")
	bodyFile := fs.String("body-file", "", "read the requirement from a file (- for stdin)")
	var deps multi
	fs.Var(&deps, "depends-on", "dependency id (repeatable)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if *title == "" && len(pos) > 0 {
		*title = strings.Join(pos, " ")
	} else if len(pos) > 0 {
		return usageErr("unexpected arguments: %s", strings.Join(pos, " "))
	}
	text, err := a.readBody(*body, *bodyFile)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	in := domain.NewIssueInput{Title: *title, Kind: domain.Kind(*kind), Body: text}
	if *parent != "" {
		if in.Parent, err = t.Resolve(*parent); err != nil {
			return err
		}
	}
	for _, d := range deps {
		id, err := t.Resolve(d)
		if err != nil {
			return err
		}
		in.DependsOn = append(in.DependsOn, id)
	}
	c, err := t.PlanNew(in, a.now())
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
	a.afterWrite(t, fmt.Sprintf("prep: new %s %s", c.IssueID, in.Title), files)
	a.reportWrite(writeResult{OK: true, Op: "new", ID: c.IssueID, State: domain.StateOpen, Files: files})
	return nil
}

// readBody returns the requirement from --body-file (- for stdin) or --body.
func (a *app) readBody(body, file string) (string, error) {
	if file == "" {
		return body, nil
	}
	var b []byte
	var err error
	if file == "-" {
		b, err = io.ReadAll(a.stdin)
	} else {
		b, err = os.ReadFile(file)
	}
	return string(b), err
}

func cmdEdit(a *app, args []string) error {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	title := fs.String("title", "", "new title")
	kind := fs.String("kind", "", "code, manual, research or decision")
	parent := fs.String("parent", "", "parent issue id ('' removes the parent)")
	body := fs.String("body", "", "new requirement, including its Open questions section")
	bodyFile := fs.String("body-file", "", "read the new requirement from a file (- for stdin)")
	var deps multi
	fs.Var(&deps, "depends-on", "dependency id (repeatable, replaces the list; '' clears it)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("edit", pos)
	if err != nil {
		return err
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if set["body"] && set["body-file"] {
		return usageErr("pass --body or --body-file, not both")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(ref)
	if err != nil {
		return err
	}
	in := domain.EditInput{Actor: a.actor, Now: a.now()}
	if set["title"] {
		in.Title = title
	}
	if set["kind"] {
		k := domain.Kind(*kind)
		in.Kind = &k
	}
	if set["parent"] {
		p := ""
		if *parent != "" {
			if p, err = t.Resolve(*parent); err != nil {
				return err
			}
		}
		in.Parent = &p
	}
	if set["depends-on"] {
		list := []string{}
		if !(len(deps) == 1 && deps[0] == "") {
			for _, d := range deps {
				if d == "" {
					return usageErr("--depends-on '' clears the list and cannot be combined with IDs")
				}
				did, err := t.Resolve(d)
				if err != nil {
					return err
				}
				list = append(list, did)
			}
		}
		in.DependsOn = &list
	}
	if set["body"] || set["body-file"] {
		text, err := a.readBody(*body, *bodyFile)
		if err != nil {
			return err
		}
		in.Body = &text
	}
	c, err := t.PlanEdit(id, in)
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
	a.afterWrite(t, fmt.Sprintf("prep: edit %s (%s)", id, strings.Join(c.Edit.Fields, ", ")), files)
	after, err := a.load(false)
	if err != nil {
		return err
	}
	a.reportWrite(writeResult{OK: true, Op: "edit", ID: id, State: after.State(id), Stale: after.Stale(id), Files: files})
	return nil
}

func opCmd(op domain.Op) func(*app, []string) error {
	return func(a *app, args []string) error {
		fs := flag.NewFlagSet(string(op), flag.ContinueOnError)
		in := domain.Input{}
		var docs multi
		switch op {
		case domain.OpReady, domain.OpClaim:
			fs.StringVar(&in.Note, "note", "", "note recorded with the sign-off")
		case domain.OpRelease:
			fs.StringVar(&in.Reason, "reason", "", "why work stops")
		case domain.OpComplete:
			fs.StringVar(&in.Commit, "commit", "", "commit hash (code issues)")
			fs.Var(&docs, "docs", "knowledge entry updated or created (repeatable)")
			fs.StringVar(&in.NoImpact, "no-impact", "", "why the knowledge base is unaffected")
			fs.StringVar(&in.Note, "note", "", "resolution note")
		case domain.OpDrop:
			fs.StringVar(&in.Reason, "reason", "", "why the issue is dropped")
			fs.StringVar(&in.Note, "note", "", "resolution note")
		}
		pos, err := parse(fs, args)
		if err != nil {
			return err
		}
		ref, err := one(string(op), pos)
		if err != nil {
			return err
		}
		in.Docs = docs
		t, err := a.load(false)
		if err != nil {
			return err
		}
		id, err := t.Resolve(ref)
		if err != nil {
			return err
		}
		in.Actor = a.actor
		in.Now = a.now()
		c, err := t.Plan(id, op, in)
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
		a.afterWrite(t, fmt.Sprintf("prep: %s %s %s", op, id, t.Issues[id].Title), files)
		a.reportWrite(writeResult{OK: true, Op: string(op), ID: id, State: t.Apply(c).State(id), Files: files})
		return nil
	}
}

func cmdFmt(a *app, args []string) error {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	check := fs.Bool("check", false, "report files that are not canonical; change nothing")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if err := a.open(); err != nil {
		return err
	}
	cfg, _ := a.store.LoadConfig()
	files, err := a.store.Fmt(*check)
	if err != nil {
		return err
	}
	if *check {
		if a.json {
			a.emit(map[string]any{"ok": len(files) == 0, "files": nonNil(files)})
		} else {
			for _, f := range files {
				a.printf("%s\n", f)
			}
		}
		if len(files) > 0 {
			if !a.json {
				a.printf("%d files not in canonical format; run prep fmt\n", len(files))
			}
			return silentErr{&domain.Error{Code: domain.ErrCheck, Message: "not canonical"}}
		}
		return nil
	}
	a.afterWrite(&domain.Tree{Project: domain.Project{Config: cfg}}, "prep: fmt", files)
	a.reportWrite(writeResult{OK: true, Op: "fmt", Files: files})
	return nil
}

func cmdFix(a *app, args []string) error {
	if err := a.open(); err != nil {
		return err
	}
	cfg, _ := a.store.LoadConfig()
	files, err := a.store.Fix()
	if err != nil {
		return err
	}
	a.afterWrite(&domain.Tree{Project: domain.Project{Config: cfg}}, "prep: fix", files)
	a.reportWrite(writeResult{OK: true, Op: "fix", Files: files})
	return nil
}

// migrations upgrade the storage schema one version at a time. Index k
// migrates from schema k+1 to k+2.
var migrations []func(a *app) ([]string, error)

func cmdMigrate(a *app, args []string) error {
	t, err := a.load(false)
	if err != nil {
		return err
	}
	from := t.Project.Schema
	switch {
	case from == 0:
		return &domain.Error{Code: domain.ErrSchema, Message: "project.md has no schema version"}
	case from > domain.SchemaVersion:
		return &domain.Error{Code: domain.ErrSchema, Message: fmt.Sprintf("project uses schema %d; this prep supports %d — update prep", from, domain.SchemaVersion)}
	}
	var files []string
	for v := from; v < domain.SchemaVersion; v++ {
		changed, err := migrations[v-1](a)
		if err != nil {
			return err
		}
		files = append(files, changed...)
		if err := a.store.SetSchema(v + 1); err != nil {
			return err
		}
		files = append(files, ".prep/project.md")
	}
	a.afterWrite(t, fmt.Sprintf("prep: migrate schema %d to %d", from, domain.SchemaVersion), files)
	a.reportWrite(writeResult{OK: true, Op: "migrate", Files: files})
	return nil
}

func openAt(dir string) *mdstore.Store {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	return mdstore.Open(abs)
}
