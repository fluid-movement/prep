package mdstore

import (
	"path"
	"slices"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// ResolutionReviewWorkScopes adds the stored knowledge entry content to the
// resolution's reported work scopes. Entry revisions are also part of the
// immutable resolution identity; including the files detects later changes.
func ResolutionReviewWorkScopes(r *domain.Resolution) []string {
	scopes := domain.ReviewWorkScopes(r)
	if r != nil {
		for _, review := range r.KnowledgeReviews {
			if review.Entry != "" {
				scopes = append(scopes, path.Join(Dir, "knowledge", strings.TrimPrefix(review.Entry, "/")))
			}
		}
	}
	slices.Sort(scopes)
	return slices.Compact(scopes)
}
