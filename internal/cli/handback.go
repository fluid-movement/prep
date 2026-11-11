package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func cmdHandback(a *app, args []string) error {
	fs := flag.NewFlagSet("handback", flag.ContinueOnError)
	check := fs.Bool("check", false, "hand back a checker's report")
	report := fs.String("report", "", "report file (- for stdin; default is the issue's hand-off directory)")
	commit := fs.String("commit", "", "commit hash passed to completion")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("handback", pos)
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
	id, err := t.Resolve(ref)
	if err != nil {
		return err
	}
	if *report == "" {
		_, name := mdstore.HandoffFiles(*check)
		*report = a.handoffPath(id, name)
	}
	text, err := a.readBody("", *report)
	if err != nil {
		return &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("cannot read report %s: %v; use --report <file> (or --report - for stdin)", *report, err)}
	}
	r, err := domain.ParseReport(text, t.Issues[id], *check)
	if err != nil {
		return err
	}
	// Plan and validate the entire batch in memory before writing anything,
	// including knowledge records and the final completion gates.
	planned := t
	var changes []*domain.Change
	var lines []string
	add := func(c *domain.Change, err error) error {
		if err != nil {
			return err
		}
		if c.Knowledge != nil {
			c.KnowledgeEntry, err = a.kstore.Render(c.Knowledge)
			if err != nil {
				return err
			}
		}
		if err := planned.CheckWrite(c); err != nil {
			return err
		}
		changes = append(changes, c)
		planned = planned.Apply(c)
		line := string(c.Op) + " " + id
		if c.Knowledge != nil {
			line = string(c.Op) + " " + c.Knowledge.Path
		}
		lines = append(lines, line)
		return nil
	}
	log := func(text string) error {
		return add(planned.PlanRecord(id, domain.OpLog, domain.RecordInput{Actor: a.actor, Now: a.now(), Text: text}))
	}
	for _, group := range []struct {
		typ     string
		entries []domain.ReportEntry
	}{{"decision", r.Decisions}, {"finding", r.Findings}} {
		for _, e := range group.entries {
			recordID, err := domain.NewULID(a.now(), a.entropy)
			if err != nil {
				return err
			}
			if err := add(planned.PlanKnowledgeRecord(domain.KnowledgeRecord{
				ID: recordID, Type: group.typ, Title: e.Title, Description: e.Description,
				Body: e.Body, Scope: e.Scope, Issue: id, At: a.now(),
			})); err != nil {
				return err
			}
		}
	}
	var failed []string
	resolves := *check || t.Issues[id].Kind == domain.KindResearch
	if resolves {
		for _, c := range r.Criteria {
			op := "check"
			if !c.Met {
				op = "uncheck"
				failed = append(failed, fmt.Sprintf("criterion %d not met: %s", c.Number, c.Evidence))
			}
			if err := add(planned.PlanRecord(id, domain.OpCriterion, domain.RecordInput{
				Actor: a.actor, Now: a.now(), Acceptance: []domain.AcceptanceOp{{Op: op, Index: c.Number}},
			})); err != nil {
				return err
			}
		}
		if r.Problems != "" {
			failed = append(failed, r.Problems)
		}
		if len(failed) > 0 {
			if err := log("Check failed: " + strings.Join(failed, "; ")); err != nil {
				return err
			}
		}
	} else {
		if err := log("Implemented, ready for check: " + r.Summary); err != nil {
			return err
		}
		if r.Problems != "" {
			if err := log("Problems: " + r.Problems); err != nil {
				return err
			}
		}
	}
	// Supporting records precede state changes; the resolution is always last.
	if planned.State(id) == domain.StateReady {
		if err := add(planned.Plan(id, domain.OpClaim, domain.Input{Actor: a.actor, Now: a.now()})); err != nil {
			return err
		}
	}
	if resolves && len(failed) == 0 {
		if err := add(planned.Plan(id, domain.OpComplete, domain.Input{Actor: a.actor, Now: a.now(), Commit: *commit, Note: r.Summary})); err != nil {
			return err
		}
	}
	files, err := a.applyAll(t, changes)
	if err != nil {
		return err
	}
	if *report != "-" {
		ext := filepath.Ext(*report)
		done := strings.TrimSuffix(*report, ext) + ".done" + ext
		if err := os.Rename(*report, done); err != nil {
			return fmt.Errorf("records written, but cannot move report aside: %w", err)
		}
	}
	a.note(issueEvent(planned, "handback", id, "handed back the report"))
	if a.json {
		a.reportWrite(writeResult{OK: true, Op: "handback", ID: id, State: planned.State(id), Files: files})
	} else {
		for _, line := range lines {
			a.printf("%s\n", line)
		}
	}
	return nil
}
