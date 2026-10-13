package cli

import (
	"flag"

	"github.com/fluid-movement/prep/internal/domain"
)

func cmdPolicy(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" && args[0] != "set" {
		return usageErr("policy needs show or set --verification <method> --documentation <method>")
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
	if len(pos) != 0 {
		return usageErr("unexpected policy arguments")
	}
	t, err := a.load(false)
	if err != nil {
		return err
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
