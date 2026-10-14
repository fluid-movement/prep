package cli

import (
	"flag"
	"slices"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/worktree"
)

const reviewCoverageLimit = "fingerprints cover only the declared work/configuration scopes and selected knowledge-review paths; unlisted work may be absent. Human actor names are local attribution, not authentication."

type resolutionReviewView struct {
	ID             string                    `json:"id"`
	Status         string                    `json:"status"`
	ResolutionID   string                    `json:"resolution_id,omitempty"`
	Generation     uint64                    `json:"generation,omitempty"`
	Reviews        []domain.ResolutionReview `json:"reviews"`
	WorkScopes     []string                  `json:"work_scopes"`
	ConfigScopes   []string                  `json:"config_scopes"`
	CoverageLimits string                    `json:"coverage_limits"`
}

func cmdReview(a *app, args []string) error {
	if len(args) == 0 {
		return usageErr("review needs show or approve; run prep help")
	}
	switch args[0] {
	case "show":
		return cmdReviewShow(a, args[1:])
	case "approve":
		return cmdReviewApprove(a, args[1:])
	default:
		return usageErr("review needs show or approve; run prep help")
	}
}

func reviewView(t *domain.Tree, id string) resolutionReviewView {
	i := t.Issues[id]
	v := resolutionReviewView{ID: id, Status: t.ResolutionReviewStatus(id), Reviews: slices.Clone(i.ResolutionReviews), CoverageLimits: reviewCoverageLimit}
	if i.Resolution != nil && i.Resolution.Outcome == domain.OutcomeDone {
		v.ResolutionID = i.Resolution.ContentID()
		v.Generation = i.Resolution.Generation
		v.WorkScopes = domain.ReviewWorkScopes(i.Resolution)
		v.ConfigScopes = domain.ReviewConfigScopes(i.Resolution)
	}
	if v.Reviews == nil {
		v.Reviews = []domain.ResolutionReview{}
	}
	if v.WorkScopes == nil {
		v.WorkScopes = []string{}
	}
	if v.ConfigScopes == nil {
		v.ConfigScopes = []string{}
	}
	return v
}

func cmdReviewShow(a *app, args []string) error {
	fs := flag.NewFlagSet("review show", flag.ContinueOnError)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("review show", pos)
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
	v := reviewView(t, id)
	a.note(issueEvent(t, "review-show", id, "read resolution review status"))
	if a.json {
		a.emit(v)
		return nil
	}
	a.printf("%s resolution review: %s\n", id, v.Status)
	if v.ResolutionID != "" {
		a.printf("resolution: %s (generation %d)\n", v.ResolutionID, v.Generation)
	}
	for _, r := range v.Reviews {
		a.printf("- %s: approved by %s at %s — %s\n", r.ID, r.By, r.At.Format("2006-01-02T15:04:05Z07:00"), r.Reason)
		a.printf("  work scopes: %s\n", strings.Join(r.Scopes, ", "))
		if len(r.ConfigScopes) > 0 {
			a.printf("  config scopes: %s\n", strings.Join(r.ConfigScopes, ", "))
		}
	}
	a.printf("Coverage: %s\n", v.CoverageLimits)
	return nil
}

func cmdReviewApprove(a *app, args []string) error {
	fs := flag.NewFlagSet("review approve", flag.ContinueOnError)
	reason := fs.String("reason", "", "why this resolution is approved")
	var explicit multi
	fs.Var(&explicit, "scope", "work path covered by this review (repeatable; needed if resolution has no scopes)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	ref, err := one("review approve", pos)
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
	issue := t.Issues[id]
	if issue.Resolution == nil || issue.Resolution.Outcome != domain.OutcomeDone {
		return &domain.Error{Code: domain.ErrGate, Message: "review approval requires a done resolution"}
	}
	for _, p := range explicit {
		if strings.TrimSpace(p) == "" {
			return usageErr("--scope values must not be empty")
		}
	}
	workScopes := domain.ReviewWorkScopes(issue.Resolution)
	workScopes = append(workScopes, explicit...)
	slices.Sort(workScopes)
	workScopes = slices.Compact(workScopes)
	configScopes := domain.ReviewConfigScopes(issue.Resolution)
	if len(workScopes) == 0 {
		return usageErr("this resolution has no recorded work scope; pass --scope <path> (repeatable) to identify what the human reviewed")
	}
	paths := append(slices.Clone(workScopes), configScopes...)
	fingerprint, err := worktree.Fingerprint(a.store.Root, paths)
	if err != nil {
		return usageErr("cannot fingerprint review scopes: %v", err)
	}
	in := domain.Input{Actor: a.actor, Now: a.now(), Reason: *reason, Scopes: workScopes, ConfigScopes: configScopes, Snapshot: fingerprint}
	c, err := t.Plan(id, domain.OpResolutionReview, in)
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
	a.note(issueEvent(t, string(domain.OpResolutionReview), id, opVerbs[domain.OpResolutionReview]))
	a.reportWrite(writeResult{OK: true, Op: "review-approve", ID: id, State: t.Apply(c).State(id), Files: files})
	return nil
}
