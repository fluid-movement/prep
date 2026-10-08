package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/update"
)

// summary is the list view of one issue.
type summary struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Kind       domain.Kind      `json:"kind"`
	State      domain.State     `json:"state"`
	Stale      bool             `json:"stale"`
	Blocked    bool             `json:"blocked"`
	Actionable bool             `json:"actionable"`
	Parent     string           `json:"parent,omitempty"`
	DependsOn  []string         `json:"depends_on,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	Priority   domain.Priority  `json:"priority"`
	Children   int              `json:"children"`
	Progress   *domain.Progress `json:"progress,omitempty"`
	Depth      int              `json:"depth,omitempty"`
}

func summarize(t *domain.Tree, id string) summary {
	i := t.Issues[id]
	s := summary{ID: id, Title: i.Title, Kind: i.Kind, State: t.State(id), Stale: t.Stale(id), Blocked: t.Blocked(id),
		Actionable: t.Actionable(id), Parent: i.Parent, DependsOn: i.DependsOn, Tags: i.Tags, Priority: i.Priority.Effective(), Children: len(t.Children(id))}
	if s.Children > 0 {
		p := t.ChildProgress(id)
		s.Progress = &p
	}
	return s
}

func (s summary) flags() string {
	var f []string
	if s.Stale {
		f = append(f, "stale")
	}
	if s.Blocked {
		f = append(f, "blocked")
	}
	if s.Actionable {
		f = append(f, "actionable")
	}
	if s.Progress != nil {
		f = append(f, fmt.Sprintf("%d/%d", s.Progress.Done+s.Progress.Dropped, s.Progress.Total))
	}
	return strings.Join(f, ",")
}

func (a *app) printSummaries(ss []summary) {
	if len(ss) == 0 {
		a.printf("no issues\n")
		return
	}
	for _, s := range ss {
		indent := strings.Repeat("  ", s.Depth)
		a.printf("%s  %-11s %-8s %-18s %s%s%s%s\n", s.ID, s.State, s.Kind, s.flags(), indent, priorityMark(s.Priority), s.Title, tagSuffix(s.Tags))
	}
}

// byPriority orders summaries as Tree.ByPriority orders their IDs.
func byPriority(t *domain.Tree, ss []summary) []summary {
	at := map[string]summary{}
	ids := make([]string, len(ss))
	for k, s := range ss {
		at[s.ID], ids[k] = s, s.ID
	}
	out := make([]summary, 0, len(ss))
	for _, id := range t.ByPriority(ids) {
		out = append(out, at[id])
	}
	return out
}

// priorityMark prefixes a title in text output; medium, the default, has none.
func priorityMark(p domain.Priority) string {
	switch p.Effective() {
	case domain.PriorityCritical:
		return "!crit "
	case domain.PriorityHigh:
		return "!high "
	case domain.PriorityLow:
		return "low "
	}
	return ""
}

func cmdList(a *app, args []string) error {
	tree := false
	var query []string
	view := ""
	for k := 0; k < len(args); k++ {
		switch {
		case args[k] == "--tree":
			tree = true
		case args[k] == "--view":
			if k+1 >= len(args) {
				return usageErr("--view needs a name")
			}
			k++
			view = args[k]
		case strings.HasPrefix(args[k], "--view="):
			view = strings.TrimPrefix(args[k], "--view=")
		default:
			query = append(query, args[k])
		}
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	if view != "" {
		flags, ok := t.Project.Config.Views[view]
		if !ok {
			return &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("no saved view %q; see prep views", view)}
		}
		query = append(strings.Fields(flags), query...)
	}
	f, err := domain.ParseFilter(query)
	if err != nil {
		return usageErr("%v", err)
	}
	ids, err := t.Query(f)
	if err != nil {
		return err
	}
	var out []summary
	if tree {
		matched := map[string]bool{}
		for _, id := range ids {
			matched[id] = true
		}
		all := t.TreeOrder(t.WithAncestors(ids))
		set := map[string]bool{}
		for _, id := range all {
			set[id] = true
		}
		for _, id := range all {
			s := summarize(t, id)
			for _, anc := range t.Ancestors(id) {
				if set[anc] {
					s.Depth++
				}
			}
			out = append(out, s)
		}
	} else {
		for _, id := range ids {
			out = append(out, summarize(t, id))
		}
	}
	if a.json {
		if out == nil {
			out = []summary{}
		}
		a.emit(map[string]any{"issues": out})
		return nil
	}
	a.printSummaries(out)
	return nil
}

func cmdNext(a *app, args []string) error {
	fs := flag.NewFlagSet("next", flag.ContinueOnError)
	var under multi
	fs.Var(&under, "under", "limit to descendants of an issue")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	b := true
	ids, err := t.Query(domain.Filter{Actionable: &b, Under: under})
	if err != nil {
		return err
	}
	out := []summary{}
	for _, id := range ids {
		out = append(out, summarize(t, id))
	}
	if a.json {
		a.emit(map[string]any{"issues": out})
		return nil
	}
	a.printSummaries(out)
	return nil
}

func cmdViews(a *app, args []string) error {
	t, err := a.load(false)
	if err != nil {
		return err
	}
	views := t.Project.Config.Views
	if a.json {
		if views == nil {
			views = map[string]string{}
		}
		a.emit(map[string]any{"views": views})
		return nil
	}
	for _, n := range domain.ViewNames(t.Project.Config) {
		a.printf("%-14s %s\n", n, views[n])
	}
	return nil
}

// cmdFlags lists every query flag with the values it accepts now, for
// writing views and list queries.
func cmdFlags(a *app, args []string) error {
	fs := flag.NewFlagSet("flags", flag.ContinueOnError)
	if _, err := parse(fs, args); err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	flags := t.QueryFlags()
	var views []domain.FlagValue
	for _, n := range domain.ViewNames(t.Project.Config) {
		views = append(views, domain.FlagValue{Value: n, Label: t.Project.Config.Views[n]})
	}
	flags = append(flags,
		domain.FlagInfo{Name: "view", Kind: domain.FlagList, Desc: "prep list only: a saved view's query, before the other flags", Values: views},
		domain.FlagInfo{Name: "tree", Kind: domain.FlagBool, Desc: "prep list only: show the hierarchy"},
	)
	if a.json {
		a.emit(map[string]any{"flags": flags})
		return nil
	}
	a.printf("A leading not- negates a value: --state not-open is every issue that is not open (! works too, quoted in the shell).\n\n")
	for k, f := range flags {
		// Switches stay together; flags with values get their own block.
		if k > 0 && (f.Kind != domain.FlagBool || flags[k-1].Kind != domain.FlagBool) {
			a.printf("\n")
		}
		head := "--" + f.Name
		switch {
		case f.Name == "under":
			head += " [not-]<id>  (repeatable)"
		case f.Name == "view":
			head += " <name>"
		case f.Kind == domain.FlagList:
			head += " [not-]<value>[,...]"
		case f.Kind == domain.FlagText:
			head += " <text>  (repeatable)"
		}
		a.printf("%s  %s", head, f.Desc)
		if f.Count != nil {
			a.printf(" (%d issues)", *f.Count)
		}
		a.printf("\n")
		if f.Kind == domain.FlagList && len(f.Values) == 0 {
			a.printf("  (none yet)\n")
		}
		for _, v := range f.Values {
			switch {
			case f.Name == "view":
				q := v.Label
				if q == "" {
					q = "(all issues)"
				}
				a.printf("  %-26s %s\n", v.Value, q)
			case v.Label != "":
				a.printf("  %-26s %4d  %s\n", v.Value, v.Count, v.Label)
			default:
				a.printf("  %-26s %4d\n", v.Value, v.Count)
			}
		}
	}
	return nil
}

func cmdShow(a *app, args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("show", pos)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(ref)
	if err != nil {
		return err
	}
	i := t.Issues[id]
	pending := a.resolveEvidence(t, id)
	dod, opt := t.EffectiveDoD(id)
	s := summarize(t, id)
	if a.json {
		a.emit(map[string]any{
			"issue": i, "dir": mdstore.IssueDir(id), "priority": i.Priority.Effective(), "state": s.State, "stale": s.Stale, "blocked": s.Blocked,
			"actionable": s.Actionable, "children": nonNil(t.Children(id)), "blocks": nonNil(t.Blocks(id)),
			"progress": s.Progress, "definition_of_done": nonNil(dod), "dod_opt_outs": opt,
			"context": i.Context, "findings": i.Findings, "history": i.History, "evidence_pending": pending,
		})
		return nil
	}
	a.printf("%s  %s\n", id, i.Title)
	a.printf("kind: %s   state: %s", i.Kind, s.State)
	if f := s.flags(); f != "" {
		a.printf("   (%s)", f)
	}
	a.printf("\ndir: %s\n", mdstore.IssueDir(id))
	if i.Parent != "" {
		a.printf("parent: %s %s\n", i.Parent, titleOf(t, i.Parent))
	}
	for _, c := range t.Children(id) {
		a.printf("child: %s [%s] %s\n", c, t.State(c), titleOf(t, c))
	}
	if len(i.Tags) > 0 {
		a.printf("tags: %s\n", strings.Join(i.Tags, ", "))
	}
	if i.Priority.Effective() != domain.PriorityMedium {
		a.printf("priority: %s\n", i.Priority)
	}
	for _, d := range i.DependsOn {
		a.printf("depends on: %s [%s] %s\n", d, t.State(d), titleOf(t, d))
	}
	for _, b := range t.Blocks(id) {
		a.printf("blocks: %s [%s] %s\n", b, t.State(b), titleOf(t, b))
	}
	a.printf("\n## Requirement\n\n%s\n", orNone(i.Prose))
	if strings.TrimSpace(i.OpenQuestions) != "" {
		a.printf("\n## Open questions\n\n%s\n", i.OpenQuestions)
	}
	if len(i.Criteria) > 0 {
		a.printf("\n## Acceptance\n\n")
		for k, c := range i.Criteria {
			mark := " "
			if c.Checked {
				mark = "x"
			}
			a.printf("%d. [%s] %s\n", k+1, mark, c.Text)
		}
	}
	if len(dod) > 0 {
		a.printf("\n## Definition of Done (effective)\n\n")
		for _, d := range dod {
			a.printf("- %s\n", d)
		}
	}
	if strings.TrimSpace(i.Context) != "" {
		a.printf("\n## Context\n\n%s\n", i.Context)
	}
	if len(i.Decisions) > 0 {
		a.printf("\n## Decisions\n\n")
		for _, d := range i.Decisions {
			extra := ""
			if d.Supersedes != "" {
				extra += " supersedes " + d.Supersedes
			}
			if d.Outcome {
				extra += " (outcome)"
			}
			a.printf("- %s: %s [%s]%s\n", d.ID, d.Title, d.Date, extra)
		}
	}
	if i.Findings != nil && *i.Findings != "" {
		a.printf("\n## Findings\n\n%s\n", *i.Findings)
	}
	a.printf("\n## Records\n\n")
	for _, b := range i.Baselines {
		kind := "defined"
		if b.Ack {
			kind = "acknowledged"
		}
		a.printf("- baseline %s: %s by %s\n", b.Name, kind, b.By)
	}
	if i.Ready != nil {
		a.printf("- ready: by %s at %s against baseline %s\n", i.Ready.By, i.Ready.At.Format(time.RFC3339), i.Ready.Baseline)
	}
	if i.Claim != nil {
		a.printf("- claim: by %s at %s\n", i.Claim.By, i.Claim.At.Format(time.RFC3339))
	}
	if r := i.Resolution; r != nil {
		a.printf("- resolution: %s by %s at %s", r.Outcome, r.By, r.At.Format(time.RFC3339))
		if r.Reason != "" {
			a.printf(" — %s", r.Reason)
		}
		if r.Evidence != "" {
			a.printf(" (commit %s)", r.Evidence)
		} else if pending {
			a.printf(" (evidence: the commit that adds its resolution.md, not made yet)")
		}
		a.printf("\n")
	}
	return nil
}

func titleOf(t *domain.Tree, id string) string {
	if i := t.Issues[id]; i != nil {
		return i.Title
	}
	return "(missing)"
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// resolveEvidence fills a done code issue's evidence in memory when it was
// completed without --commit: the commit that added its resolution.md. It
// reports whether that commit is not made yet.
func (a *app) resolveEvidence(t *domain.Tree, id string) (pending bool) {
	i := t.Issues[id]
	r := i.Resolution
	if r == nil || r.Evidence != "" || r.Outcome != domain.OutcomeDone || i.Kind != domain.KindCode || t.IsParent(id) {
		return false
	}
	if !gitx.IsRepo(a.store.Root) {
		return true
	}
	r.Evidence = gitx.AddedIn(a.store.Root, mdstore.IssueDir(id)+"/resolution.md")
	return r.Evidence == ""
}

// touchedFiles lists code paths an issue touched, from its commit evidence.
func (a *app) touchedFiles(t *domain.Tree, id string) []string {
	a.resolveEvidence(t, id)
	r := t.Issues[id].Resolution
	if r == nil || r.Evidence == "" || !gitx.IsRepo(a.store.Root) || !gitx.CommitExists(a.store.Root, r.Evidence) {
		return nil
	}
	files, _ := gitx.FilesInCommit(a.store.Root, r.Evidence)
	return files
}

func cmdGuide(a *app, args []string) error {
	fs := flag.NewFlagSet("guide", flag.ContinueOnError)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("guide", pos)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(ref)
	if err != nil {
		return err
	}
	g := t.BuildGuide(id, mdstore.IssueDir, a.touchedFiles(t, id))
	if a.json {
		if g.Knowledge == nil {
			g.Knowledge = []domain.Candidate{}
		}
		a.emit(g)
		return nil
	}
	for _, al := range g.Alerts {
		a.printf("! %s\n\n", al)
	}
	a.printf("%s  %s\n", g.ID, g.Title)
	a.printf("kind: %s   state: %s", g.Kind, g.State)
	if g.Stale {
		a.printf("   STALE")
	}
	if g.Parent {
		a.printf("   (parent)")
	}
	a.printf("\nstep: %s\n\n", g.Step)
	for _, s := range g.Instructions {
		a.printf("- %s\n", s)
	}
	a.printf("\nTransitions:\n")
	for _, tr := range g.Transitions {
		mark := "ok     "
		if !tr.Allowed {
			mark = "blocked"
		}
		a.printf("  [%s] %s\n", mark, tr.Command)
		for _, u := range tr.Unmet {
			a.printf("            - %s (%s)\n", u.Message, u.Code)
		}
	}
	if len(g.Inputs) > 0 {
		a.printf("\nRead:\n")
		for _, p := range g.Inputs {
			a.printf("  %s — %s\n", p.Path, p.What)
		}
	}
	if len(g.Knowledge) > 0 {
		a.printf("\nKnowledge candidates (read what you need: prep knowledge show <entry>[#section], --outline first for long ones):\n")
		for _, c := range g.Knowledge {
			a.printf("  %s — %s: %s [%s]\n", c.Path, c.Title, c.Description, c.Via)
		}
	}
	if len(g.Outputs) > 0 {
		a.printf("\nWrite:\n")
		for _, p := range g.Outputs {
			if p.Command != "" {
				a.printf("  %s — %s (%s)\n", p.Command, p.What, p.Path)
			} else {
				a.printf("  %s — %s\n", p.Path, p.What)
			}
		}
	}
	if len(g.DoD) > 0 {
		a.printf("\nDefinition of Done:\n")
		for _, d := range g.DoD {
			a.printf("  - %s\n", d)
		}
	}
	if len(g.Diagnostics) > 0 {
		a.printf("\nDiagnostics:\n")
		for _, d := range g.Diagnostics {
			a.printf("  %s\n", formatDiag(d))
		}
	}
	return nil
}

func cmdCheck(a *app, args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	noDrift := fs.Bool("no-drift", false, "skip the knowledge drift check")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	t, err := a.load(!*noDrift)
	if err != nil {
		return err
	}
	ds := domain.Validate(t)
	errs, warns := 0, 0
	for _, d := range ds {
		if d.Severity == domain.SevError {
			errs++
		} else {
			warns++
		}
	}
	if a.json {
		if ds == nil {
			ds = []domain.Diagnostic{}
		}
		a.emit(map[string]any{"ok": errs == 0, "errors": errs, "warnings": warns, "issues": len(t.Issues), "entries": len(t.Knowledge), "diagnostics": ds})
	} else {
		for _, d := range ds {
			a.printf("%s\n", formatDiag(d))
		}
		a.printf("%d issues, %d knowledge entries: %d errors, %d warnings\n", len(t.Issues), len(t.Knowledge), errs, warns)
	}
	if errs > 0 {
		return silentErr{&domain.Error{Code: domain.ErrCheck, Message: fmt.Sprintf("%d errors", errs)}}
	}
	return nil
}

// silentErr sets the exit code without printing again.
type silentErr struct{ err *domain.Error }

func (s silentErr) Error() string { return s.err.Message }

// primeBrief is the capped session-start briefing.
type primeBrief struct {
	Issues     int            `json:"issues"`
	Actionable int            `json:"actionable"`
	Parents    []summary      `json:"parents"`
	Next       []summary      `json:"next"`
	Stale      []summary      `json:"stale"`
	Claims     []claimSummary `json:"claims"`
	Check      checkSummary   `json:"check"`
	Bootstrap  string         `json:"bootstrap,omitempty"` // alert while the knowledge base is not bootstrapped
	Plugin     string         `json:"plugin,omitempty"`    // alert when the harness plugin and the binary differ
	Hint       string         `json:"hint"`
}

type claimSummary struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	By    string    `json:"by"`
	Since time.Time `json:"since"`
}

type checkSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

func cmdPrime(a *app, args []string) error {
	fs := flag.NewFlagSet("prime", flag.ContinueOnError)
	max := fs.Int("max", 8, "maximum entries per section")
	hook := fs.Bool("hook", false, "run from a harness hook: print nothing outside a prep project")
	plugin := fs.String("plugin", "", "harness plugin directory; warn when its version differs from this binary")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		var de *domain.Error
		if *hook && errors.As(err, &de) && de.Code == domain.ErrNoProject {
			return nil
		}
		return err
	}
	cap := func(ss []summary) []summary {
		if len(ss) > *max {
			return ss[:*max]
		}
		return ss
	}
	b := primeBrief{Issues: len(t.Issues), Parents: []summary{}, Next: []summary{}, Stale: []summary{}, Claims: []claimSummary{},
		Hint: "Run prep guide <id> for the issue you work on; it lists what to read and where outputs go. Read knowledge with prep knowledge find <words> or prep knowledge show <entry>[#section], not the files, and tell subagents to do the same."}
	for _, id := range t.IDs() {
		i := t.Issues[id]
		s := t.State(id)
		top := i.Parent == "" || t.Issues[i.Parent] == nil
		if top && t.IsParent(id) && !s.Terminal() {
			b.Parents = append(b.Parents, summarize(t, id))
		}
		if t.Actionable(id) {
			b.Actionable++
			b.Next = append(b.Next, summarize(t, id))
		}
		if t.Stale(id) {
			b.Stale = append(b.Stale, summarize(t, id))
		}
		if s == domain.StateInProgress {
			b.Claims = append(b.Claims, claimSummary{ID: id, Title: i.Title, By: i.Claim.By, Since: i.Claim.At})
		}
	}
	b.Next, b.Stale = byPriority(t, b.Next), byPriority(t, b.Stale)
	b.Parents, b.Next, b.Stale = cap(b.Parents), cap(b.Next), cap(b.Stale)
	if len(b.Claims) > *max {
		b.Claims = b.Claims[:*max]
	}
	for _, d := range domain.Validate(t) {
		if d.Severity == domain.SevError {
			b.Check.Errors++
		} else {
			b.Check.Warnings++
		}
	}
	b.Bootstrap = t.BootstrapAlert()
	if *plugin != "" {
		b.Plugin = pluginAlert(*plugin)
	}
	if a.json {
		a.emit(b)
		return nil
	}
	if b.Plugin != "" {
		a.printf("! %s\n\n", b.Plugin)
	}
	if b.Bootstrap != "" {
		a.printf("! %s\n\n", b.Bootstrap)
	}
	a.printf("prep: %d issues, %d actionable; check: %d errors, %d warnings\n", b.Issues, b.Actionable, b.Check.Errors, b.Check.Warnings)
	section := func(title string, ss []summary) {
		if len(ss) == 0 {
			return
		}
		a.printf("\n%s:\n", title)
		for _, s := range ss {
			extra := ""
			if s.Progress != nil {
				extra = fmt.Sprintf(" (%d/%d resolved)", s.Progress.Done+s.Progress.Dropped, s.Progress.Total)
			}
			a.printf("  %s [%s] %s%s%s\n", s.ID, s.State, priorityMark(s.Priority), s.Title, extra)
		}
	}
	section("Top-level parents", b.Parents)
	section("Actionable (prep next)", b.Next)
	section("Stale (requirement changed since baseline)", b.Stale)
	if len(b.Claims) > 0 {
		a.printf("\nIn progress:\n")
		for _, c := range b.Claims {
			a.printf("  %s %s — claimed by %s at %s\n", c.ID, c.Title, c.By, c.Since.Format(time.RFC3339))
		}
	}
	if b.Check.Errors > 0 {
		a.printf("\nprep check reports %d errors; run prep check.\n", b.Check.Errors)
	}
	a.printf("\n%s\n", b.Hint)
	return nil
}

func cmdSkill(a *app, args []string) error {
	a.printf("%s", skillText)
	return nil
}

func cmdVersion(a *app, args []string) error {
	if a.json {
		a.emit(map[string]any{"version": Version, "schema": domain.SchemaVersion})
		return nil
	}
	a.printf("prep %s (schema %d)\n", Version, domain.SchemaVersion)
	return nil
}

// tagSuffix renders tags after a title as "  #a #b".
func tagSuffix(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "  #" + strings.Join(tags, " #")
}

// pluginAlert compares a harness plugin's version with this binary's and
// names the fix when they differ. Development builds on either side are
// not compared.
func pluginAlert(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		return ""
	}
	var m struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	pv, bv := "v"+strings.TrimPrefix(m.Version, "v"), Version
	if !update.IsRelease(pv) || !update.IsRelease(bv) || strings.TrimPrefix(pv, "v") == strings.TrimPrefix(bv, "v") {
		return ""
	}
	if newer, err := update.Newer(pv, bv); err == nil && newer {
		return fmt.Sprintf("The prep plugin (%s) is older than the prep binary (%s): run prep setup --refresh.", pv, bv)
	}
	return fmt.Sprintf("The prep binary (%s) is older than the prep plugin (%s): run prep update.", bv, pv)
}
