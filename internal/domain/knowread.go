package domain

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Agents read knowledge through prep (list, find, show) rather than opening
// files, so a read can be as narrow as one section of one entry.

// Section is one part of an entry's body under a heading. Text runs from the
// heading to the next heading of the same or a higher level, so it includes
// its subsections; Own stops at the next heading of any level.
type Section struct {
	Anchor  string `json:"anchor"` // slug of the heading; "" for the text before the first heading
	Heading string `json:"heading,omitempty"`
	Level   int    `json:"level,omitempty"`
	Size    int    `json:"size"` // bytes of Text
	Text    string `json:"-"`
	Own     string `json:"-"`
}

// Sections splits an entry body at its markdown headings, ignoring lines in
// fenced code blocks. Text before the first heading, if any, comes first
// with an empty anchor. Repeated anchors get -1, -2, ... as on GitHub.
func Sections(body string) []Section {
	lines := strings.Split(body, "\n")
	type head struct{ line, level int }
	var heads []head
	var fence Fence
	for k, l := range lines {
		if in, _ := fence.Line(l); in {
			continue
		}
		if n := headingLevel(l); n > 0 {
			heads = append(heads, head{k, n})
		}
	}
	join := func(from, to int) string { return strings.TrimSpace(strings.Join(lines[from:to], "\n")) }
	var out []Section
	first := len(lines)
	if len(heads) > 0 {
		first = heads[0].line
	}
	if intro := join(0, first); intro != "" {
		out = append(out, Section{Size: len(intro), Text: intro, Own: intro})
	}
	seen := map[string]int{}
	for k, h := range heads {
		end, own := len(lines), len(lines)
		if k+1 < len(heads) {
			own = heads[k+1].line
		}
		for _, n := range heads[k+1:] {
			if n.level <= h.level {
				end = n.line
				break
			}
		}
		heading := strings.TrimSpace(strings.TrimLeft(lines[h.line], "#"))
		anchor := Slug(heading)
		if c := seen[anchor]; c > 0 {
			seen[anchor] = c + 1
			anchor = anchor + "-" + strconv.Itoa(c)
		} else {
			seen[anchor] = 1
		}
		text := join(h.line, end)
		out = append(out, Section{Anchor: anchor, Heading: heading, Level: h.level, Size: len(text), Text: text, Own: join(h.line, own)})
	}
	return out
}

func headingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// Slug turns a heading into its anchor: lower case, letters and digits kept,
// spaces and hyphens joined by single hyphens, everything else dropped.
func Slug(heading string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			dash = true
		}
	}
	return b.String()
}

// KnowledgeHit is one block of an entry, a paragraph or a top-level list
// item with what it nests, matching a knowledge search.
type KnowledgeHit struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Anchor  string `json:"anchor,omitempty"` // the section the block is in
	Heading string `json:"heading,omitempty"`
	Text    string `json:"text"`
	Score   int    `json:"score"`
	Matched int    `json:"matched"` // how many query terms the block or its heading matched
}

// FindKnowledge ranks the blocks of all entries against a query, so a search
// returns the few paragraphs or list items that answer it rather than whole
// entries. A term matches a word exactly, as its prefix, inside it, or one
// typo away (terms of five letters or more). A block must match a term
// itself; terms count as matched in the block or its section heading, and
// matches in its entry's title, description and scope add to its score.
// Blocks matching more terms come first, then by score. At most limit hits
// (0: all).
func (t *Tree) FindKnowledge(query string, limit int) []KnowledgeHit {
	terms := words(query)
	if len(terms) == 0 {
		return nil
	}
	var hits []KnowledgeHit
	for path, e := range t.Knowledge {
		title, desc := words(e.Title), words(e.Description)
		scope := words(strings.Join(e.Scope, " ") + " " + path)
		for _, s := range Sections(e.Body) {
			heading := words(s.Heading)
			for _, b := range Blocks(s.Own) {
				body := words(b)
				h := KnowledgeHit{Path: path, Title: e.Title, Anchor: s.Anchor, Heading: s.Heading, Text: b}
				own := false
				for _, term := range terms {
					inBody := 0
					for _, w := range body {
						inBody += matchWord(term, w)
						if inBody >= 9 {
							break
						}
					}
					inHeading := match(term, heading)
					h.Score += max(4*match(term, title), 3*match(term, desc), 3*inHeading, 2*match(term, scope)) + 2*inBody
					if inBody > 0 {
						own = true
					}
					if inBody > 0 || inHeading > 0 {
						h.Matched++
					}
				}
				if own {
					hits = append(hits, h)
				}
			}
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		a, b := hits[i], hits[j]
		if a.Matched != b.Matched {
			return a.Matched > b.Matched
		}
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Text < b.Text
	})
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

// Blocks splits section text into paragraphs and top-level list items (with
// their nested lines), leaving out headings. A fenced code block stays whole.
func Blocks(text string) []string {
	var out []string
	var cur []string
	flush := func() {
		if b := strings.TrimSpace(strings.Join(cur, "\n")); b != "" {
			out = append(out, b)
		}
		cur = nil
	}
	var fence Fence
	for _, l := range strings.Split(text, "\n") {
		t := strings.TrimSpace(l)
		switch in, _ := fence.Line(l); {
		case in:
			cur = append(cur, l)
		case t == "":
			flush()
		case headingLevel(l) > 0:
			flush()
		case isListItem(l):
			flush()
			cur = append(cur, l)
		default:
			cur = append(cur, l)
		}
	}
	flush()
	return out
}

// isListItem reports whether a line starts a top-level list item.
func isListItem(l string) bool {
	if strings.HasPrefix(l, "- ") || strings.HasPrefix(l, "* ") || strings.HasPrefix(l, "+ ") {
		return true
	}
	k := 0
	for k < len(l) && l[k] >= '0' && l[k] <= '9' {
		k++
	}
	return k > 0 && k+1 < len(l) && (l[k] == '.' || l[k] == ')') && l[k+1] == ' '
}

// words lower-cases s and splits it into words of letters and digits.
func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// match is the best matchWord of term against ws.
func match(term string, ws []string) int {
	best := 0
	for _, w := range ws {
		if m := matchWord(term, w); m > best {
			best = m
			if best == 3 {
				break
			}
		}
	}
	return best
}

// matchWord scores term against one word: 3 exact, 2 prefix, 1 inside the
// word or one typo away.
func matchWord(term, w string) int {
	switch {
	case term == w:
		return 3
	case strings.HasPrefix(w, term):
		return 2
	case len(term) >= 3 && strings.Contains(w, term):
		return 1
	case len(term) >= 5 && oneEdit(term, w):
		return 1
	}
	return 0
}

// oneEdit reports whether a and b differ by one insertion, deletion,
// substitution or swap of adjacent letters.
func oneEdit(a, b string) bool {
	if a == b {
		return false
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) > len(rb) {
		ra, rb = rb, ra
	}
	if len(rb)-len(ra) > 1 {
		return false
	}
	k := 0
	for k < len(ra) && ra[k] == rb[k] {
		k++
	}
	if len(ra) == len(rb) {
		if k+1 < len(ra) && ra[k] == rb[k+1] && ra[k+1] == rb[k] {
			return string(ra[k+2:]) == string(rb[k+2:])
		}
		return string(ra[k+1:]) == string(rb[k+1:])
	}
	return string(ra[k:]) == string(rb[k+1:])
}
