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
// part after the last comma, skipping values already listed; a leading
// not- or ! stays, negating the value). Bare words search titles and get no
// candidates. A value typed in full has none; a flag typed in full still
// offers itself, whose insert adds the space before its value. After a
// finished flag (its value, or a switch) and a space, the flags not used
// yet are offered.
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
	if word == "" {
		if used, done := finished(flags, prev); done {
			return completion{start: pos, end: pos, items: nextFlags(flags, used)}
		}
	}
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
	for k, l := range listed {
		for _, p := range domain.NegationPrefixes {
			l = strings.TrimPrefix(l, p)
		}
		listed[k] = l
	}
	bang := "" // a typed negation prefix stays
	for _, p := range domain.NegationPrefixes {
		if strings.HasPrefix(typed, p) {
			bang = p
			break
		}
	}
	var values []domain.FlagValue
	for _, v := range flag.Values {
		if !slices.Contains(listed, v.Value) {
			values = append(values, v)
		}
	}
	// A value typed in full is done: no candidates, so enter applies.
	for _, v := range values {
		if bang+v.Value == typed {
			return completion{}
		}
	}
	var items []suggestion
	for _, v := range rank(values, func(v domain.FlagValue) string { return v.Value + " " + v.Label }, strings.TrimPrefix(typed, bang)) {
		count := v.Count
		if bang != "" {
			count = -1 // the count is for the value, not its negation
		}
		items = append(items, suggestion{insert: bang + v.Value, value: bang + v.Value, label: v.Label, count: count})
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

// finished reads the tokens before the word being typed and reports the
// flags used so far and whether the last token completes a flag: a list or
// text flag's value, or a switch. A bare word or a flag still waiting for
// its value does not.
func finished(flags []domain.FlagInfo, tokens []string) (map[string]bool, bool) {
	kinds := map[string]string{}
	for _, f := range flags {
		kinds[f.Name] = f.Kind
	}
	used := map[string]bool{}
	done := false
	for k := 0; k < len(tokens); k++ {
		name, hasVal := strings.CutPrefix(tokens[k], "--")
		if !hasVal {
			done = false // a bare word searches titles
			continue
		}
		name, _, inline := strings.Cut(name, "=")
		used[name] = true
		switch kinds[name] {
		case domain.FlagList, domain.FlagText:
			if inline {
				done = true
			} else if k+1 < len(tokens) {
				k++
				done = true
			} else {
				done = false // waiting for its value
			}
		default:
			done = true
		}
	}
	return used, done
}

// nextFlags lists the flags not used yet, in their order, then the ones
// that can repeat (--under, --text) even when used.
func nextFlags(flags []domain.FlagInfo, used map[string]bool) []suggestion {
	repeatable := map[string]bool{"under": true, "text": true}
	var first, last []suggestion
	for _, f := range flags {
		s := suggestion{insert: "--" + f.Name + " ", value: "--" + f.Name, count: -1}
		switch {
		case repeatable[f.Name]:
			last = append(last, s)
		case !used[f.Name]:
			first = append(first, s)
		}
	}
	return append(first, last...)
}
