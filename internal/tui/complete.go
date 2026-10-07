package tui

import (
	"slices"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
)

// maxSuggestions is how many candidates the filter bar shows at once.
const maxSuggestions = 6

// suggestion is one candidate under the filter bar.
type suggestion struct {
	insert string // replaces the typed part
	value  string // shown: a flag name or a value
	label  string // the title, for issue references
	count  int    // issues the value matches; -1 for flag names
}

// completion is what the word being typed can become: the rune range of
// the input it replaces and the candidates, best first.
type completion struct {
	start, end int
	items      []suggestion
}

// complete finds candidates for the word before pos in a filter bar
// query: flag names after "--", a list flag's values after the flag (the
// part after the last comma, skipping values already listed). Bare words
// search titles and get no candidates.
func complete(flags []domain.FlagInfo, text string, pos int) completion {
	runes := []rune(text)
	pos = min(max(pos, 0), len(runes))
	before := string(runes[:pos])
	start := strings.LastIndexAny(before, " \t") + 1
	word := before[start:]
	startRune := len([]rune(before[:start]))

	if strings.HasPrefix(word, "--") {
		typed := strings.TrimPrefix(word, "--")
		var names []string
		for _, f := range flags {
			names = append(names, f.Name)
		}
		var items []suggestion
		for _, n := range rank(names, func(n string) string { return n }, typed) {
			items = append(items, suggestion{insert: "--" + n + " ", value: "--" + n, count: -1})
		}
		return completion{start: startRune, end: pos, items: items}
	}

	prev := strings.Fields(before[:start])
	if len(prev) == 0 || !strings.HasPrefix(prev[len(prev)-1], "--") {
		return completion{}
	}
	name := strings.TrimPrefix(prev[len(prev)-1], "--")
	var flag *domain.FlagInfo
	for k := range flags {
		if flags[k].Name == name && flags[k].Kind == domain.FlagList {
			flag = &flags[k]
		}
	}
	if flag == nil {
		return completion{}
	}
	listed := strings.Split(word, ",")
	typed := listed[len(listed)-1]
	listed = listed[:len(listed)-1]
	var values []domain.FlagValue
	for _, v := range flag.Values {
		if !slices.Contains(listed, v.Value) {
			values = append(values, v)
		}
	}
	var items []suggestion
	for _, v := range rank(values, func(v domain.FlagValue) string { return v.Value + " " + v.Label }, typed) {
		items = append(items, suggestion{insert: v.Value, value: v.Value, label: v.Label, count: v.Count})
	}
	return completion{start: startRune + len([]rune(word)) - len([]rune(typed)), end: pos, items: items}
}

// rank keeps the items whose text contains typed (case-insensitive),
// those starting with it first, in their original order otherwise.
func rank[T any](items []T, text func(T) string, typed string) []T {
	typed = strings.ToLower(typed)
	var first, rest []T
	for _, it := range items {
		s := strings.ToLower(text(it))
		switch {
		case strings.HasPrefix(s, typed):
			first = append(first, it)
		case strings.Contains(s, typed):
			rest = append(rest, it)
		}
	}
	return append(first, rest...)
}
