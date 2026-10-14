package domain

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	OpKnowledgeReview   Op = "knowledge-review"
	GateKnowledge          = "G_KNOWLEDGE"
	KnowledgeUpdated       = "updated"
	KnowledgeUnchanged     = "unchanged"
	KnowledgeNoImpact      = "no-impact"
	KnowledgeProvenance    = "reported"
)

// KnowledgeReview is an agent-reported assessment bound to one completion
// contract, scoped work snapshot, and (when named) canonical entry revision.
// It records an assessment; it does not prove semantic completeness.
type KnowledgeReview struct {
	ID            string    `json:"id" yaml:"id"`
	Previous      string    `json:"previous,omitempty" yaml:"previous,omitempty"`
	Contract      string    `json:"contract" yaml:"contract"`
	Kind          string    `json:"kind" yaml:"kind"`
	Entry         string    `json:"entry,omitempty" yaml:"entry,omitempty"`
	EntryRevision string    `json:"entry_revision,omitempty" yaml:"entry_revision,omitempty"`
	Observation   string    `json:"observation" yaml:"observation"`
	Provenance    string    `json:"provenance" yaml:"provenance"`
	By            string    `json:"by" yaml:"by"`
	At            time.Time `json:"at" yaml:"at"`
	Scopes        []string  `json:"scopes" yaml:"scopes"`
	ConfigScopes  []string  `json:"config_scopes,omitempty" yaml:"config_scopes,omitempty"`
	Snapshot      string    `json:"snapshot" yaml:"snapshot"`
}

func (r KnowledgeReview) ContentID() string { r.ID = ""; return contentID(r) }

// KnowledgeEntryRevision excludes confirmation metadata and derived indexes:
// neither changes what an agent reviewed about an entry.
func KnowledgeEntryRevision(e *Entry) string {
	if e == nil {
		return ""
	}
	scope := append([]string(nil), e.Scope...)
	sort.Strings(scope)
	return contentID(struct {
		Path, Type, Title, Description, Status string
		Scope                                  []string
		Body                                   string
	}{e.Path, e.Type, e.Title, e.Description, e.Status, scope, strings.TrimSpace(e.Body)})
}

func (r KnowledgeReview) Validate() error {
	snapshot, err := hex.DecodeString(r.Snapshot)
	if r.Contract == "" || (r.Kind != KnowledgeUpdated && r.Kind != KnowledgeUnchanged && r.Kind != KnowledgeNoImpact) ||
		strings.TrimSpace(r.By) == "" || r.At.IsZero() || r.Provenance != KnowledgeProvenance ||
		len(r.Scopes) == 0 || err != nil || len(snapshot) != 32 {
		return fmt.Errorf("knowledge review needs a contract, reported kind, actor, time, work scopes and fingerprint")
	}
	for _, scope := range append(append([]string(nil), r.Scopes...), r.ConfigScopes...) {
		if strings.TrimSpace(scope) == "" {
			return fmt.Errorf("knowledge review scope paths must not be blank")
		}
	}
	switch r.Kind {
	case KnowledgeUpdated, KnowledgeUnchanged:
		revision, err := hex.DecodeString(r.EntryRevision)
		if NormalizeEntryPath(r.Entry) != r.Entry || r.Entry == "" || err != nil || len(revision) != 32 {
			return fmt.Errorf("entry review needs a normalized entry path and content revision")
		}
		if strings.TrimSpace(r.Observation) == "" {
			return fmt.Errorf("entry review needs an observation")
		}
	case KnowledgeNoImpact:
		if r.Entry != "" || r.EntryRevision != "" || strings.TrimSpace(r.Observation) == "" {
			return fmt.Errorf("no-impact review needs a reason and cannot name an entry")
		}
	}
	return nil
}

func LatestKnowledgeReview(i *Issue, entry string) *KnowledgeReview {
	for n := len(i.KnowledgeReviews) - 1; n >= 0; n-- {
		r := &i.KnowledgeReviews[n]
		if r.Entry == entry {
			return r
		}
	}
	return nil
}

func latestNoImpactReview(i *Issue) *KnowledgeReview {
	for n := len(i.KnowledgeReviews) - 1; n >= 0; n-- {
		r := &i.KnowledgeReviews[n]
		if r.Kind == KnowledgeNoImpact {
			return r
		}
	}
	return nil
}

// KnowledgeReviewApplicable checks the adapter's work-scope fingerprint and
// the current entry content, while the domain owns contract and record rules.
func (t *Tree) KnowledgeReviewApplicable(id string, r *KnowledgeReview) bool {
	return t.KnowledgeReviewProblem(id, r) == ""
}

func (t *Tree) KnowledgeReviewProblem(id string, r *KnowledgeReview) string {
	i := t.Issues[id]
	if i == nil || r == nil || r.Validate() != nil || r.ID != r.ContentID() {
		return "review record is invalid"
	}
	if !i.KnowledgeReviewCurrent[r.ID] {
		return "scoped work or configuration changed"
	}
	if i.Ready == nil || i.Ready.Contract == nil || r.Contract != i.Ready.Contract.ID || t.PolicyReviewRequired(id) {
		return "completion contract changed; reconcile it before reusing the report"
	}
	if r.Entry != "" && KnowledgeEntryRevision(t.Knowledge[r.Entry]) != r.EntryRevision {
		return "knowledge entry content changed"
	}
	return ""
}

// KnowledgeReviewsForCompletion selects exactly the reports used by the
// completion decision. A docs reference is only a selector, never evidence.
func (t *Tree) KnowledgeReviewsForCompletion(id string, docs []string, noImpact string) ([]KnowledgeReview, []string) {
	i := t.Issues[id]
	if i == nil {
		return nil, []string{"issue not found"}
	}
	var selected []KnowledgeReview
	var problems []string
	if len(docs) > 0 {
		for _, doc := range docs {
			entry := NormalizeEntryPath(doc)
			r := LatestKnowledgeReview(i, entry)
			if r == nil || (r.Kind != KnowledgeUpdated && r.Kind != KnowledgeUnchanged) || !t.KnowledgeReviewApplicable(id, r) {
				problems = append(problems, fmt.Sprintf("knowledge entry %s needs a current updated or unchanged review; use prep knowledge review record %s --entry %s --kind updated|unchanged --observation <what you checked> --scope <work-path>", entry, id, entry))
				continue
			}
			selected = append(selected, *r)
		}
		return selected, problems
	}
	r := latestNoImpactReview(i)
	if r == nil || !t.KnowledgeReviewApplicable(id, r) || strings.TrimSpace(noImpact) == "" || r.Observation != strings.TrimSpace(noImpact) {
		problems = append(problems, fmt.Sprintf("no-impact needs a current justified knowledge review for this work; use prep knowledge review record %s --no-impact --reason <reason> --scope <work-path>", id))
		return nil, problems
	}
	return []KnowledgeReview{*r}, nil
}

func (t *Tree) PlanKnowledgeReview(id string, r KnowledgeReview) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found"}
	}
	if t.State(id).Terminal() || i.Ready == nil || i.Ready.Contract == nil || !t.ReadyValid(i) || t.PolicyReviewRequired(id) || t.Stale(id) {
		return nil, &Error{Code: ErrGate, Message: "knowledge review requires an unresolved issue with a valid reconciled completion contract"}
	}
	r.Contract = i.Ready.Contract.ID
	r.Provenance = KnowledgeProvenance
	r.By = strings.TrimSpace(r.By)
	r.At = r.At.UTC().Truncate(time.Second)
	r.Observation = strings.TrimSpace(r.Observation)
	r.Entry = NormalizeEntryPath(r.Entry)
	r.Scopes = append([]string(nil), r.Scopes...)
	r.ConfigScopes = append([]string(nil), r.ConfigScopes...)
	sort.Strings(r.Scopes)
	sort.Strings(r.ConfigScopes)
	if r.Kind == KnowledgeUpdated || r.Kind == KnowledgeUnchanged {
		e := t.Knowledge[r.Entry]
		if e == nil {
			return nil, &Error{Code: ErrGate, Message: "knowledge entry does not exist"}
		}
		r.EntryRevision = KnowledgeEntryRevision(e)
	} else if r.Kind == KnowledgeNoImpact {
		r.Entry, r.EntryRevision = "", ""
	}
	if err := r.Validate(); err != nil {
		return nil, &Error{Code: ErrGate, Message: err.Error(), Unmet: []Unmet{{GateKnowledge, err.Error()}}}
	}
	r.Previous = ""
	if n := len(i.KnowledgeReviews); n > 0 {
		r.Previous = i.KnowledgeReviews[n-1].ID
	}
	r.ID = r.ContentID()
	records := append(append([]KnowledgeReview(nil), i.KnowledgeReviews...), r)
	return &Change{Op: OpKnowledgeReview, IssueID: id, KnowledgeReviews: &records, History: fmt.Sprintf("- %s %s: reported knowledge %s review %s", r.At.Format(time.RFC3339), r.By, r.Kind, r.ID)}, nil
}
