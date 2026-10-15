package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/worktree"
)

func writeApproval(a *app, t *domain.Tree, c *domain.Change, op string) error {
	if err := t.CheckWrite(c); err != nil {
		return err
	}
	files, err := a.store.Apply(c)
	if err != nil {
		return err
	}
	a.afterWrite(files)
	a.reportWrite(writeResult{OK: true, Op: op, ID: c.IssueID, Files: files, Next: "prep guide " + c.IssueID})
	return nil
}

func cmdException(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" && args[0] != "approve" {
		return usageErr("exception needs show <id> or approve <id> --requirement <hash> --reason <text> --scope <path> --by human:<name>")
	}
	fs := flag.NewFlagSet("exception "+args[0], flag.ContinueOnError)
	var approval domain.CompletionApproval
	var scopes, configs multi
	if args[0] == "approve" {
		fs.StringVar(&approval.Requirement, "requirement", "", "requirement ID from prep policy show")
		fs.StringVar(&approval.Reason, "reason", "", "why the human approves this exception")
		fs.Var(&scopes, "scope", "project-relative work path (repeatable)")
		fs.Var(&configs, "config-scope", "project-relative check configuration path (repeatable)")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("exception needs one issue ID")
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
			Approval   domain.CompletionApproval `json:"approval"`
			Applicable bool                      `json:"applicable"`
			Problem    string                    `json:"problem,omitempty"`
		}
		records := []report{}
		for _, approval := range t.Issues[id].Approvals {
			if approval.Kind != domain.ApprovalException {
				continue
			}
			current := t.ExceptionApproval(id, approval.Requirement)
			r := report{Approval: approval, Applicable: current != nil && current.ID == approval.ID}
			if !r.Applicable {
				r.Problem = "approval no longer matches the current contract, report or scoped work; obtain renewed approval"
			}
			records = append(records, r)
		}
		if a.json {
			a.emit(map[string]any{"id": id, "records": records})
		} else {
			if len(records) == 0 {
				a.printf("No recorded verification exceptions.\n")
			}
			for _, r := range records {
				a.printf("%s applicable=%t approved by %s\n  %s\n", r.Approval.ID, r.Applicable, r.Approval.By, r.Approval.Reason)
				if r.Problem != "" {
					a.printf("  %s\n", r.Problem)
				}
			}
		}
		return nil
	}
	approval.Scopes, approval.ConfigScopes = scopes, configs
	approval.Snapshot, err = worktree.Fingerprint(a.store.Root, append(append([]string(nil), scopes...), configs...))
	if err != nil {
		return usageErr("cannot fingerprint approval scopes: %v", err)
	}
	approval.By, approval.At = a.actor, a.now()
	c, err := t.PlanException(id, approval)
	if err != nil {
		return err
	}
	return writeApproval(a, t, c, "exception-approve")
}

func cmdApproveOptOut(a *app, args []string) error {
	fs := flag.NewFlagSet("dod approve", flag.ContinueOnError)
	item := fs.String("item", "", "exact requested opt-out item on this issue")
	reason := fs.String("reason", "", "why the human approves the opt-out")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("dod approve needs <id> --item <item> --reason <text> --by human:<name>")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	c, err := t.PlanApproveOptOut(id, *item, domain.Input{Actor: a.actor, Now: a.now(), Reason: *reason})
	if err != nil {
		return err
	}
	return writeApproval(a, t, c, "dod-approve")
}
