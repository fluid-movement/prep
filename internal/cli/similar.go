package cli

import (
	"sort"
	"strings"
	"unicode"

	"github.com/fluid-movement/prep/internal/domain"
)

// Hints deliberately neither block creation nor pick an issue: lexical
// overlap is evidence to inspect, never authority to merge different work.
func similarIssues(t *domain.Tree, title string) []summary {
	query := similarWords(title)
	if len(query) == 0 {
		return nil
	}
	type match struct {
		id    string
		score int
	}
	var matches []match
	for _, id := range t.IDs() {
		if t.State(id).Terminal() {
			continue
		}
		i := t.Issues[id]
		name, details := similarWords(i.Title), similarWords(i.Body)
		shared, score := 0, 0
		for word := range query {
			if name[word] {
				shared++
				score += 3
			} else if details[word] {
				shared++
				score++
			}
		}
		minimum := 2
		if len(query) == 1 {
			minimum = 1
		}
		if shared >= minimum {
			matches = append(matches, match{id, score})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].id < matches[j].id
	})
	if len(matches) > 5 {
		matches = matches[:5]
	}
	var out []summary
	for _, m := range matches {
		out = append(out, summarize(t, m.id))
	}
	return out
}

func similarWords(text string) map[string]bool {
	stop := map[string]bool{"a": true, "an": true, "the": true, "to": true, "of": true, "in": true, "for": true, "and": true, "or": true, "is": true, "it": true, "with": true, "issue": true, "fix": true, "add": true, "update": true, "support": true, "implement": true, "should": true, "be": true}
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len(w) < 3 || stop[w] {
			continue
		}
		if len(w) > 4 {
			w = strings.TrimSuffix(w, "s")
		}
		out[w] = true
	}
	return out
}
