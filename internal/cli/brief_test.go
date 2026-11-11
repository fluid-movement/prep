package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
)

// briefFixture is a project with a parent and a ready code child, and the
// decisions that reach or miss the child.
type briefFixture struct {
	*harness
	parent, child string
}

func newBriefFixture(t *testing.T) *briefFixture {
	h := newHarness(t)
	h.ok("dod", "--add", "the tests pass")
	p := h.newIssue("--title", "Exports", "--kind", "code", "--body", "Exports of every kind.")
	c := h.newIssue("--title", "CSV export", "--kind", "code", "--body", "Users can export rows as CSV.", "--parent", p)
	h.ok("define", c)
	h.write(c, "acceptance.md", "- [ ] the file has a header row\n- [ ] empty tables export a header only\n")
	h.write(c, "context.md", "Change `internal/export/csv.go`; run `go test ./...`.\n")
	h.ok("ready", c)
	h.ok("decide", c, "--title", "Semicolon delimiter", "--description", "CSV uses semicolons", "--body", "Spreadsheets in the target locale expect them.")
	h.ok("decide", p, "--title", "Exports stream", "--description", "No export buffers the whole table", "--body", "PARENT-BODY memory stays flat.")
	h.ok("decide", "--title", "Export errors wrap", "--description", "Errors carry the table name", "--body", "SCOPED-BODY for debugging.", "--scope", "internal/export/**")
	h.ok("decide", "--title", "Wide rule", "--description", "Applies to all of internal", "--body", "x", "--scope", "internal/**")
	h.ok("decide", "--title", "TUI colors", "--description", "Unrelated to exports", "--body", "x", "--scope", "internal/tui/**")
	var old struct {
		Entry string `json:"entry"`
	}
	h.jsonOf(&old, "decide", c, "--title", "Comma delimiter", "--description", "CSV uses commas", "--body", "The first idea.")
	h.ok("decide", c, "--title", "Quote every field", "--description", "All fields are quoted", "--body", "Replaces the comma idea.", "--supersedes", strings.TrimSuffix(filepath.Base(old.Entry), ".md"))
	return &briefFixture{h, p, c}
}

var briefCommandRe = regexp.MustCompile(`\bprep (guide|brief|handback|log|claim|complete|criterion|decide|finding|show|knowledge|next|prime|check|release|ready|define|edit|context)\b`)

// assertPrepFree fails when a brief names a prep command or a .prep path
// other than the report path.
func assertPrepFree(t *testing.T, text, reportPath string) {
	t.Helper()
	if m := briefCommandRe.FindString(text); m != "" {
		t.Fatalf("brief names the command %q:\n%s", m, text)
	}
	if rest := strings.ReplaceAll(text, reportPath, ""); strings.Contains(rest, ".prep/") {
		t.Fatalf("brief names a .prep path besides the report:\n%s", text)
	}
}

func (f *briefFixture) reportPath(id, name string) string {
	return mdstore.Open(f.dir).HandoffPath(id, name)
}

func TestBriefImplementer(t *testing.T) {
	f := newBriefFixture(t)
	report := f.reportPath(f.child, mdstore.HandoffReport)
	if !filepath.IsAbs(report) || !strings.HasSuffix(filepath.ToSlash(report), ".prep/local/handoff/"+f.child+"/report.md") {
		t.Fatalf("report path = %s", report)
	}
	for _, state := range []string{"ready", "in_progress"} {
		if state == "in_progress" {
			f.ok("claim", f.child)
		}
		out := f.ok("brief", f.child)
		order := []string{
			"You are asked to implement", "# CSV export",
			"## Requirement", "Users can export rows as CSV.",
			"## Implementation context", "Change `internal/export/csv.go`",
			"## Acceptance criteria", "1. the file has a header row", "2. empty tables export a header only",
			"## Decisions that apply", "### Semicolon delimiter", "Spreadsheets in the target locale expect them.", "### Quote every field",
			"- Exports stream: No export buffers the whole table", "- Export errors wrap: Errors carry the table name", "- Wide rule: Applies to all of internal",
			"## Definition of Done", "- the tests pass",
			"## Rules", "Implement exactly", "Never run git commit or git push", "Do not bend tests", "`.prep` directory",
			"## Report", "## Summary", "## Decisions", "## Findings", "## Problems",
			"Write the report to this file: " + report,
		}
		at := 0
		for _, want := range order {
			k := strings.Index(out[at:], want)
			if k < 0 {
				t.Fatalf("%s: brief lacks %q after offset %d:\n%s", state, want, at, out)
			}
			at += k
		}
		if !strings.HasSuffix(out, "Then answer with one line saying the report is written. Do not repeat the report in your answer.\n") {
			t.Fatalf("%s: brief does not end with the one-line answer:\n%s", state, out)
		}
		for _, not := range []string{"## Criteria", "Comma delimiter", "TUI colors", "prep check passes", "PARENT-BODY", "SCOPED-BODY", "Open questions", "Last failed check", f.child + " "} {
			if strings.Contains(out, not) {
				t.Fatalf("%s: brief contains %q:\n%s", state, not, out)
			}
		}
		assertPrepFree(t, out, report)
	}
}

func TestBriefChecker(t *testing.T) {
	f := newBriefFixture(t)
	f.ok("claim", f.child)
	f.ok("log", f.child, "Check failed: no header row")
	out := f.ok("brief", f.child, "--check")
	report := f.reportPath(f.child, mdstore.HandoffCheckReport)
	for _, want := range []string{
		"You are asked to check", "Users can export rows as CSV.", "1. the file has a header row", "### Semicolon delimiter", "- the tests pass",
		"You are the checker, not the implementer", "Run the checks yourself", "from the worktree or branch", "Change nothing",
		"## Summary", "## Criteria", "## Problems",
		"Write the report to this file: " + report,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("checker brief lacks %q:\n%s", want, out)
		}
	}
	for _, not := range []string{"## Findings", "## Decisions\n", "Implement exactly", "Last failed check"} {
		if strings.Contains(out, not) {
			t.Fatalf("checker brief contains %q:\n%s", not, out)
		}
	}
	assertPrepFree(t, out, report)
}

func TestBriefResearchAsksForFindings(t *testing.T) {
	h := newHarness(t)
	r := h.newIssue("--title", "Which CSV library", "--kind", "research", "--body", "Which library handles quoting?")
	h.ok("define", r)
	h.enrich(r)
	h.ok("ready", r)
	out := h.ok("brief", r)
	for _, want := range []string{"You are asked to answer the research question", "## Findings", "The report must hold your findings", "## Summary"} {
		if !strings.Contains(out, want) {
			t.Fatalf("research brief lacks %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "## Criteria") || !strings.Contains(out, "verdict for every criterion") {
		t.Fatalf("research brief lacks criteria verdict guidance:\n%s", out)
	}
	assertPrepFree(t, out, mdstore.Open(h.dir).HandoffPath(r, mdstore.HandoffReport))
}

func TestBriefResearchAlwaysAsksForCriteria(t *testing.T) {
	for _, hasCriteria := range []bool{false, true} {
		t.Run(map[bool]string{false: "no criteria", true: "with criteria"}[hasCriteria], func(t *testing.T) {
			h := newHarness(t)
			id := h.newIssue("--title", "Research", "--kind", "research", "--body", "Investigate quoting.")
			h.ok("define", id)
			h.enrich(id)
			h.ok("ready", id)
			if !hasCriteria {
				// Exercise an existing ready record without criteria; new
				// ready transitions require them, but briefs can read it.
				h.write(id, "acceptance.md", "")
			}
			for _, check := range []bool{false, true} {
				args := []string{"brief", id}
				if check {
					args = append(args, "--check")
				}
				out := h.ok(args...)
				if !strings.Contains(out, "\n## Criteria\n") || !strings.Contains(out, "except ## Summary and ## Criteria") {
					t.Fatalf("research report missing mandatory Criteria (check=%v): %s", check, out)
				}
			}
		})
	}
}

func TestBriefFailedCheck(t *testing.T) {
	f := newBriefFixture(t)
	f.ok("claim", f.child)
	f.ok("log", f.child, "Implemented, ready for check: wrote the exporter")
	if out := f.ok("brief", f.child); strings.Contains(out, "Last failed check") {
		t.Fatalf("brief shows a failed check before any:\n%s", out)
	}
	f.ok("log", f.child, "Check failed: criterion 2 not met, empty tables export nothing")
	f.ok("log", f.child, "a note in between")
	out := f.ok("brief", f.child)
	if !strings.Contains(out, "## Last failed check\n\nCheck failed: criterion 2 not met, empty tables export nothing\n") {
		t.Fatalf("brief lacks the failed check:\n%s", out)
	}
	if strings.Index(out, "## Last failed check") < strings.Index(out, "## Decisions that apply") || strings.Index(out, "## Last failed check") > strings.Index(out, "## Definition of Done") {
		t.Fatalf("failed check is out of place:\n%s", out)
	}
	assertPrepFree(t, out, f.reportPath(f.child, mdstore.HandoffReport))
	// The implementer's next hand-back makes it history.
	f.ok("log", f.child, "Implemented, ready for check: fixed the empty table")
	if out := f.ok("brief", f.child); strings.Contains(out, "Last failed check") {
		t.Fatalf("brief repeats a failed check that was answered:\n%s", out)
	}
}

func TestBriefDecisionsBound(t *testing.T) {
	f := newBriefFixture(t)
	f.ok("decide", f.child, "--title", "Long reasoning", "--description", "A very long decision", "--body", strings.Repeat("word ", 2000))
	for n := 0; n < 40; n++ {
		f.ok("decide", f.parent, "--title", "Parent rule "+strings.Repeat("x", 100), "--description", strings.Repeat("y", 100), "--body", "b")
	}
	var b domain.Brief
	f.jsonOf(&b, "brief", f.child)
	cut := 0
	for _, d := range b.Decisions {
		if d.Cut {
			cut++
			if d.Title != "Long reasoning" || !strings.Contains(d.Body, "[Cut: the first ") {
				t.Fatalf("cut decision = %+v", d)
			}
		}
	}
	if cut != 1 || b.DecisionsDropped == 0 {
		t.Fatalf("cut %d, dropped %d", cut, b.DecisionsDropped)
	}
	out := f.ok("brief", f.child)
	from := strings.Index(out, "## Decisions that apply")
	to := strings.Index(out, "## Definition of Done")
	if section := out[from:to]; len(section) > domain.BriefDecisionsBound+40 {
		t.Fatalf("decisions section is %d bytes, bound %d", len(section), domain.BriefDecisionsBound)
	}
	if !strings.Contains(out, "[Cut: the first ") || !strings.Contains(out, "[Left out to keep this section short: ") {
		t.Fatalf("brief does not say what was cut or dropped:\n%s", out)
	}
	// The issue's own decisions come first and survive the bound.
	if !strings.Contains(out, "### Semicolon delimiter") {
		t.Fatalf("own decision lost at the bound:\n%s", out)
	}
}

func TestBriefRefusals(t *testing.T) {
	f := newBriefFixture(t)
	open := f.newIssue("--title", "Open one", "--kind", "code", "--body", "Not defined.")
	f.fails("it is open", "brief", open)
	f.fails("E_GATE", "brief", open)
	f.fails("G_PARENT", "brief", f.parent, "--json")
	f.fails("work is in its children", "brief", f.parent)

	for _, kind := range []string{"decision", "manual"} {
		id := f.newIssue("--title", "A "+kind, "--kind", kind, "--body", "Something to settle.")
		f.ok("define", id)
		f.fails("a brief is for a code or research issue, this one is "+kind, "brief", id)
		f.fails("G_KIND", "brief", id, "--json")
	}

	blocked := f.newIssue("--title", "Blocked", "--kind", "code", "--body", "Needs the export.", "--depends-on", f.child)
	f.ok("define", blocked)
	f.enrich(blocked)
	f.ok("ready", blocked)
	f.fails("unresolved dependencies: "+f.child, "brief", blocked)
	f.fails("G_DEPS", "brief", blocked, "--json")

	f.replace(f.child, "issue.md", "Users can export rows as CSV.", "Users can export rows as TSV.")
	f.fails("requirement changed since its baseline", "brief", f.child)
	f.fails("G_STALE", "brief", f.child, "--json")
	f.fails("usage: prep brief", "brief")
}

func TestBriefOutAndJSON(t *testing.T) {
	f := newBriefFixture(t)
	text := f.ok("brief", f.child)
	def := f.reportPath(f.child, mdstore.HandoffBrief)
	read := func(p string) string {
		t.Helper()
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	// A bare --out, before or after the issue, uses the issue's place.
	for _, args := range [][]string{{"brief", f.child, "--out"}, {"brief", "--out", f.child}} {
		os.Remove(def)
		if out := f.ok(args...); out != def+"\n" {
			t.Fatalf("%v printed %q, want only %q", args, out, def)
		}
		if read(def) != text {
			t.Fatalf("%v wrote another text than the brief", args)
		}
	}
	checkDef := f.reportPath(f.child, mdstore.HandoffCheckBrief)
	if out := f.ok("brief", f.child, "--out", "--check"); out != checkDef+"\n" || read(checkDef) != f.ok("brief", f.child, "--check") {
		t.Fatalf("--out --check printed %q", out)
	}
	// A named file, in each spelling.
	for n, args := range [][]string{{"brief", f.child, "--out", "FILE"}, {"brief", "--out", "FILE", f.child}, {"brief", f.child, "--out=FILE"}} {
		file := filepath.Join(t.TempDir(), "b.md")
		for k := range args {
			args[k] = strings.Replace(args[k], "FILE", file, 1)
		}
		if out := f.ok(args...); out != file+"\n" || read(file) != text {
			t.Fatalf("case %d printed %q", n, out)
		}
	}

	var b domain.Brief
	f.jsonOf(&b, "brief", f.child)
	if b.ID != f.child || b.Role != domain.RoleImplementer || b.Requirement != "Users can export rows as CSV." || len(b.Criteria) != 2 ||
		!strings.Contains(b.Context, "internal/export/csv.go") || len(b.Decisions) != 5 || b.Decisions[0].Via != domain.ViaIssue || b.Decisions[2].Via != domain.ViaAncestor ||
		b.Decisions[3].Title != "Export errors wrap" || b.Decisions[3].Body != "" || len(b.DoD) != 1 || len(b.Rules) == 0 ||
		b.ReportFormat == "" || b.ReportPath != f.reportPath(f.child, mdstore.HandoffReport) {
		t.Fatalf("brief json = %+v", b)
	}
	var w struct {
		OK         bool   `json:"ok"`
		Path       string `json:"path"`
		ReportPath string `json:"report_path"`
		Role       string `json:"role"`
	}
	f.jsonOf(&w, "brief", f.child, "--check", "--out")
	if !w.OK || w.Path != checkDef || w.ReportPath != f.reportPath(f.child, mdstore.HandoffCheckReport) || w.Role != domain.RoleChecker {
		t.Fatalf("brief --out json = %+v", w)
	}
	// The hand-off files are local: nothing of them is a record or staged.
	if out := f.ok("check"); strings.Contains(out, "handoff") {
		t.Fatalf("check reports the hand-off files:\n%s", out)
	}
	if out := f.ok("help"); !strings.Contains(out, "brief <id> [--check] [--out [file]]") || strings.Index(out, "brief <id>") > strings.Index(out, "Write commands") {
		t.Fatalf("help lacks brief among the reads:\n%s", out)
	}
}

// The brief prints the report format from the definition the hand-back
// parser reads: every section it asks for is one of domain.ReportSections,
// under domain.ReportHeading.
func TestBriefReportFormatIsShared(t *testing.T) {
	f := newBriefFixture(t)
	for _, check := range []bool{false, true} {
		args := []string{"brief", f.child}
		if check {
			args = append(args, "--check")
		}
		var b domain.Brief
		f.jsonOf(&b, args...)
		text := f.ok(args...)
		if b.ReportFormat != domain.ReportFormat(b.ReportSections) || !strings.Contains(text, strings.TrimSpace(b.ReportFormat)) {
			t.Fatalf("brief does not print domain.ReportFormat:\n%s", text)
		}
		for _, s := range b.ReportSections {
			known := false
			for _, k := range domain.ReportSections {
				known = known || k == s
			}
			if !known || !strings.Contains(b.ReportFormat, "\n"+domain.ReportHeading(s)+"\n") {
				t.Fatalf("section %q is not from the shared definition", s)
			}
		}
		if check && (!strings.Contains(b.ReportFormat, domain.ReportMet+domain.ReportVerdictSep) || !strings.Contains(b.ReportFormat, domain.ReportNotMet+domain.ReportVerdictSep)) {
			t.Fatalf("criteria line format missing:\n%s", b.ReportFormat)
		}
	}
}
