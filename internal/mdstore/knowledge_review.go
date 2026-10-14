package mdstore

import (
	"fmt"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

type knowledgeReviewsFM struct {
	Reviews []domain.KnowledgeReview `yaml:"reviews"`
}

func parseKnowledgeReviews(raw string) ([]domain.KnowledgeReview, error) {
	fm, body, ok, err := splitFrontmatter(raw)
	if err != nil || !ok || strings.TrimSpace(body) != "" {
		return nil, fmt.Errorf("knowledge-reviews.md requires reviews frontmatter and no body")
	}
	var f knowledgeReviewsFM
	if err := decodeStrict(fm, &f); err != nil {
		return nil, err
	}
	previous := ""
	for _, r := range f.Reviews {
		if err := r.Validate(); err != nil {
			return nil, err
		}
		if r.Previous != previous || r.ID != r.ContentID() {
			return nil, fmt.Errorf("invalid knowledge review content hash or history chain")
		}
		previous = r.ID
	}
	return f.Reviews, nil
}

func renderKnowledgeReviews(reviews []domain.KnowledgeReview) string {
	return withFrontmatter(encodeYAML(knowledgeReviewsFM{Reviews: reviews}), "")
}
