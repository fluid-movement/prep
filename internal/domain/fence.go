package domain

import "strings"

// Fence follows fenced code blocks line by line. A block opens with a run
// of three or more backticks or tildes and closes with a line holding only
// a run of the same character at least as long, so a ~~~ line inside a ```
// block (or a shorter run) stays code. The zero value is outside any block.
type Fence struct {
	ch byte
	n  int
}

// Line advances over one line and reports whether it belongs to a block:
// in is true for the fence lines and the code between them, delim for the
// opening and closing fence lines only.
func (f *Fence) Line(l string) (in, delim bool) {
	t := strings.TrimSpace(l)
	c, n := fenceRun(t)
	if f.n == 0 {
		if n >= 3 {
			f.ch, f.n = c, n
			return true, true
		}
		return false, false
	}
	if c == f.ch && n >= f.n && strings.TrimSpace(t[n:]) == "" {
		f.n = 0
		return true, true
	}
	return true, false
}

// fenceRun returns the fence character a trimmed line starts with and how
// often it repeats; n is 0 when the line starts with neither.
func fenceRun(t string) (c byte, n int) {
	if t == "" || (t[0] != '`' && t[0] != '~') {
		return 0, 0
	}
	for n < len(t) && t[n] == t[0] {
		n++
	}
	return t[0], n
}
