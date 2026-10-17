package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// KnowledgeRecord creates an immutable decision or finding in the knowledge base.
type KnowledgeRecord struct {
	Path        string
	ID          string
	Type        string
	Title       string
	Description string
	Body        string
	Issue       string
	Supersedes  []string
	Scope       []string
	Outcome     bool
	At          time.Time
}

// PlanKnowledgeRecord validates a new record. Existing records are never changed.
func (t *Tree) PlanKnowledgeRecord(in KnowledgeRecord) (*Change, error) {
	if in.Type != "decision" && in.Type != "finding" {
		return nil, &Error{Code: ErrUsage, Message: "record type must be decision or finding"}
	}
	if !ValidID(in.ID) {
		return nil, &Error{Code: ErrUsage, Message: "record ID must be a ULID"}
	}
	in.Title, in.Description = strings.TrimSpace(in.Title), strings.TrimSpace(in.Description)
	if in.Title == "" || in.Description == "" || strings.ContainsAny(in.Description, "\r\n") {
		return nil, &Error{Code: ErrUsage, Message: "--title and a one-line --description are required"}
	}
	if in.Issue != "" && t.Issues[in.Issue] == nil {
		return nil, &Error{Code: ErrNotFound, Message: "no issue " + in.Issue}
	}
	if in.Type == "finding" && (in.Outcome || len(in.Supersedes) > 0) {
		return nil, &Error{Code: ErrUsage, Message: "outcome and supersedes are decision fields"}
	}
	in.At = in.At.UTC().Truncate(time.Second)
	for _, scope := range in.Scope {
		if strings.TrimSpace(scope) == "" {
			return nil, &Error{Code: ErrUsage, Message: "scope must be a non-empty path glob"}
		}
	}
	in.Supersedes = append([]string(nil), in.Supersedes...)
	for n, ref := range in.Supersedes {
		e, err := t.ResolveRecord(ref)
		if err != nil {
			return nil, err
		}
		if e.ID >= in.ID {
			return nil, &Error{Code: ErrUsage, Message: "supersedes must name an older record"}
		}
		in.Supersedes[n] = e.ID
	}
	in.Path = "/" + in.Type + "s/" + in.ID + ".md"
	if t.Knowledge[in.Path] != nil {
		return nil, &Error{Code: ErrConflict, Message: "record already exists: " + in.ID}
	}
	return &Change{Op: Op(in.Type), IssueID: in.Issue, Knowledge: &in}, nil
}

// ResolveRecord accepts a record ID, unique suffix, or bundle path.
func (t *Tree) ResolveRecord(ref string) (*Entry, error) {
	if strings.TrimSpace(ref) == "" {
		return nil, &Error{Code: ErrUsage, Message: "record reference cannot be empty"}
	}
	var matches []*Entry
	for _, e := range t.Knowledge {
		if e.ID != "" && (e.ID == strings.ToUpper(ref) || strings.HasSuffix(e.ID, strings.ToUpper(ref)) || e.Path == NormalizeEntryPath(ref)) {
			matches = append(matches, e)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return nil, &Error{Code: ErrAmbiguous, Message: fmt.Sprintf("record %q is ambiguous", ref)}
	}
	return nil, &Error{Code: ErrNotFound, Message: fmt.Sprintf("no knowledge record %q", ref)}
}

// IssueRecords returns linked records in stable chronological order.
func (t *Tree) IssueRecords(id string) []*Entry {
	var out []*Entry
	for _, e := range t.Knowledge {
		if e.ID != "" && e.Issue == id {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
func (t *Tree) HasRecord(id, typ string, outcome bool) bool {
	for _, e := range t.IssueRecords(id) {
		if e.Type == typ && (!outcome || e.Outcome) {
			return true
		}
	}
	return false
}

// DeriveSuperseded recomputes status solely from newer records.
func (t *Tree) DeriveSuperseded() {
	byID := map[string]*Entry{}
	for _, e := range t.Knowledge {
		e.Superseded = false
		if e.ID != "" {
			byID[e.ID] = e
		}
	}
	for _, e := range t.Knowledge {
		for _, id := range e.Supersedes {
			if old := byID[id]; old != nil && old.ID < e.ID {
				old.Superseded = true
			}
		}
	}
}
