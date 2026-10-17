package cli

import (
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

func validate(t *domain.Tree) []domain.Diagnostic {
	return domain.Validate(t)
}

// resolveDeps resolves local IDs or unique suffixes.
func (a *app) resolveDeps(t *domain.Tree, deps []string) ([]string, error) {
	var ids []string
	for _, d := range deps {
		if strings.Contains(d, ":") {
			return nil, usageErr("--depends-on accepts only issue IDs or suffixes in this project")
		}
		id, err := t.Resolve(d)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
