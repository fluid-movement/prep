package cli

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/gitx"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/projects"
	"github.com/fluid-movement/prep/internal/tui"
	"github.com/fluid-movement/prep/internal/worktree"
)

const tuiReviewDiffLimit = 24 * 1024

func loadTUIReview(root, ref string, extraScopes []string) (tui.ReviewSnapshot, error) {
	t, err := loadTUIReviewTree(root)
	if err != nil {
		return tui.ReviewSnapshot{}, err
	}
	id, err := t.Resolve(ref)
	if err != nil {
		return tui.ReviewSnapshot{}, err
	}
	i := t.Issues[id]
	if i.Resolution == nil || i.Resolution.Outcome != domain.OutcomeDone {
		return tui.ReviewSnapshot{}, fmt.Errorf("issue %s has no done resolution to review", id)
	}
	workScopes := mdstore.ResolutionReviewWorkScopes(i.Resolution)
	for _, scope := range extraScopes {
		if strings.TrimSpace(scope) == "" {
			return tui.ReviewSnapshot{}, fmt.Errorf("explicit review scopes must not be blank")
		}
		workScopes = append(workScopes, scope)
	}
	slices.Sort(workScopes)
	workScopes = slices.Compact(workScopes)
	configScopes := domain.ReviewConfigScopes(i.Resolution)

	// Preserve the recorded resolution exactly. Git-derived evidence below is
	// presentation metadata and must not change the approval identity.
	issue := *i
	resolution := *i.Resolution
	issue.Resolution = &resolution
	contractID := ""
	if contract := t.Contract(id); contract != nil {
		contractID = contract.ID
	}
	snapshot := tui.ReviewSnapshot{
		Issue:             &issue,
		Status:            t.ResolutionReviewStatus(id),
		ResolutionID:      i.Resolution.ContentID(),
		CurrentContractID: contractID,
		WorkScopes:        workScopes,
		ConfigScopes:      slices.Clone(configScopes),
		KnowledgeCurrent:  selectedKnowledgeCurrent(t, id, i.Resolution),
	}
	if len(workScopes) == 0 {
		snapshot.ChangesNote = "This legacy resolution has no reported work scope. Select at least one explicit scope and inspect the refreshed bundle before approval."
		return snapshot, nil
	}
	paths := append(slices.Clone(workScopes), configScopes...)
	before, err := worktree.Fingerprint(root, paths)
	if err != nil {
		return tui.ReviewSnapshot{}, fmt.Errorf("cannot fingerprint review scopes: %w", err)
	}
	snapshot.Fingerprint = before
	snapshot.Evidence = reviewGitEvidence(root, id, i.Resolution)
	snapshot.Changes, snapshot.ChangesNote = reviewGitChanges(root, snapshot.Evidence, paths)
	after, err := worktree.Fingerprint(root, paths)
	if err != nil {
		return tui.ReviewSnapshot{}, fmt.Errorf("cannot fingerprint review scopes: %w", err)
	}
	if before != after {
		return tui.ReviewSnapshot{}, fmt.Errorf("review scopes changed while loading the review; reload before approval")
	}
	return snapshot, nil
}

func loadTUIReviewTree(root string) (*domain.Tree, error) {
	t, err := domain.Load(mdstore.Open(root), &okf.Store{Root: root}, false)
	if err != nil {
		return nil, err
	}
	projects.ResolveExternal(t)
	return t, nil
}

func reviewGitEvidence(root, id string, r *domain.Resolution) string {
	if r == nil {
		return ""
	}
	if r.Evidence != "" && gitx.CommitExists(root, r.Evidence) {
		return r.Evidence
	}
	return gitx.LastCommit(root, mdstore.Dir+"/issues/"+id+"/resolution.md")
}

func reviewGitChanges(root, evidence string, paths []string) (string, string) {
	if !gitx.IsRepo(root) {
		return "", "Git diff unavailable: this project is not in a Git work tree. The scoped content fingerprint is still available."
	}
	var sections []string
	if evidence == "" {
		sections = append(sections, "No Git commit evidence is available for this resolution; scoped file differences are listed where Git can identify them.")
	} else {
		args := []string{"show", "--format=commit %H%n%s", "--no-ext-diff", "--no-color", "--find-renames", "--unified=2", evidence, "--"}
		args = append(args, paths...)
		if diff, err := gitx.Run(root, args...); err != nil {
			sections = append(sections, "Commit diff unavailable: "+err.Error())
		} else if diff == "" {
			sections = append(sections, "The evidence commit has no diff in the declared scopes.")
		} else {
			sections = append(sections, diff)
		}
	}
	if head, err := gitx.Head(root); err == nil {
		currentDiffArgs := []string{"diff", "--no-ext-diff", "--no-color", "--unified=2", head, "--"}
		currentDiffArgs = append(currentDiffArgs, paths...)
		if diff, err := gitx.Run(root, currentDiffArgs...); err == nil && diff != "" {
			sections = append(sections, "Current tracked patch:\n"+diff)
		} else if err != nil {
			sections = append(sections, "Current tracked patch unavailable: "+err.Error())
		}
		if changed, err := gitx.ChangedSince(root, head, paths); err == nil {
			if len(changed) == 0 {
				sections = append(sections, "Current Git working tree has no tracked differences in the declared scopes.")
			} else {
				sections = append(sections, "Current tracked differences: "+strings.Join(changed, ", "))
			}
		} else {
			sections = append(sections, "Current tracked differences unavailable: "+err.Error())
		}
		if status, err := gitx.Run(root, append([]string{"status", "--porcelain", "--untracked-files=all", "--"}, paths...)...); err == nil && status != "" {
			sections = append(sections, "Current staged, unstaged, or untracked scoped files:\n"+status)
		}
	} else {
		sections = append(sections, "Current Git differences unavailable: no HEAD commit.")
	}
	text := strings.Join(sections, "\n\n")
	if len(text) > tuiReviewDiffLimit {
		text = text[:tuiReviewDiffLimit] + "\n… diff truncated for display"
	}
	note := ""
	if evidence == "" {
		note = "No commit evidence; approval is still allowed because the scoped content snapshot is captured directly."
	}
	return text, note
}

func approveTUIReview(root string, write func(func(*domain.Tree) (*domain.Change, error)) (string, error), actor string, req tui.ReviewApprovalRequest) (string, error) {
	s := req.Snapshot
	if s.Issue == nil || s.Issue.Resolution == nil || s.Issue.ID == "" {
		return "", fmt.Errorf("review snapshot has no completed issue")
	}
	if s.Issue.Resolution.ContentID() != s.ResolutionID {
		return "", fmt.Errorf("displayed resolution does not match its captured identity; reload before approval")
	}
	if !strings.HasPrefix(actor, "human:") || strings.TrimSpace(strings.TrimPrefix(actor, "human:")) == "" {
		return "", fmt.Errorf("resolution review requires a human:<name> actor")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return "", fmt.Errorf("resolution review requires a reason")
	}
	if len(s.WorkScopes) == 0 || s.Fingerprint == "" {
		return "", fmt.Errorf("review has no captured work fingerprint; select explicit legacy scopes and reload")
	}
	return write(func(t *domain.Tree) (*domain.Change, error) {
		i := t.Issues[s.Issue.ID]
		if i == nil || i.Resolution == nil || i.Resolution.Outcome != domain.OutcomeDone {
			return nil, fmt.Errorf("issue no longer has a done resolution; reload the review")
		}
		if i.Resolution.ContentID() != s.ResolutionID {
			return nil, fmt.Errorf("resolution changed since this review was displayed; reload before approval")
		}
		contractID := ""
		if contract := t.Contract(i.ID); contract != nil {
			contractID = contract.ID
		}
		if contractID != s.CurrentContractID {
			return nil, fmt.Errorf("current completion contract changed since this review was displayed; reload before approval")
		}
		if issueReviewContentChanged(s.Issue, i) {
			return nil, fmt.Errorf("displayed issue requirements or acceptance content changed; reload before approval")
		}
		if status := t.ResolutionReviewStatus(i.ID); status != s.Status {
			return nil, fmt.Errorf("resolution review status changed from %s to %s since it was displayed; reload before approval", s.Status, status)
		}
		knowledgeCurrent := selectedKnowledgeCurrent(t, i.ID, i.Resolution)
		if !sameBoolMap(s.KnowledgeCurrent, knowledgeCurrent) {
			return nil, fmt.Errorf("selected knowledge-review applicability changed since it was displayed; reload before approval")
		}
		workScopes := mdstore.ResolutionReviewWorkScopes(i.Resolution)
		workScopes = append(workScopes, s.WorkScopes...)
		slices.Sort(workScopes)
		workScopes = slices.Compact(workScopes)
		configScopes := domain.ReviewConfigScopes(i.Resolution)
		if !sameStrings(s.ConfigScopes, configScopes) {
			return nil, fmt.Errorf("review configuration coverage changed; reload before approval")
		}
		// The displayed work scope selection may include explicit legacy scopes,
		// but must contain every scope reported by the resolution.
		recorded := mdstore.ResolutionReviewWorkScopes(i.Resolution)
		for _, scope := range recorded {
			if !slices.Contains(s.WorkScopes, scope) {
				return nil, fmt.Errorf("review omitted recorded scope %q; reload before approval", scope)
			}
		}
		if !sameStrings(workScopes, s.WorkScopes) {
			return nil, fmt.Errorf("review work scope coverage changed; reload before approval")
		}
		paths := append(slices.Clone(s.WorkScopes), configScopes...)
		current, err := worktree.Fingerprint(root, paths)
		if err != nil {
			return nil, fmt.Errorf("cannot fingerprint review scopes: %w", err)
		}
		if current != s.Fingerprint {
			return nil, fmt.Errorf("review scopes changed since they were displayed; reload before approval")
		}
		return t.Plan(i.ID, domain.OpResolutionReview, domain.Input{
			Actor: actor, Now: time.Now().UTC(), Reason: req.Reason,
			Scopes: slices.Clone(s.WorkScopes), ConfigScopes: slices.Clone(configScopes), Snapshot: s.Fingerprint,
		})
	})
}

func selectedKnowledgeCurrent(t *domain.Tree, id string, r *domain.Resolution) map[string]bool {
	current := make(map[string]bool, len(r.KnowledgeReviews))
	for n := range r.KnowledgeReviews {
		report := &r.KnowledgeReviews[n]
		current[report.ID] = t.KnowledgeReviewApplicable(id, report)
	}
	return current
}

func sameBoolMap(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if other, ok := b[key]; !ok || other != value {
			return false
		}
	}
	return true
}

func issueReviewContentChanged(displayed, current *domain.Issue) bool {
	if displayed == nil || current == nil {
		return displayed != current
	}
	if displayed.Body != current.Body || displayed.Prose != current.Prose || displayed.OpenQuestions != current.OpenQuestions || !slices.Equal(displayed.Criteria, current.Criteria) {
		return true
	}
	if displayed.Ready == nil || current.Ready == nil {
		return displayed.Ready != current.Ready
	}
	if displayed.Ready.Baseline != current.Ready.Baseline {
		return true
	}
	return !sameBaseline(baselineFor(displayed, displayed.Ready.Baseline), baselineFor(current, current.Ready.Baseline))
}

func baselineFor(i *domain.Issue, name string) *domain.Baseline {
	for n := range i.Baselines {
		if i.Baselines[n].Name == name {
			return &i.Baselines[n]
		}
	}
	return nil
}

func sameBaseline(a, b *domain.Baseline) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func sameStrings(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
