package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
)

func cmdPolicy(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" && args[0] != "set" && args[0] != "adopt" {
		return usageErr("policy needs show [id], adopt <id>, or set --verification <method> --documentation <method>")
	}
	fs := flag.NewFlagSet("policy "+args[0], flag.ContinueOnError)
	var p domain.CompletionPolicy
	if args[0] == "set" {
		fs.StringVar(&p.Verification, "verification", "", "how work is verified (prep never executes checks)")
		fs.StringVar(&p.Documentation, "documentation", "", "how documentation impact is reviewed")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) > 1 || args[0] == "set" && len(pos) != 0 || args[0] == "adopt" && len(pos) != 1 {
		return usageErr("policy accepts show [id], adopt <id>, or set flags")
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	if len(pos) == 1 {
		id, err := t.Resolve(pos[0])
		if err != nil {
			return err
		}
		if args[0] == "adopt" {
			c, err := t.PlanAdoptPolicy(id, domain.Input{Actor: a.actor, Now: a.now()})
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
			a.reportWrite(writeResult{OK: true, Op: "policy-adopt", ID: id, Files: files})
			return nil
		}
		var bound *domain.CompletionContract
		if r := t.Issues[id].Ready; r != nil {
			bound = r.Contract
		}
		current := t.Contract(id)
		if a.json {
			a.emit(map[string]any{"id": id, "bound": bound, "current": current, "changed": t.PolicyChanged(id), "legacy": bound == nil})
		} else {
			if bound == nil {
				a.printf("No bound completion contract (legacy or not ready).\n")
			} else {
				a.printf("Bound contract: %s\n", bound.ID)
			}
			if current == nil {
				a.printf("No configured project completion policy.\n")
			} else {
				a.printf("Current contract: %s\nChanged: %t\n", current.ID, t.PolicyChanged(id))
				for _, r := range current.Requirements {
					a.printf("  %s [%s] %s\n", r.ID, r.Kind, r.Text)
				}
			}
		}
		return nil
	}
	if args[0] == "show" {
		if a.json {
			a.emit(struct {
				Configured bool                     `json:"configured"`
				Policy     *domain.CompletionPolicy `json:"policy"`
				DoD        []string                 `json:"dod"`
			}{t.Project.Policy.Valid(), t.Project.Policy, t.Project.DoD})
		} else if t.Project.Policy == nil {
			a.printf("No explicit completion policy. Set one with prep policy set --verification <method> --documentation <method> before marking work ready.\n")
		} else {
			p := t.Project.Policy
			a.printf("Verification: %s\nDocumentation: %s\nRecorded by: %s at %s\n", p.Verification, p.Documentation, p.By, p.At.Format("2006-01-02T15:04:05Z"))
		}
		return nil
	}
	p.By, p.At = a.actor, a.now()
	c, err := t.PlanPolicy(p)
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
	a.reportWrite(writeResult{OK: true, Op: "policy", Files: files})
	return nil
}
