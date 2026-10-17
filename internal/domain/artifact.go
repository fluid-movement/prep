package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// ArtifactReference identifies external authority without copying or trusting it.
// Location is an opaque URI or path; prep never fetches it or executes it.
type ArtifactReference struct {
	Source   string `json:"source" yaml:"source"`
	Location string `json:"location" yaml:"location"`
	Title    string `json:"title,omitempty" yaml:"title,omitempty"`
}

func ValidateArtifactReferences(refs []ArtifactReference) error {
	seen := map[[2]string]bool{}
	for _, ref := range refs {
		if strings.TrimSpace(ref.Source) == "" || strings.TrimSpace(ref.Location) == "" {
			return fmt.Errorf("source identity and location are required")
		}
		for _, value := range []string{ref.Source, ref.Location, ref.Title} {
			if strings.ContainsFunc(value, unicode.IsControl) {
				return fmt.Errorf("source, location and title must be single-line text without control characters")
			}
		}
		key := [2]string{ref.Source, ref.Location}
		if seen[key] {
			return fmt.Errorf("duplicate reference for %s at %s", ref.Source, ref.Location)
		}
		seen[key] = true
	}
	return nil
}

type ArtifactInput struct {
	Reference ArtifactReference
	Remove    bool
	Actor     string
	Now       time.Time
}

// PlanArtifact changes reference metadata only; lifecycle gates still apply.
func (t *Tree) PlanArtifact(id string, in ArtifactInput) (*Change, error) {
	i := t.Issues[id]
	if i == nil {
		return nil, &Error{Code: ErrNotFound, Message: "issue not found: " + id}
	}
	r := in.Reference
	r.Source, r.Location, r.Title = strings.TrimSpace(r.Source), strings.TrimSpace(r.Location), strings.TrimSpace(r.Title)
	if err := ValidateArtifactReferences([]ArtifactReference{r}); err != nil {
		return nil, &Error{Code: ErrUsage, Message: err.Error(), Next: "prep artifact add " + id + " --source <identity> --location <URI-or-path>"}
	}
	refs := make([]ArtifactReference, 0, len(i.Artifacts)+1)
	found := false
	for _, existing := range i.Artifacts {
		if existing.Source == r.Source && existing.Location == r.Location {
			found = true
			if in.Remove {
				continue
			}
		}
		refs = append(refs, existing)
	}
	if in.Remove && !found {
		return nil, &Error{Code: ErrNotFound, Message: "artifact reference not found", Next: "prep show " + id}
	}
	if !in.Remove && found {
		return nil, &Error{Code: ErrInvalid, Message: "artifact reference already exists; remove it before replacing it", Next: "prep show " + id}
	}
	op := "add"
	if in.Remove {
		op = "remove"
	} else {
		refs = append(refs, r)
	}
	return &Change{Op: Op("artifact " + op), IssueID: id, Artifacts: &refs,
		History: fmt.Sprintf("- %s artifact reference %s by %s: %s at %s", in.Now.UTC().Truncate(time.Second).Format(time.RFC3339), op, in.Actor, r.Source, r.Location)}, nil
}
