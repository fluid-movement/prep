package domain

import (
	"fmt"
	"sort"
	"strings"
)

// InboxCategory identifies an unresolved item that needs human attention.
type InboxCategory string

const (
	InboxCompletionReview  InboxCategory = "completion-review"
	InboxReadySignoff      InboxCategory = "ready-signoff"
	InboxOpenQuestions     InboxCategory = "open-questions"
	InboxStaleRequirement  InboxCategory = "stale-requirement"
	InboxStalePolicy       InboxCategory = "stale-policy"
	InboxUnknownDependency InboxCategory = "unknown-dependency"
)

// InboxItem is a derived pointer to work needing attention. Reasons and review
// status explain why it was selected; the classifier does not change gates.
type InboxItem struct {
	IssueID      string
	Category     InboxCategory
	Reason       string
	ReviewStatus string
}

// Inbox returns at most one highest-priority attention item for each issue.
// Ordinary backlog and dropped issues stay in the regular issue views.
func (t *Tree) Inbox() []InboxItem {
	if t == nil {
		return nil
	}
	items := make([]InboxItem, 0)
	for id, issue := range t.Issues {
		if issue == nil || t.State(id) == StateDropped {
			continue
		}
		if item, ok := t.inboxItem(id, issue); ok {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(a, b int) bool {
		left, right := items[a], items[b]
		if inboxCategoryRank(left.Category) != inboxCategoryRank(right.Category) {
			return inboxCategoryRank(left.Category) < inboxCategoryRank(right.Category)
		}
		li, ri := t.Issues[left.IssueID], t.Issues[right.IssueID]
		if li.Priority.Rank() != ri.Priority.Rank() {
			return li.Priority.Rank() < ri.Priority.Rank()
		}
		if left.Category == InboxCompletionReview && !li.Resolution.At.Equal(ri.Resolution.At) {
			return li.Resolution.At.After(ri.Resolution.At)
		}
		return left.IssueID < right.IssueID
	})
	return items
}

func inboxCategoryRank(category InboxCategory) int {
	switch category {
	case InboxCompletionReview:
		return 0
	case InboxReadySignoff:
		return 1
	case InboxOpenQuestions:
		return 2
	case InboxStaleRequirement:
		return 3
	case InboxStalePolicy:
		return 4
	case InboxUnknownDependency:
		return 5
	default:
		return 6
	}
}

func (t *Tree) inboxItem(id string, issue *Issue) (InboxItem, bool) {
	state := t.State(id)
	if state == StateDone {
		status := t.ResolutionReviewStatus(id)
		if status == ReviewPending || status == ReviewStale {
			return InboxItem{IssueID: id, Category: InboxCompletionReview,
				Reason: "completed work needs review", ReviewStatus: status}, true
		}
		return InboxItem{}, false
	}
	if state.Terminal() {
		return InboxItem{}, false
	}
	if state == StateDefined && !t.Stale(id) {
		if unmet, _ := t.Gates(id, OpReady, nil); len(unmet) == 0 {
			return InboxItem{IssueID: id, Category: InboxReadySignoff, Reason: "issue is ready for human sign-off"}, true
		}
	}
	if strings.TrimSpace(issue.OpenQuestions) != "" {
		return InboxItem{IssueID: id, Category: InboxOpenQuestions, Reason: "open questions need answers"}, true
	}
	if t.Stale(id) {
		return InboxItem{IssueID: id, Category: InboxStaleRequirement, Reason: "requirement or kind changed since the latest baseline"}, true
	}
	if t.PolicyReviewRequired(id) {
		return InboxItem{IssueID: id, Category: InboxStalePolicy, Reason: "bound completion policy changed and needs review"}, true
	}
	var unknown []string
	for _, dep := range issue.DependsOn {
		if _, _, external := SplitRef(dep); !external {
			continue
		}
		if _, known := t.DepState(dep); !known && t.DependencyOverrideStatus(id, dep) != "current" {
			unknown = append(unknown, dep)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return InboxItem{IssueID: id, Category: InboxUnknownDependency,
			Reason: fmt.Sprintf("external dependency state is unavailable: %s", strings.Join(unknown, ", "))}, true
	}
	return InboxItem{}, false
}
