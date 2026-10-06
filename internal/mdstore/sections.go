package mdstore

import (
	"strings"
)

// section is one "## " section of a record file: the heading line and the
// lines under it up to the next heading. The text before the first heading
// is a section with an empty heading.
type section struct {
	heading string   // the heading line as written; "" for the preamble
	key     string   // the heading text, lowercased and trimmed, for comparing
	lines   []string // the heading line, then the content
}

// empty reports whether the section has no content under its heading.
func (s section) empty() bool {
	return strings.TrimSpace(strings.Join(s.lines[1:], "\n")) == ""
}

// splitSections splits text at its "## " headings. Lines inside fenced
// code blocks are never headings.
func splitSections(text string) []section {
	secs := []section{{lines: []string{""}}}
	fence := ""
	for l := range strings.SplitSeq(text, "\n") {
		t := strings.TrimSpace(l)
		switch {
		case fence != "":
			if strings.HasPrefix(t, fence) {
				fence = ""
			}
		case strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~"):
			fence = t[:3]
		case strings.HasPrefix(l, "## "):
			secs = append(secs, section{heading: l, key: strings.ToLower(strings.TrimSpace(l[3:])), lines: []string{l}})
			continue
		}
		last := &secs[len(secs)-1]
		last.lines = append(last.lines, l)
	}
	secs[0].lines = secs[0].lines[1:] // the preamble has no heading line
	return secs
}

// duplicate is a "## " heading that appears more than once in a file.
type duplicate struct {
	heading string // the first copy as written, without "## "
	key     string
	count   int
	// mergeable is set when at most one copy has content, so the copies can
	// be merged without deciding where content goes.
	mergeable bool
}

// duplicateHeadings lists the headings that repeat in text, in the order
// of their first appearance.
func duplicateHeadings(text string) []duplicate {
	var out []duplicate
	at := map[string]int{}
	full := map[string]int{} // copies with content
	for _, s := range splitSections(text) {
		if s.key == "" {
			continue
		}
		if !s.empty() {
			full[s.key]++
		}
		if k, ok := at[s.key]; ok {
			out[k].count++
			continue
		}
		at[s.key] = len(out)
		out = append(out, duplicate{heading: strings.TrimSpace(s.heading[3:]), key: s.key, count: 1})
	}
	var dups []duplicate
	for _, d := range out {
		if d.count > 1 {
			d.mergeable = full[d.key] <= 1
			dups = append(dups, d)
		}
	}
	return dups
}

// mergeDuplicateSections drops the empty copies of every repeated heading
// whose copies hold content at most once: the copy with content stays where
// it is, or the first copy when all are empty. Headings repeated with
// content in several copies are left alone. It reports whether anything
// changed.
func mergeDuplicateSections(text string) (string, bool) {
	merge := map[string]bool{}
	for _, d := range duplicateHeadings(text) {
		if d.mergeable {
			merge[d.key] = true
		}
	}
	if len(merge) == 0 {
		return text, false
	}
	secs := splitSections(text)
	keep := map[string]int{} // key → index of the copy that stays
	for k, s := range secs {
		if !merge[s.key] {
			continue
		}
		if prev, ok := keep[s.key]; !ok || secs[prev].empty() && !s.empty() {
			keep[s.key] = k
		}
	}
	var lines []string
	for k, s := range secs {
		if merge[s.key] && keep[s.key] != k {
			continue
		}
		lines = append(lines, s.lines...)
	}
	return strings.Join(lines, "\n"), true
}
