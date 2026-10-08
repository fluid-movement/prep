package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// note sets the activity event a successful command records; Main
// appends it after the command returns without error.
func (a *app) note(e activity.Event) {
	if e.Kind == "" {
		e.Kind = activity.KindPrep
	}
	a.event = &e
}

// noteRead is note for read commands: the event also carries the size of
// what the command printed.
func (a *app) noteRead(e activity.Event) {
	a.note(e)
	a.eventReads = true
}

// prepDir is the .prep directory of the opened project, or "" outside one.
func (a *app) prepDir() string {
	if a.store == nil {
		return ""
	}
	dir := filepath.Join(a.store.Root, mdstore.Dir)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return ""
	}
	return dir
}

// recordEvent appends the noted event. Activity is UI state: a failure to
// record never fails the command. Reads are recorded only for a named
// actor (--by or PREP_ACTOR), as agents run them: panels, hooks and people
// reading without a name would bury what the agent does.
func (a *app) recordEvent(printed int) {
	dir := a.prepDir()
	if a.event == nil || dir == "" || a.eventReads && !a.named {
		return
	}
	e := *a.event
	e.At, e.Actor, e.Session = a.now().UTC(), a.actor, os.Getenv("PREP_SESSION")
	if a.eventReads {
		e.Chars = printed
	}
	_ = activity.Append(dir, e)
}

// counter counts what a command prints.
type counter struct {
	w io.Writer
	n int
}

func (c *counter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
}

// issueEvent is the event of a command on one issue.
func issueEvent(t *domain.Tree, op, id, verb string) activity.Event {
	return activity.Event{Op: op, Issue: id, Verb: verb, Target: titleOf(t, id), Area: activity.AreaIssue}
}

var opVerbs = map[domain.Op]string{
	domain.OpDefine:   "defined the requirement",
	domain.OpAck:      "acknowledged the requirement change",
	domain.OpReady:    "signed off the enrichment",
	domain.OpClaim:    "claimed the issue",
	domain.OpRelease:  "released the issue",
	domain.OpComplete: "completed the issue",
	domain.OpDrop:     "dropped the issue",
	domain.OpContext:  "wrote the context",
	domain.OpFindings: "wrote the findings",
	domain.OpDoD:      "changed the Definition of Done",
}

// recordVerb says what a record write did.
func recordVerb(op domain.Op, in domain.RecordInput, c *domain.Change) string {
	switch op {
	case domain.OpDecide:
		if c.Decision != nil {
			return fmt.Sprintf("recorded decision %s: %s", c.Decision.ID, c.Decision.Title)
		}
		return "recorded a decision"
	case domain.OpLog:
		return "logged: " + in.Text
	case domain.OpCriterion:
		var parts []string
		adds := 0
		for _, o := range in.Acceptance {
			switch o.Op {
			case "add":
				adds++
			case "check":
				parts = append(parts, fmt.Sprintf("checked criterion %d", o.Index))
			case "uncheck":
				parts = append(parts, fmt.Sprintf("unchecked criterion %d", o.Index))
			case "remove":
				parts = append(parts, fmt.Sprintf("removed criterion %d", o.Index))
			}
		}
		switch {
		case adds == 1:
			parts = append(parts, "added a criterion")
		case adds > 1:
			parts = append(parts, fmt.Sprintf("added %d criteria", adds))
		}
		if len(parts) == 0 {
			return "changed the criteria"
		}
		return strings.Join(parts, ", ")
	}
	if v, ok := opVerbs[op]; ok {
		return v
	}
	return string(op)
}

func cmdFocus(a *app, args []string) error {
	fs := flag.NewFlagSet("focus", flag.ContinueOnError)
	clear := fs.Bool("clear", false, "clear the current issue")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	if a.prepDir() == "" {
		return &domain.Error{Code: domain.ErrNoProject, Message: "no .prep directory"}
	}
	if *clear {
		if len(pos) > 0 {
			return usageErr("pass an issue or --clear, not both")
		}
		a.note(activity.Event{Kind: activity.KindFocus, Verb: "cleared the focus"})
		if a.json {
			a.emit(map[string]any{"ok": true, "actor": a.actor, "issue": nil})
		} else {
			a.printf("focus cleared for %s\n", a.actor)
		}
		return nil
	}
	ref, err := one("focus", pos)
	if err != nil {
		return err
	}
	id, err := t.Resolve(ref)
	if err != nil {
		return err
	}
	a.note(activity.Event{Kind: activity.KindFocus, Issue: id, Verb: "turned to the issue", Target: titleOf(t, id), Area: activity.AreaIssue})
	if a.json {
		a.emit(map[string]any{"ok": true, "actor": a.actor, "issue": id})
	} else {
		a.printf("focus: %s %s (%s)\n", id, titleOf(t, id), a.actor)
	}
	return nil
}

func cmdActivity(a *app, args []string) error {
	if len(args) > 0 && args[0] == "add" {
		return activityAdd(a, args[1:])
	}
	fs := flag.NewFlagSet("activity", flag.ContinueOnError)
	max := fs.Int("max", 20, "number of newest events")
	actor := fs.String("actor", "", "only this actor's events")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) > 0 {
		return usageErr("usage: prep activity [--max N] [--actor a] | add")
	}
	if _, err := a.load(false); err != nil {
		return err
	}
	dir := a.prepDir()
	events, err := activity.Read(dir, 0)
	if err != nil {
		return err
	}
	if *actor != "" {
		var mine []activity.Event
		for _, e := range events {
			if e.Actor == *actor {
				mine = append(mine, e)
			}
		}
		events = mine
	}
	foci := activity.Foci(events)
	if *max > 0 && len(events) > *max {
		events = events[len(events)-*max:]
	}
	if a.json {
		if events == nil {
			events = []activity.Event{}
		}
		a.emit(map[string]any{"events": events, "focus": sortedFoci(foci)})
		return nil
	}
	if len(events) == 0 {
		a.printf("no activity recorded yet\n")
		return nil
	}
	for _, e := range events {
		line := fmt.Sprintf("%s  %-22s %-8s", e.At.Local().Format("15:04:05"), e.Actor, e.Kind)
		if e.Issue != "" {
			line += " " + short(e.Issue)
		}
		what := e.Verb
		if what == "" {
			what = e.Op
		}
		if e.Target != "" {
			what += " · " + e.Target
		}
		a.printf("%s %s\n", line, what)
	}
	return nil
}

// short is the issue suffix the TUI shows.
func short(id string) string {
	if len(id) > 6 {
		return id[len(id)-6:]
	}
	return id
}

func sortedFoci(foci map[string]activity.Focus) []activity.Focus {
	out := make([]activity.Focus, 0, len(foci))
	for _, f := range foci {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// activityAdd takes harness events, one JSON object per line on stdin.
func activityAdd(a *app, args []string) error {
	fs := flag.NewFlagSet("activity add", flag.ContinueOnError)
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if _, err := a.load(false); err != nil {
		return err
	}
	dir := a.prepDir()
	if dir == "" {
		return &domain.Error{Code: domain.ErrNoProject, Message: "no .prep directory"}
	}
	events, err := activity.Parse(a.stdin, 0)
	if err != nil {
		return usageErr("activity add: %v", err)
	}
	if len(events) == 0 {
		return usageErr("activity add reads events as JSON lines on stdin")
	}
	var bad []string
	for k, e := range events {
		if err := e.Validate(); err != nil {
			bad = append(bad, fmt.Sprintf("event %d: %v", k+1, err))
		}
	}
	if len(bad) > 0 {
		return usageErr("%s", strings.Join(bad, "; "))
	}
	for _, e := range events {
		if e.At.IsZero() {
			e.At = a.now().UTC()
		}
		if e.Actor == "" {
			e.Actor = a.actor
		}
		if e.Session == "" {
			e.Session = os.Getenv("PREP_SESSION")
		}
		if err := activity.Append(dir, e); err != nil {
			return err
		}
	}
	if a.json {
		a.emit(map[string]any{"ok": true, "added": len(events)})
	}
	return nil
}

// recentFoci are the newest current issues for prime, at most n, from the
// last day.
func (a *app) recentFoci(t *domain.Tree, n int) []focusSummary {
	dir := a.prepDir()
	if dir == "" {
		return nil
	}
	events, err := activity.Read(dir, 0)
	if err != nil {
		return nil
	}
	var out []focusSummary
	for _, f := range sortedFoci(activity.Foci(events)) {
		if len(out) == n || a.now().Sub(f.At) > 24*time.Hour {
			break
		}
		if t.Issues[f.Issue] == nil {
			continue
		}
		out = append(out, focusSummary{ID: f.Issue, Title: titleOf(t, f.Issue), State: t.State(f.Issue), By: f.Actor, At: f.At})
	}
	return out
}

type focusSummary struct {
	ID    string       `json:"id"`
	Title string       `json:"title"`
	State domain.State `json:"state"`
	By    string       `json:"by"`
	At    time.Time    `json:"at"`
}
