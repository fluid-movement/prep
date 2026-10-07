package tui

import (
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestCompleteFilterWords(t *testing.T) {
	flags := []domain.FlagInfo{
		{Name: "state", Kind: domain.FlagList, Values: []domain.FlagValue{{Value: "open", Count: 3}, {Value: "defined", Count: 1}, {Value: "done", Count: 9}}},
		{Name: "tag", Kind: domain.FlagList, Values: []domain.FlagValue{{Value: "cli", Count: 2}, {Value: "tui", Count: 5}}},
		{Name: "under", Kind: domain.FlagList, Values: []domain.FlagValue{{Value: "01M4A", Label: "Release 0.2.0", Count: 8}}},
		{Name: "text", Kind: domain.FlagText},
		{Name: "stale", Kind: domain.FlagBool},
	}
	cases := []struct {
		text       string
		want       string // candidate values, space separated
		start, end int
	}{
		{"--st", "--state --stale", 0, 4},
		{"--tag tui --s", "--state --stale", 10, 13},
		{"--state ", "open defined done", 8, 8},
		{"--state d", "defined done", 8, 9},
		{"--state en", "open", 8, 10}, // substring after prefix matches
		{"--state open,d", "defined done", 13, 14},
		{"--state open,", "defined done", 13, 13}, // listed values are skipped
		{"--under rel", "01M4A", 8, 11},           // the title counts
		{"--stale ", "", 0, 0},
		{"--text ", "", 0, 0},
		{"writer", "", 0, 0},
		{"", "", 0, 0},
	}
	for _, c := range cases {
		got := complete(flags, c.text, len([]rune(c.text)))
		var values []string
		for _, s := range got.items {
			values = append(values, s.value)
		}
		if strings.Join(values, " ") != c.want || (c.want != "" && (got.start != c.start || got.end != c.end)) {
			t.Errorf("complete(%q) = %v [%d,%d), want %q [%d,%d)", c.text, values, got.start, got.end, c.want, c.start, c.end)
		}
	}
	// Completion works at the cursor, not only at the end.
	if got := complete(flags, "--sta tui", 5); len(got.items) != 2 || got.start != 0 || got.end != 5 {
		t.Errorf("mid-text completion = %+v", got)
	}
}
