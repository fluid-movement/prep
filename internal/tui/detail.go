package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// mdLinkRe matches markdown links; the detail shows them as text plus path,
// since terminals cannot follow knowledge links and Glamour wraps URLs badly.
var mdLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)

// detailMarkdown builds the detail document of an issue as markdown; the
// detail pane renders it with ui.Markdown.
func detailMarkdown(t *domain.Tree, id string) string {
	i := t.Issues[id]
	if i == nil {
		return ""
	}
	var b strings.Builder
	p := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	section := func(title, body string) {
		if strings.TrimSpace(body) == "" {
			return
		}
		p("\n## %s\n\n%s\n", title, strings.TrimSpace(body))
	}

	p("# %s\n\n", i.Title)
	meta := []string{"**" + ui.StateLabel(t.State(id)) + "**", string(i.Kind), "`" + id + "`"}
	if t.Stale(id) {
		meta = append(meta, "**stale**: the requirement changed since the newest baseline")
	}
	if t.Blocked(id) {
		meta = append(meta, "blocked")
	}
	if n := len(t.Children(id)); n > 0 {
		pr := t.ChildProgress(id)
		meta = append(meta, fmt.Sprintf("%d/%d children resolved", pr.Done+pr.Dropped, pr.Total))
	}
	p("%s\n", strings.Join(meta, " · "))
	if len(i.Tags) > 0 {
		p("\n%s\n", "#"+strings.Join(i.Tags, " #"))
	}

	section("Requirement", i.Prose)
	section("Open questions", i.OpenQuestions)
	section("Context", i.Context)
	if len(i.Decisions) > 0 {
		var d strings.Builder
		for _, x := range i.Decisions {
			fmt.Fprintf(&d, "### %s: %s\n\n", x.ID, x.Title)
			line := []string{x.Date}
			if x.Supersedes != "" {
				line = append(line, "supersedes "+x.Supersedes)
			}
			if x.Outcome {
				line = append(line, "**outcome**")
			}
			fmt.Fprintf(&d, "*%s*\n\n", strings.Join(line, " · "))
			if x.Body != "" {
				d.WriteString(x.Body + "\n\n")
			}
		}
		section("Decisions", d.String())
	}
	if len(i.Criteria) > 0 {
		var c strings.Builder
		for k, x := range i.Criteria {
			mark := " "
			if x.Checked {
				mark = "x"
			}
			fmt.Fprintf(&c, "- [%s] **%d.** %s\n", mark, k+1, x.Text)
		}
		section("Acceptance", c.String())
	}
	if dod, opt := t.EffectiveDoD(id); len(dod) > 0 || len(opt) > 0 {
		var c strings.Builder
		for _, d := range dod {
			c.WriteString("- " + d + "\n")
		}
		for _, o := range opt {
			c.WriteString("- ~~" + o.Item + "~~ (opted out: " + o.Reason + ")\n")
		}
		section("Definition of Done", c.String())
	}
	if i.Findings != nil {
		section("Findings", *i.Findings)
	}
	section("History", i.History)
	if r := i.Resolution; r != nil {
		var c strings.Builder
		fmt.Fprintf(&c, "**%s** by %s at %s\n", r.Outcome, r.By, r.At.UTC().Format(time.RFC3339))
		if r.Reason != "" {
			c.WriteString("\nReason: " + r.Reason + "\n")
		}
		if r.Evidence != "" {
			c.WriteString("\nEvidence: `" + r.Evidence + "`\n")
		}
		if d := r.Documentation; d != nil {
			for _, e := range d.Entries {
				c.WriteString("\n- Knowledge: `" + e + "`")
			}
			if d.NoImpact != "" {
				c.WriteString("\n- No knowledge impact: " + d.NoImpact)
			}
			c.WriteString("\n")
		}
		if r.Note != "" {
			c.WriteString("\n" + r.Note + "\n")
		}
		section("Resolution", c.String())
	}
	return mdLinkRe.ReplaceAllString(b.String(), "$1 (`$2`)")
}
