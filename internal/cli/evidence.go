package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/worktree"
)

func cmdEvidence(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" && args[0] != "record" {
		return usageErr("evidence needs show <id> or record <id> --requirement <hash> --result pass|fail --method manual|automated --observation <text> --scope <path> [--config-scope <path>] [--command <text>] [--source <reference>]")
	}
	fs := flag.NewFlagSet("evidence "+args[0], flag.ContinueOnError)
	var e domain.VerificationEvidence
	var scopes, configs multi
	if args[0] == "record" {
		fs.StringVar(&e.Requirement, "requirement", "", "requirement ID from prep policy show")
		fs.StringVar(&e.Result, "result", "", "pass or fail")
		fs.StringVar(&e.Method, "method", "", "manual or automated")
		fs.StringVar(&e.Observation, "observation", "", "what was observed, including failures")
		fs.StringVar(&e.Command, "command", "", "command already executed by the agent, never run by prep")
		fs.StringVar(&e.Source, "source", "", "report or CI reference (reported, not authenticated)")
		fs.Var(&scopes, "scope", "project-relative work path (repeatable)")
		fs.Var(&configs, "config-scope", "project-relative check configuration path (repeatable)")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("evidence needs one issue ID")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	if args[0] == "show" {
		type report struct {
			Evidence   domain.VerificationEvidence `json:"evidence"`
			Applicable bool                        `json:"applicable"`
			Problem    string                      `json:"problem,omitempty"`
		}
		records := make([]report, 0, len(t.Issues[id].Verification))
		current := t.Contract(id)
		for _, e := range t.Issues[id].Verification {
			paths := append(append([]string(nil), e.Scopes...), e.ConfigScopes...)
			fp, err := worktree.Fingerprint(a.store.Root, paths)
			r := report{Evidence: e, Applicable: err == nil && fp == e.Snapshot && current != nil && e.Contract == current.ID && !t.Stale(id)}
			if err != nil {
				r.Problem = err.Error()
			} else if fp != e.Snapshot {
				r.Problem = "scoped work or check configuration changed"
			} else if current == nil || e.Contract != current.ID {
				r.Problem = "completion contract changed; review prep policy show " + id
			} else if t.Stale(id) {
				r.Problem = "issue requirement changed; reconcile it before reusing evidence"
			}
			records = append(records, r)
		}
		if a.json {
			a.emit(map[string]any{"id": id, "records": records})
		} else {
			if len(records) == 0 {
				a.printf("No recorded verification evidence.\n")
			}
			for _, r := range records {
				a.printf("%s %s [%s, %s] applicable=%t\n  %s\n", r.Evidence.ID, r.Evidence.Result, r.Evidence.Method, r.Evidence.Provenance, r.Applicable, r.Evidence.Observation)
				if r.Problem != "" {
					a.printf("  %s\n", r.Problem)
				}
			}
		}
		return nil
	}
	e.Scopes, e.ConfigScopes = scopes, configs
	e.Snapshot, err = worktree.Fingerprint(a.store.Root, append(append([]string(nil), scopes...), configs...))
	if err != nil {
		return usageErr("cannot fingerprint evidence scopes: %v", err)
	}
	e.By, e.At = a.actor, a.now()
	c, err := t.PlanEvidence(id, e)
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
	a.reportWrite(writeResult{OK: true, Op: "evidence", ID: id, Files: files})
	return nil
}
