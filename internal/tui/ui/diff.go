package ui

import (
	"strings"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// DiffLine is one line of a line diff: Op is ' ' (unchanged), '-' (only in
// the old text) or '+' (only in the new text).
type DiffLine struct {
	Op   byte
	Text string
}

// DiffLines computes a line diff with a longest common subsequence. Inputs
// are small (requirements), so the quadratic table is fine.
func DiffLines(old, new []string) []DiffLine {
	n, m := len(old), len(new)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if old[i] == new[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []DiffLine
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case old[i] == new[j]:
			out = append(out, DiffLine{' ', old[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffLine{'-', old[i]})
			i++
		default:
			out = append(out, DiffLine{'+', new[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, DiffLine{'-', old[i]})
	}
	for ; j < m; j++ {
		out = append(out, DiffLine{'+', new[j]})
	}
	return out
}

// Diff renders diff lines wrapped to width: removed lines in the error
// color, added lines in the success color, unchanged lines muted. Blank
// unchanged lines are kept so paragraphs stay apart.
func Diff(t *theme.Theme, lines []DiffLine, width int) string {
	var out []string
	for _, l := range lines {
		sign, st := "  ", t.S.Muted
		switch l.Op {
		case '-':
			sign, st = "- ", t.R.NewStyle().Foreground(t.C.Error)
		case '+':
			sign, st = "+ ", t.R.NewStyle().Foreground(t.C.Success)
		}
		if l.Text == "" {
			out = append(out, st.Render(strings.TrimRight(sign, " ")))
			continue
		}
		wrapped := strings.Split(st.Width(max(1, width-2)).Render(l.Text), "\n")
		for k, w := range wrapped {
			prefix := "  "
			if k == 0 {
				prefix = sign
			}
			out = append(out, st.UnsetWidth().Render(prefix)+strings.TrimRight(w, " "))
		}
	}
	return strings.Join(out, "\n")
}
