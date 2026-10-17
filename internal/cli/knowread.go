package cli

import (
	"flag"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
)

// The knowledge reads: agents read entries through these rather than the
// files, so a read can be one section or an outline.

type entryLine struct {
	ID          string           `json:"id,omitempty"`
	Date        string           `json:"date,omitempty"`
	Issue       string           `json:"issue,omitempty"`
	Supersedes  []string         `json:"supersedes,omitempty"`
	Outcome     bool             `json:"outcome,omitempty"`
	Superseded  bool             `json:"superseded,omitempty"`
	Path        string           `json:"path"`
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      string           `json:"status,omitempty"`
	Scope       []string         `json:"scope,omitempty"`
	Size        int              `json:"size"`
	Sections    []domain.Section `json:"sections"`
}

func describeEntry(e *domain.Entry) entryLine {
	return entryLine{Path: e.Path, Type: e.Type, Title: e.Title, Description: e.Description, Status: e.Status, ID: e.ID, Date: e.Date, Issue: e.Issue, Supersedes: e.Supersedes, Outcome: e.Outcome, Superseded: e.Superseded, Scope: e.Scope, Size: len(e.Body), Sections: domain.Sections(e.Body)}
}

// tokens renders a byte count as approximate tokens.
func tokens(bytes int) string {
	n := (bytes + 3) / 4
	if n >= 1000 {
		return fmt.Sprintf("~%.1fk tok", float64(n)/1000)
	}
	return fmt.Sprintf("~%d tok", n)
}

func knowledgeList(a *app, args []string) error {
	limit, all := 30, false
	var query []string
	for k := 0; k < len(args); k++ {
		arg := args[k]
		switch {
		case arg == "--all":
			all = true
		case arg == "--max" || strings.HasPrefix(arg, "--max="):
			value := strings.TrimPrefix(arg, "--max=")
			if arg == "--max" {
				k++
				if k == len(args) {
					return usageErr("--max needs a number")
				}
				value = args[k]
			}
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return usageErr("--max needs a positive number")
			}
			limit = n
		default:
			query = append(query, arg)
			if strings.HasPrefix(arg, "--") && !strings.Contains(arg, "=") && k+1 < len(args) {
				k++
				query = append(query, args[k])
			}
		}
	}
	f, err := domain.ParseKnowledgeFilter(query)
	if err != nil {
		return usageErr("knowledge list: %v", err)
	}
	t, err := a.load()
	if err != nil {
		return err
	}
	a.noteRead(activity.Event{Op: "knowledge list", Verb: "listed knowledge", Area: activity.AreaKnowledge})
	paths := t.QueryKnowledge(f)
	if a.json {
		out := make([]entryLine, 0, len(paths))
		for _, p := range paths {
			out = append(out, describeEntry(t.Knowledge[p]))
		}
		a.emit(map[string]any{"entries": out})
		return nil
	}
	slices.SortFunc(paths, func(a, b string) int {
		x, y := t.Knowledge[a], t.Knowledge[b]
		if x.Superseded != y.Superseded {
			if x.Superseded {
				return 1
			}
			return -1
		}
		if c := strings.Compare(y.Date, x.Date); c != 0 {
			return c
		}
		if c := strings.Compare(y.ID, x.ID); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})
	total := len(paths)
	if !all && len(paths) > limit {
		paths = paths[:limit]
	}
	for _, p := range paths {
		e := t.Knowledge[p]
		a.printf("%s  %s  %s — %s (%s)\n", p, e.Type, recordTitle(e.Title, e.Superseded), e.Description, tokens(len(e.Body)))
	}
	if total > len(paths) {
		a.printf("%d more; --max %d or --all\n", total-len(paths), total)
	}
	if len(paths) == 0 {
		a.printf("no knowledge entries match\n")
	}
	return nil
}

func knowledgeFind(a *app, args []string) error {
	fs := flag.NewFlagSet("knowledge find", flag.ContinueOnError)
	limit := fs.Int("max", 5, "most blocks to print (0: all; explicit --max bypasses the default text budget)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		return usageErr("usage: prep knowledge find <query> [--max N]")
	}
	t, err := a.load()
	if err != nil {
		return err
	}
	a.noteRead(activity.Event{Op: "knowledge find", Verb: "searched knowledge", Target: strings.Join(pos, " "), Area: activity.AreaKnowledge})
	explicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "max" {
			explicit = true
		}
	})
	if *limit < 0 {
		return usageErr("--max must not be negative")
	}
	hits := t.FindKnowledge(strings.Join(pos, " "), 0)
	if a.json && explicit && *limit > 0 && len(hits) > *limit {
		hits = hits[:*limit]
	}
	if a.json {
		if hits == nil {
			hits = []domain.KnowledgeHit{}
		}
		a.emit(map[string]any{"hits": hits})
		return nil
	}
	if len(hits) == 0 {
		a.printf("nothing matches; prep knowledge list shows every entry\n")
		return nil
	}
	total := len(hits)
	if *limit > 0 && len(hits) > *limit {
		hits = hits[:*limit]
	}
	// Grouped by entry, entries in the order of their best block.
	var order []string
	by := map[string][]domain.KnowledgeHit{}
	for _, h := range hits {
		if by[h.Path] == nil {
			order = append(order, h.Path)
		}
		by[h.Path] = append(by[h.Path], h)
	}
	var output strings.Builder
	var blockEnds []int
	for k, p := range order {
		if k > 0 {
			fmt.Fprintf(&output, "\n")
		}
		fmt.Fprintf(&output, "== %s — %s (whole entry %s)\n", p, recordTitle(by[p][0].Title, by[p][0].Superseded), tokens(len(t.Knowledge[p].Body)))
		section := ""
		for _, h := range by[p] {
			// Label the section when it changes, unless it is the entry's own title heading.
			if h.Anchor != section && h.Heading != h.Title {
				fmt.Fprintf(&output, "[#%s %s]\n", h.Anchor, h.Heading)
			}
			section = h.Anchor
			fmt.Fprintf(&output, "%s\n", h.Text)
			blockEnds = append(blockEnds, len([]rune(output.String())))
		}
	}
	text := output.String()
	shown := len(hits)
	cut := total > shown
	if !explicit {
		const budget = 3000
		footer := fmt.Sprintf("%d more matches; --max %d\n", total, total)
		room := budget - len([]rune(footer)) - 2
		if runes := []rune(text); len(runes) > room {
			text = string(runes[:room]) + "…\n"
			cut = true
			// Count a partially printed block as omitted so --max recovers it too.
			printed := 0
			for _, end := range blockEnds {
				if end <= room {
					printed++
				}
			}
			shown = printed
		}
	}
	a.printf("%s", text)
	if cut {
		a.printf("%d more matches; --max %d\n", total-shown, total)
	}

	return nil
}

type shownEntry struct {
	entryLine
	Anchor string `json:"anchor,omitempty"`
	Text   string `json:"text,omitempty"`
}

func knowledgeShow(a *app, args []string) error {
	fs := flag.NewFlagSet("knowledge show", flag.ContinueOnError)
	outline := fs.Bool("outline", false, "title, description, scope and section headings instead of the body")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		return usageErr("usage: prep knowledge show <entry>[#section]... [--outline]")
	}
	t, err := a.load()
	if err != nil {
		return err
	}
	a.noteRead(activity.Event{Op: "knowledge show", Verb: "read knowledge", Target: strings.Join(pos, " "), Area: activity.AreaKnowledge})
	var shown []shownEntry
	for _, ref := range pos {
		raw, anchor, _ := strings.Cut(ref, "#")
		path := domain.NormalizeEntryPath(raw)
		e := t.Knowledge[path]
		if e == nil {
			return &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("no knowledge entry %s; prep knowledge list shows every entry", path)}
		}
		s := shownEntry{entryLine: describeEntry(e), Anchor: anchor}
		switch {
		case anchor != "":
			k := slices.IndexFunc(s.Sections, func(x domain.Section) bool { return x.Anchor == anchor })
			if k < 0 {
				var have []string
				for _, x := range s.Sections {
					if x.Anchor != "" {
						have = append(have, x.Anchor)
					}
				}
				return &domain.Error{Code: domain.ErrNotFound, Message: fmt.Sprintf("%s has no section #%s; sections: %s", path, anchor, strings.Join(have, ", "))}
			}
			s.Text = s.Sections[k].Text
		case !*outline:
			s.Text = e.Body
		}
		shown = append(shown, s)
	}
	if a.json {
		a.emit(map[string]any{"entries": shown})
		return nil
	}
	for k, s := range shown {
		if k > 0 {
			a.printf("\n")
		}
		ref := s.Path
		if s.Anchor != "" {
			ref += "#" + s.Anchor
		}
		a.printf("== %s — %s\n", ref, recordTitle(s.Title, s.Superseded))
		if *outline {
			a.printf("%s: %s\n", s.Type, s.Description)
			if len(s.Scope) > 0 {
				a.printf("scope: %s\n", strings.Join(s.Scope, ", "))
			}
			for _, x := range s.Sections {
				if x.Anchor == "" {
					a.printf("  (intro) %s\n", tokens(x.Size))
					continue
				}
				a.printf("  %s#%s %s (%s)\n", strings.Repeat("  ", max(x.Level-2, 0)), x.Anchor, x.Heading, tokens(x.Size))
			}
		}
		if s.Text != "" {
			a.printf("%s\n", s.Text)
		}
	}
	return nil
}
