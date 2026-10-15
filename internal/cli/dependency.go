package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

func cmdDependency(a *app, args []string) error {
	if len(args) == 0 || args[0] != "show" && args[0] != "override" {
		return usageErr("dependency needs show <id> or override <id> <project:id> --reason <text> --by human:<name>")
	}
	op := args[0]
	fs := flag.NewFlagSet("dependency "+op, flag.ContinueOnError)
	var reason, by string
	if op == "override" {
		fs.StringVar(&reason, "reason", "", "why a human accepts proceeding with this unknown dependency")
		fs.StringVar(&by, "by", "", "human:<name> who approved the override")
	}
	pos, err := parse(fs, args[1:])
	if err != nil {
		return err
	}
	if op == "show" && len(pos) != 1 || op == "override" && len(pos) != 2 {
		return usageErr("dependency %s has invalid arguments", op)
	}
	t, err := a.load(false)
	if err != nil {
		return err
	}
	id, err := t.Resolve(pos[0])
	if err != nil {
		return err
	}
	i := t.Issues[id]
	if op == "show" {
		type retiredOverride struct {
			Override domain.DependencyOverride `json:"override"`
			Status   string                    `json:"status"`
		}
		var retired []retiredOverride
		seen := map[string]bool{}
		for n := len(i.DependencyOverrides) - 1; n >= 0; n-- {
			o := i.DependencyOverrides[n]
			if seen[o.Ref] {
				continue
			}
			seen[o.Ref] = true
			declared := false
			for _, ref := range i.DependsOn {
				if ref == o.Ref {
					declared = true
					break
				}
			}
			if !declared {
				retired = append(retired, retiredOverride{o, t.DependencyOverrideStatus(id, o.Ref)})
			}
		}
		type item struct {
			Ref         string        `json:"ref"`
			Observation domain.XIssue `json:"observation"`
		}
		items := []item{}
		for _, ref := range i.DependsOn {
			if _, _, cross := domain.SplitRef(ref); cross {
				items = append(items, item{ref, t.ExternalDependency(id, ref)})
			}
		}
		if a.json {
			data := map[string]any{"id": id, "dependencies": items}
			if len(retired) > 0 {
				data["retired_overrides"] = retired
			}
			a.emit(data)
			return nil
		}
		if len(items) == 0 {
			a.printf("No external dependencies.\n")
		}
		for _, it := range items {
			x := it.Observation
			reviewStatus := x.ReviewStatus
			if x.RequiresApproval && reviewStatus == domain.ReviewNone && x.Status != domain.XFound {
				reviewStatus = "unknown"
			}
			a.printf("%s: %s", it.Ref, x.Outcome)
			if x.RequiresApproval {
				a.printf("; approval %s", reviewStatus)
			}
			if x.OverrideStatus != "none" {
				a.printf("; override %s by %s: %s", x.OverrideStatus, x.OverrideBy, x.OverrideReason)
			}
			if x.Uncertainty != "" {
				a.printf("; unknown: %s", x.Uncertainty)
			}
			a.printf("; source %s (%s), checkout %s, revision %s\n", x.Project, x.SourceProjectID, x.Root, x.Revision)
			if x.ProvenanceWarning != "" {
				a.printf("  provenance warning: %s\n", x.ProvenanceWarning)
			}
		}
		for _, r := range retired {
			a.printf("retired override %s: %s by %s: %s (dependency removed or renamed)\n", r.Override.Ref, r.Status, r.Override.By, r.Override.Reason)
		}
		return nil
	}
	if by == "" && a.named {
		by = a.actor
	}
	if !strings.HasPrefix(by, "human:") || strings.TrimSpace(strings.TrimPrefix(by, "human:")) == "" {
		return usageErr("dependency override requires --by human:<name>")
	}
	ref := pos[1]
	// Resolve an optional suffix without replacing the rich fresh observation
	// from a.load with ResolveDep's narrower link-resolution result.
	resolveTree := *t
	resolveTree.External = map[string]domain.XIssue{}
	resolved, _, err := a.resolveDeps(&resolveTree, []string{ref})
	if err != nil {
		return err
	}
	if len(resolved) != 1 {
		return usageErr("dependency override requires one external dependency")
	}
	ref = resolved[0]
	c, err := t.PlanDependencyOverride(id, domain.DependencyOverrideInput{Ref: ref, Reason: reason, Actor: by, Now: a.now()})
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
	a.reportWrite(writeResult{OK: true, Op: "dependency override", ID: id, Files: files, Next: "prep dependency show " + id})
	a.noteRead(issueEvent(t, "dependency override", id, fmt.Sprintf("record a human override for %s", ref)))
	return nil
}
