package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// harness runs prep commands in a temporary project with a controllable clock.
type harness struct {
	t   *testing.T
	dir string
	now time.Time
}

func newHarness(t *testing.T) *harness {
	h := &harness{t: t, dir: t.TempDir(), now: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)}
	clock = func() time.Time { h.now = h.now.Add(time.Second); return h.now }
	t.Cleanup(func() { clock = time.Now })
	h.ok("init", "--no-bootstrap")
	h.ok("knowledge", "new", "/overview.md", "--type", "overview", "--title", "Project overview", "--description", "Entry point to the knowledge base.", "--body", "The test project.")
	return h
}

func (h *harness) run(args ...string) (int, string) {
	var out, errb bytes.Buffer
	code := Main(append(args, "--root", h.dir), strings.NewReader(""), &out, &errb)
	return code, out.String() + errb.String()
}

func (h *harness) ok(args ...string) string {
	h.t.Helper()
	code, out := h.run(args...)
	if code != 0 {
		h.t.Fatalf("prep %s: exit %d\n%s", strings.Join(args, " "), code, out)
	}
	return out
}

// fails asserts a non-zero exit and that the output mentions want.
func (h *harness) fails(want string, args ...string) {
	h.t.Helper()
	code, out := h.run(args...)
	if code == 0 {
		h.t.Fatalf("prep %s: expected failure, got\n%s", strings.Join(args, " "), out)
	}
	if !strings.Contains(out, want) {
		h.t.Fatalf("prep %s: output lacks %q:\n%s", strings.Join(args, " "), want, out)
	}
}

func (h *harness) jsonOf(v any, args ...string) {
	h.t.Helper()
	out := h.ok(append(args, "--json")...)
	if err := json.Unmarshal([]byte(out), v); err != nil {
		h.t.Fatalf("prep %s: bad json: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func (h *harness) newIssue(args ...string) string {
	h.t.Helper()
	var r writeResult
	h.jsonOf(&r, append([]string{"new"}, args...)...)
	return r.ID
}

func (h *harness) state(id string) (string, bool) {
	h.t.Helper()
	var s struct {
		State string `json:"state"`
		Stale bool   `json:"stale"`
	}
	h.jsonOf(&s, "show", id)
	return s.State, s.Stale
}

func (h *harness) path(id, name string) string {
	return filepath.Join(h.dir, ".prep", "issues", id, name)
}

func (h *harness) write(id, name, content string) {
	h.t.Helper()
	if err := os.MkdirAll(filepath.Dir(h.path(id, name)), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(h.path(id, name), []byte(content), 0o644); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) read(id, name string) string {
	h.t.Helper()
	b, err := os.ReadFile(h.path(id, name))
	if err != nil {
		h.t.Fatal(err)
	}
	return string(b)
}

func (h *harness) replace(id, name, old, new string) {
	h.t.Helper()
	c := h.read(id, name)
	if !strings.Contains(c, old) {
		h.t.Fatalf("%s/%s lacks %q", id, name, old)
	}
	h.write(id, name, strings.Replace(c, old, new, 1))
}

func (h *harness) enrich(id string) {
	h.write(id, "acceptance.md", "- [ ] it works\n")
	h.write(id, "context.md", "Change `internal/x.go`.\n")
}

func (h *harness) expectState(id, want string, wantStale bool) {
	h.t.Helper()
	s, stale := h.state(id)
	if s != want || stale != wantStale {
		h.t.Fatalf("%s: state %s stale %v, want %s stale %v", id, s, stale, want, wantStale)
	}
}

func listIDs(h *harness, args ...string) []string {
	var r struct {
		Issues []summary `json:"issues"`
	}
	h.jsonOf(&r, args...)
	var ids []string
	for _, s := range r.Issues {
		ids = append(ids, s.ID)
	}
	return ids
}

func TestLifecycle(t *testing.T) {
	h := newHarness(t)
	a := h.newIssue("--title", "Export", "--kind", "code", "--body", "Export rows as CSV.")
	h.expectState(a, "open", false)

	// Define is gated on an empty Open questions section.
	h.replace(a, "issue.md", "## Open questions", "## Open questions\n\n- which delimiter?")
	h.fails("G_OPEN_QUESTIONS", "define", a, "--json")
	h.replace(a, "issue.md", "\n\n- which delimiter?", "")
	h.ok("define", a)
	h.expectState(a, "defined", false)
	h.fails("G_NOT_STALE", "define", a, "--json")

	// Ready is gated on enrichment.
	h.fails("G_CRITERIA", "ready", a, "--json")
	h.enrich(a)
	h.ok("ready", a)
	h.expectState(a, "ready", false)
	if got := listIDs(h, "next"); len(got) != 1 || got[0] != a {
		t.Fatalf("next = %v", got)
	}

	// A trivial change makes the issue stale; ack keeps it ready.
	h.replace(a, "issue.md", "Export rows as CSV.", "Export the rows as CSV.")
	h.expectState(a, "ready", true)
	if got := listIDs(h, "next"); len(got) != 0 {
		t.Fatalf("stale issue in next: %v", got)
	}
	if got := listIDs(h, "list", "--stale"); len(got) != 1 {
		t.Fatalf("list --stale = %v", got)
	}
	h.fails("G_STALE", "claim", a, "--json")
	h.ok("ack", a)
	h.expectState(a, "ready", false)

	// A real change: define again invalidates ready.md, re-enrich, ready again.
	h.replace(a, "issue.md", "Export the rows as CSV.", "Export the rows as CSV and JSON.")
	h.ok("define", a)
	h.expectState(a, "defined", false)
	if out := h.ok("check"); !strings.Contains(out, "I013") {
		t.Fatalf("check should warn about outdated ready.md:\n%s", out)
	}
	h.ok("ready", a)

	// Kind is part of the baseline.
	h.replace(a, "issue.md", "kind: code", "kind: manual")
	h.expectState(a, "ready", true)
	h.replace(a, "issue.md", "kind: manual", "kind: code")
	h.expectState(a, "ready", false)

	// Dependencies block claiming.
	b := h.newIssue("--title", "Import", "--kind", "code", "--body", "Import rows.", "--depends-on", a[len(a)-4:])
	h.ok("define", b)
	h.enrich(b)
	h.ok("ready", b)
	h.fails("G_DEPS", "claim", b, "--json")
	if got := listIDs(h, "list", "--blocked"); len(got) != 1 || got[0] != b {
		t.Fatalf("list --blocked = %v", got)
	}

	// Claim, release (recorded in history), claim, complete.
	h.ok("claim", a, "--by", "claude-code/2.0")
	h.expectState(a, "in_progress", false)
	h.ok("release", a, "--reason", "switching tasks")
	if !strings.Contains(h.read(a, "history.md"), "released by") {
		t.Fatalf("release not recorded in history.md")
	}
	h.expectState(a, "ready", false)
	h.ok("claim", a)
	h.fails("G_UNCHECKED", "complete", a, "--commit", "abc1234", "--no-impact", "none", "--json")
	h.replace(a, "acceptance.md", "- [ ]", "- [x]")
	h.fails("G_EVIDENCE", "complete", a, "--no-impact", "none", "--json")
	h.fails("G_DOCS", "complete", a, "--commit", "abc1234", "--json")
	h.fails("G_ACTOR", "complete", a, "--commit", "abc1234", "--no-impact", "none", "--by", "human:andre", "--json")
	h.ok("complete", a, "--commit", "abc1234", "--docs", "overview")
	h.expectState(a, "done", false)
	res := h.read(a, "resolution.md")
	for _, want := range []string{"outcome: done", "evidence: abc1234", "/overview.md", "prep check passes"} {
		if !strings.Contains(res, want) {
			t.Fatalf("resolution.md lacks %q:\n%s", want, res)
		}
	}
	if got := listIDs(h, "next"); len(got) != 1 || got[0] != b {
		t.Fatalf("next after completing dependency = %v", got)
	}

	// Drop needs a reason.
	h.fails("G_REASON", "drop", b, "--json")
	h.ok("drop", b, "--reason", "not needed")
	h.expectState(b, "dropped", false)
	h.fails("G_STATE", "claim", b, "--json")

	if out := h.ok("check"); !strings.Contains(out, "0 errors") {
		t.Fatalf("check:\n%s", out)
	}
	if code, out := h.run("fmt", "--check"); code != 0 {
		t.Fatalf("tool output is not canonical:\n%s", out)
	}
}

func TestParentCompletesWhenChildrenResolved(t *testing.T) {
	h := newHarness(t)
	p := h.newIssue("--title", "Goal", "--kind", "code", "--body", "Reach the goal.")
	c := h.newIssue("--title", "Step", "--kind", "manual", "--body", "Do the step.", "--parent", p)
	h.ok("define", p)
	h.write(p, "acceptance.md", "- [ ] goal reached\n")
	h.ok("ready", p) // parents need no context
	h.fails("G_PARENT", "claim", p, "--json")
	h.fails("G_CHILDREN", "complete", p, "--no-impact", "n/a", "--json")

	h.ok("define", c)
	h.write(c, "acceptance.md", "- [ ] done by hand\n")
	h.ok("ready", c) // manual issues need no context
	h.ok("claim", c)
	h.replace(c, "acceptance.md", "- [ ]", "- [x]")
	h.ok("complete", c, "--no-impact", "manual step", "--by", "human:andre")

	h.fails("G_UNCHECKED", "complete", p, "--no-impact", "n/a", "--json")
	h.replace(p, "acceptance.md", "- [ ]", "- [x]")
	h.ok("complete", p, "--no-impact", "n/a")
	h.expectState(p, "done", false)

	var ids []string
	ids = listIDs(h, "list", "--under", p)
	if len(ids) != 1 || ids[0] != c {
		t.Fatalf("list --under = %v", ids)
	}
}

func TestResearchAndDecisionEvidence(t *testing.T) {
	h := newHarness(t)
	r := h.newIssue("--title", "Which parser?", "--kind", "research", "--body", "Find a parser.")
	d := h.newIssue("--title", "Pick a format", "--kind", "decision", "--body", "Choose the format.")
	for _, id := range []string{r, d} {
		h.ok("define", id)
		h.enrich(id)
		h.replace(id, "acceptance.md", "- [ ]", "- [x]")
		h.ok("ready", id)
		h.ok("claim", id)
	}
	h.fails("G_FINDINGS", "complete", r, "--no-impact", "x", "--json")
	h.write(r, "findings.md", "Use the stdlib parser.\n")
	h.ok("complete", r, "--no-impact", "x")

	h.write(d, "decisions.md", "## D1: Use CSV\ndate: 2026-10-05\n\nDraft.\n")
	h.fails("G_OUTCOME", "complete", d, "--no-impact", "x", "--json")
	h.write(d, "decisions.md", "## D1: Use CSV\ndate: 2026-10-05\n\nDraft.\n\n## D2: Use JSON\ndate: 2026-10-06\nsupersedes: D1\noutcome: true\n\nAgreed with the user.\n")
	h.ok("complete", d, "--no-impact", "x")

	// The decision's outcome flows into dependents through guide.
	e := h.newIssue("--title", "Implement", "--kind", "code", "--body", "Build it.", "--depends-on", d, "--depends-on", r)
	out := h.ok("guide", e)
	if !strings.Contains(out, d+"/decisions.md") || !strings.Contains(out, r+"/findings.md") {
		t.Fatalf("guide does not point at upstream outcomes:\n%s", out)
	}
}

func TestWriteValidationRejectsInvalidState(t *testing.T) {
	h := newHarness(t)
	h.fails("E_NOT_FOUND", "new", "--title", "X", "--kind", "code", "--parent", "99999999-999999", "--json")
	h.fails("G_KIND", "new", "--title", "X", "--kind", "chore", "--json")
	a := h.newIssue("--title", "A", "--kind", "code", "--body", "A.")
	b := h.newIssue("--title", "B", "--kind", "code", "--body", "B.", "--depends-on", a)
	// A hand edit creates a dependency cycle; writes touching the tree must not add errors,
	// but existing errors do not block unrelated writes.
	h.replace(a, "issue.md", "kind: code", "kind: code\ndepends_on:\n  - "+b)
	if code, out := h.run("check"); code == 0 || !strings.Contains(out, "I009") {
		t.Fatalf("check should report the cycle:\n%s", out)
	}
	h.ok("new", "--title", "C", "--kind", "code")
}

func TestGuideJSONAndPrime(t *testing.T) {
	h := newHarness(t)
	p := h.newIssue("--title", "Goal", "--kind", "code", "--body", "Goal.")
	c := h.newIssue("--title", "Step", "--kind", "code", "--body", "Step.", "--parent", p)
	var g struct {
		State       string `json:"state"`
		Step        string `json:"step"`
		Transitions []struct {
			Op      string `json:"op"`
			Allowed bool   `json:"allowed"`
		} `json:"transitions"`
		Knowledge []struct {
			Path string `json:"path"`
		} `json:"knowledge"`
	}
	h.jsonOf(&g, "guide", c)
	if g.State != "open" || g.Step != "define" || len(g.Transitions) != 2 || g.Transitions[0].Op != "define" || !g.Transitions[0].Allowed {
		t.Fatalf("guide = %+v", g)
	}
	if len(g.Knowledge) != 1 || g.Knowledge[0].Path != "/overview.md" {
		t.Fatalf("knowledge candidates = %+v", g.Knowledge)
	}
	var pr primeBrief
	h.jsonOf(&pr, "prime")
	if pr.Issues != 2 || len(pr.Parents) != 1 || pr.Parents[0].ID != p {
		t.Fatalf("prime = %+v", pr)
	}
}

func TestKnowledgeRetrievalByScope(t *testing.T) {
	h := newHarness(t)
	kdir := filepath.Join(h.dir, ".prep", "knowledge", "components")
	if err := os.MkdirAll(kdir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(kdir, "export.md"), []byte("---\ntype: component\ntitle: Export\ndescription: Streams rows to CSV.\nstatus: stable\nscope:\n  - internal/export/**\n---\n\nSee [overview](/overview.md).\n"), 0o644)
	os.WriteFile(filepath.Join(kdir, "parser.md"), []byte("---\ntype: component\ntitle: Parser\ndescription: Parses rows.\nscope: internal/parse\n---\n\nSee [missing](../nope.md).\n"), 0o644)
	a := h.newIssue("--title", "A", "--kind", "code", "--body", "A.")
	h.write(a, "context.md", "Touches `internal/export/csv.go`.\n")
	var g struct {
		Knowledge []struct{ Path, Via string } `json:"knowledge"`
	}
	h.jsonOf(&g, "guide", a)
	if len(g.Knowledge) != 2 || g.Knowledge[0].Path != "/components/export.md" || g.Knowledge[1].Path != "/overview.md" {
		t.Fatalf("candidates = %+v", g.Knowledge)
	}
	if code, out := h.run("check"); code == 0 || !strings.Contains(out, "K003") {
		t.Fatalf("check should report the broken link:\n%s", out)
	}
}

// The static skill copy must match what the binary serves.
func TestStaticSkillMatchesBinary(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".claude", "skills", "prep", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != skillText {
		t.Fatal(".claude/skills/prep/SKILL.md differs from internal/cli/skill.md; copy it over")
	}
}

func TestEdit(t *testing.T) {
	h := newHarness(t)
	a := h.newIssue("--title", "Export", "--kind", "code", "--body", "Export rows as CSV.")
	b := h.newIssue("--title", "Import", "--kind", "code", "--body", "Import rows.")
	c := h.newIssue("--title", "Parse", "--kind", "code", "--body", "Parse rows.")

	h.fails("E_USAGE", "edit", a, "--json")
	h.fails("G_KIND", "edit", a, "--kind", "story", "--json")
	h.fails("G_TITLE", "edit", a, "--title", " ", "--json")

	// Flags not given keep their field; the edit is recorded in history.
	h.ok("edit", a, "--title", "Export rows", "--by", "claude-code/2.0")
	issue := h.read(a, "issue.md")
	for _, want := range []string{"title: Export rows", "kind: code", "Export rows as CSV."} {
		if !strings.Contains(issue, want) {
			t.Fatalf("issue.md lacks %q:\n%s", want, issue)
		}
	}
	if hist := h.read(a, "history.md"); !strings.Contains(hist, "edited by claude-code/2.0: title") {
		t.Fatalf("edit not recorded in history.md:\n%s", hist)
	}

	// Dependencies: repeated flags replace the list, '' clears it.
	h.ok("edit", a, "--depends-on", b, "--depends-on", c[len(c)-4:])
	if got := h.read(a, "issue.md"); !strings.Contains(got, "  - "+b+"\n  - "+c+"\n") {
		t.Fatalf("depends_on not written:\n%s", got)
	}
	h.ok("edit", a, "--depends-on", c)
	if got := h.read(a, "issue.md"); strings.Contains(got, b) {
		t.Fatalf("depends_on not replaced:\n%s", got)
	}
	h.fails("E_USAGE", "edit", a, "--depends-on", "", "--depends-on", b, "--json")
	h.fails("E_INVALID", "edit", a, "--depends-on", a, "--json")
	h.fails("E_INVALID", "edit", c, "--depends-on", a, "--json")
	h.fails("E_NOT_FOUND", "edit", a, "--depends-on", "19990101-000000", "--json")
	h.ok("edit", a, "--depends-on", "")
	if got := h.read(a, "issue.md"); strings.Contains(got, "depends_on") {
		t.Fatalf("depends_on not cleared:\n%s", got)
	}

	// Parent: set, reject cycles, remove.
	h.ok("edit", b, "--parent", a)
	if got := h.read(b, "issue.md"); !strings.Contains(got, "parent: "+a) {
		t.Fatalf("parent not written:\n%s", got)
	}
	h.fails("E_INVALID", "edit", a, "--parent", b, "--json")
	h.fails("E_INVALID", "edit", b, "--depends-on", a, "--json")
	h.ok("edit", b, "--parent", "")
	if got := h.read(b, "issue.md"); strings.Contains(got, "parent:") {
		t.Fatalf("parent not removed:\n%s", got)
	}

	// A requirement with its own Open questions section keeps exactly one.
	h.ok("edit", a, "--body", "Export rows as CSV and JSON.\n\n## Open questions\n\n- which delimiter?")
	if got := h.read(a, "issue.md"); strings.Count(got, "## Open questions") != 1 || !strings.Contains(got, "- which delimiter?") {
		t.Fatalf("open questions not kept once:\n%s", got)
	}
	h.fails("G_OPEN_QUESTIONS", "define", a, "--json")
	h.ok("edit", a, "--body", "Export rows as CSV and JSON.")
	if got := h.read(a, "issue.md"); strings.Count(got, "## Open questions") != 1 {
		t.Fatalf("empty open questions section missing:\n%s", got)
	}

	// Requirement and kind edits make a defined issue stale.
	h.ok("define", a)
	var r writeResult
	h.jsonOf(&r, "edit", a, "--body", "Export rows as CSV, JSON and XML.")
	if r.State != "defined" || !r.Stale {
		t.Fatalf("edit result = %+v, want defined and stale", r)
	}
	h.ok("ack", a)
	if out := h.ok("edit", a, "--kind", "manual"); !strings.Contains(out, "stale") {
		t.Fatalf("kind edit not reported as stale:\n%s", out)
	}
	h.ok("edit", a, "--title", "Export formats")
	h.expectState(a, "defined", true)

	// Dropped targets and resolved issues are refused.
	h.ok("drop", c, "--reason", "not needed")
	h.fails("G_DEPS", "edit", b, "--depends-on", c, "--json")
	h.fails("G_STATE", "edit", c, "--title", "Parse rows", "--json")
}

func TestNewKeepsOneOpenQuestionsSection(t *testing.T) {
	h := newHarness(t)
	a := h.newIssue("--title", "Export", "--kind", "code", "--body", "Export rows.\n\n## Open questions\n\n- which format?")
	if got := h.read(a, "issue.md"); strings.Count(got, "## Open questions") != 1 {
		t.Fatalf("prep new duplicated the section:\n%s", got)
	}
}

func TestRecordCommands(t *testing.T) {
	h := newHarness(t)
	a := h.newIssue("--title", "Export", "--kind", "code", "--body", "Export rows as CSV.")
	q := h.newIssue("--title", "Which format", "--kind", "decision", "--body", "Pick a format.")
	r := h.newIssue("--title", "Survey", "--kind", "research", "--body", "Survey formats.")

	// context and findings replace their record; findings only on research.
	h.fails("E_USAGE", "context", a, "--json")
	h.ok("context", a, "--body", "Change `internal/export.go`.")
	h.ok("context", a, "--body", "Change `internal/csv.go`.")
	if got := h.read(a, "context.md"); got != "Change `internal/csv.go`.\n" {
		t.Fatalf("context.md = %q", got)
	}
	h.fails("G_KIND", "findings", a, "--body", "x", "--json")
	h.ok("findings", r, "--body", "CSV is enough.")
	if got := h.read(r, "findings.md"); got != "CSV is enough.\n" {
		t.Fatalf("findings.md = %q", got)
	}

	// decide appends numbered, dated entries the parser reads back.
	h.ok("decide", a, "--title", "Use encoding/csv", "--body", "Standard library.")
	h.ok("decide", a, "--title", "Stream rows", "--supersedes", "D1")
	h.fails("G_DECISION", "decide", a, "--title", "x", "--supersedes", "D9", "--json")
	h.fails("G_KIND", "decide", a, "--title", "x", "--outcome", "--json")
	h.fails("G_TITLE", "decide", a, "--json")
	dec := h.read(a, "decisions.md")
	for _, want := range []string{"## D1: Use encoding/csv\ndate: 2026-10-05\n\nStandard library.", "## D2: Stream rows\ndate: 2026-10-05\nsupersedes: D1"} {
		if !strings.Contains(dec, want) {
			t.Fatalf("decisions.md lacks %q:\n%s", want, dec)
		}
	}
	h.ok("decide", q, "--title", "CSV", "--outcome")
	if !strings.Contains(h.read(q, "decisions.md"), "outcome: true") {
		t.Fatalf("outcome not written")
	}

	// criterion keeps text a human wrote and numbers like prep show.
	h.write(a, "acceptance.md", "Criteria for the export:\n\n- [ ] writes a header\n- [ ] quotes fields\n\n## Definition of Done\n\n- benchmarks run\n")
	h.ok("criterion", a, "--check", "2", "--add", "handles empty input")
	got := h.read(a, "acceptance.md")
	want := "Criteria for the export:\n\n- [ ] writes a header\n- [x] quotes fields\n- [ ] handles empty input\n\n## Definition of Done\n\n- benchmarks run\n"
	if got != want {
		t.Fatalf("acceptance.md =\n%s\nwant\n%s", got, want)
	}
	if out := h.ok("show", a); !strings.Contains(out, "2. [x] quotes fields") {
		t.Fatalf("show does not number criteria:\n%s", out)
	}
	h.fails("G_CRITERION", "criterion", a, "--check", "4", "--json")
	h.ok("criterion", a, "--remove", "1", "--uncheck", "2")
	if got := h.read(a, "acceptance.md"); strings.Contains(got, "writes a header") || !strings.Contains(got, "- [ ] quotes fields") {
		t.Fatalf("remove/uncheck failed:\n%s", got)
	}

	// dod adds, opts out with a reason and removes by item.
	h.fails("G_DOD", "dod", a, "--opt-out", "prep fmt --check passes", "--json")
	h.ok("dod", a, "--add", "docs updated", "--opt-out", "prep fmt --check passes", "--reason", "generated files")
	got = h.read(a, "acceptance.md")
	if !strings.Contains(got, "- benchmarks run\n- docs updated\n- opt-out: prep fmt --check passes — generated files\n") {
		t.Fatalf("dod lines missing:\n%s", got)
	}
	h.ok("dod", a, "--remove", "benchmarks run", "--remove", "prep fmt --check passes")
	got = h.read(a, "acceptance.md")
	if strings.Contains(got, "benchmarks run") || strings.Contains(got, "opt-out") || !strings.Contains(got, "- docs updated") {
		t.Fatalf("dod remove failed:\n%s", got)
	}
	h.fails("G_DOD", "dod", a, "--remove", "nothing like this", "--json")

	// A new issue without acceptance text gets criteria and a DoD section.
	h.ok("criterion", r, "--add", "formats compared")
	h.ok("dod", r, "--add", "sources linked")
	if got := h.read(r, "acceptance.md"); got != "- [ ] formats compared\n\n## Definition of Done\n\n- sources linked\n" {
		t.Fatalf("acceptance.md from empty = %q", got)
	}

	// log appends a line naming the actor.
	h.ok("log", a, "tried", "encoding/csv", "--by", "claude-code/2.0")
	if !strings.Contains(h.read(a, "history.md"), "claude-code/2.0: tried encoding/csv") {
		t.Fatalf("log line missing:\n%s", h.read(a, "history.md"))
	}

	// The records satisfy the lifecycle and every file stays canonical.
	h.ok("define", a)
	h.ok("ready", a)
	h.ok("fmt", "--check")
	h.ok("check")

	// Resolved issues are refused.
	h.ok("drop", r, "--reason", "not needed")
	h.fails("G_STATE", "log", r, "late note", "--json")
}

func TestKnowledgeCommands(t *testing.T) {
	h := newHarness(t)
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t@t", "-c", "user.name=t"}, args...)...)
		cmd.Dir = h.dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	entry := func(p string) string {
		b, err := os.ReadFile(filepath.Join(h.dir, ".prep", "knowledge", filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	h.fails("--title is required", "knowledge", "new", "components/export.md", "--type", "component", "--json")
	h.fails("K003", "knowledge", "new", "components/export.md", "--type", "component", "--title", "Export", "--description", "Export: formats", "--body", "See [gone](/gone.md).", "--json")
	h.fails("K002", "knowledge", "new", "components/export.md", "--type", "widget", "--title", "Export", "--description", "d", "--body", "b", "--json")
	h.ok("knowledge", "new", "components/export.md", "--type", "component", "--title", "Export", "--description", "Export: formats and writers", "--scope", "internal/export/**", "--body", "See the [overview](/overview.md).", "--by", "claude-code/2.0")
	got := entry("components/export.md")
	for _, want := range []string{"type: component\ntitle: Export\ndescription: 'Export: formats and writers'\ngenerated:\n  by: claude-code/2.0\n  at: 2026-10-05T", "scope:\n  - internal/export/**\n---\n\nSee the [overview](/overview.md).\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("new entry lacks %q:\n%s", want, got)
		}
	}
	h.fails("already exists", "knowledge", "new", "components/export.md", "--type", "component", "--title", "x", "--description", "x", "--body", "x", "--json")

	// Update changes only what is given and keeps keys prep does not model.
	p := filepath.Join(h.dir, ".prep", "knowledge", "components", "export.md")
	if err := os.WriteFile(p, []byte(strings.Replace(got, "scope:", "sources:\n  - https://example.com\nscope:", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	h.ok("knowledge", "update", "/components/export.md", "--status", "stable", "--body", "Rewritten.")
	got = entry("components/export.md")
	for _, want := range []string{"status: stable", "sources:\n  - https://example.com", "title: Export", "\n---\n\nRewritten.\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("updated entry lacks %q:\n%s", want, got)
		}
	}
	h.fails("nothing to change", "knowledge", "update", "components/export.md", "--json")
	h.fails("the body is empty", "knowledge", "update", "components/export.md", "--body", "  ", "--json")
	if !strings.Contains(entry("components/export.md"), "Rewritten.") {
		t.Fatal("an empty body was written")
	}
	h.fails("E_NOT_FOUND", "knowledge", "update", "components/missing.md", "--title", "x", "--json")
	h.fails("not a valid entry path", "knowledge", "update", "../escape.md", "--title", "x", "--json")

	// Confirm needs git and a scope; --drifted confirms what drifted.
	h.fails("git repository", "knowledge", "confirm", "components/export.md", "--json")
	git("init", "-q")
	git("add", "-A")
	git("commit", "-qm", "init")
	h.fails("has no scope", "knowledge", "confirm", "overview.md", "--json")
	h.ok("knowledge", "confirm", "components/export.md")
	if !strings.Contains(entry("components/export.md"), "confirmed_commit: ") {
		t.Fatal("confirm did not set confirmed_commit")
	}
	git("commit", "-qam", "confirm")
	if err := os.MkdirAll(filepath.Join(h.dir, "internal", "export"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(h.dir, "internal", "export", "csv.go"), []byte("package export\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-qm", "code")
	if out := h.ok("check"); !strings.Contains(out, "K005") {
		t.Fatalf("expected drift:\n%s", out)
	}
	var r struct {
		Entries []string `json:"entries"`
	}
	h.jsonOf(&r, "knowledge", "confirm", "--drifted")
	if len(r.Entries) != 1 || r.Entries[0] != "/components/export.md" {
		t.Fatalf("confirm --drifted = %v", r.Entries)
	}
	if out := h.ok("check"); strings.Contains(out, "K005") {
		t.Fatalf("drift not cleared:\n%s", out)
	}

	// Clearing the scope also drops confirmed_commit.
	h.ok("knowledge", "update", "components/export.md", "--scope", "")
	if got := entry("components/export.md"); strings.Contains(got, "scope:") || strings.Contains(got, "confirmed_commit") {
		t.Fatalf("scope not cleared:\n%s", got)
	}
}

func TestTags(t *testing.T) {
	h := newHarness(t)
	a := h.newIssue("--title", "Export", "--kind", "code", "--body", "Export rows.", "--tag", "Release", "--tag", "csv,io")
	b := h.newIssue("--title", "Import", "--kind", "code", "--body", "Import rows.")
	if got := h.read(a, "issue.md"); !strings.Contains(got, "tags:\n  - release\n  - csv\n  - io\n") {
		t.Fatalf("tags not written lowercased in order:\n%s", got)
	}
	h.fails("G_TAGS", "new", "--title", "x", "--kind", "code", "--tag", "two words", "--json")

	if got := listIDs(h, "list", "--tag", "csv"); len(got) != 1 || got[0] != a {
		t.Fatalf("list --tag csv = %v", got)
	}
	if got := listIDs(h, "list", "--tag", "nope,io"); len(got) != 1 {
		t.Fatalf("comma-separated --tag should OR: %v", got)
	}
	if out := h.ok("list"); !strings.Contains(out, "Export  #release #csv #io") {
		t.Fatalf("list does not show tags:\n%s", out)
	}
	if out := h.ok("show", a); !strings.Contains(out, "tags: release, csv, io") {
		t.Fatalf("show does not show tags:\n%s", out)
	}

	// Tags are not part of the requirement: changing them never makes an issue stale.
	h.ok("define", a)
	h.ok("edit", a, "--tag", "export")
	h.expectState(a, "defined", false)
	if got := listIDs(h, "list", "--tag", "export"); len(got) != 1 || got[0] != a {
		t.Fatalf("edit did not replace tags: %v", got)
	}
	h.ok("edit", a, "--tag", "")
	if strings.Contains(h.read(a, "issue.md"), "tags:") {
		t.Fatalf("--tag '' did not clear:\n%s", h.read(a, "issue.md"))
	}

	// Tags can change on resolved issues; nothing else can.
	h.ok("drop", b, "--reason", "not needed")
	h.ok("edit", b, "--tag", "archive")
	h.fails("only --tag works", "edit", b, "--title", "x", "--json")

	// Hand edits with malformed tags are reported.
	h.replace(b, "issue.md", "  - archive", "  - Not Valid")
	h.fails("I026", "check")
}

func TestBootstrap(t *testing.T) {
	h := &harness{t: t, dir: t.TempDir(), now: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)}
	clock = func() time.Time { h.now = h.now.Add(time.Second); return h.now }
	t.Cleanup(func() { clock = time.Now })
	h.ok("init")

	if _, err := os.Stat(filepath.Join(h.dir, ".prep", "knowledge", "overview.md")); err == nil {
		t.Fatal("init still writes a placeholder overview")
	}
	var issues []summary
	var r struct {
		Issues []summary `json:"issues"`
	}
	h.jsonOf(&r, "list", "--tag", "bootstrap")
	issues = r.Issues
	if len(issues) != 2 || issues[0].Title != "Bootstrap the knowledge base" || issues[1].Kind != "research" || issues[1].Parent != issues[0].ID {
		t.Fatalf("bootstrap issues = %+v", issues)
	}
	survey := issues[1].ID
	if ctx := h.read(survey, "context.md"); !strings.Contains(ctx, "prep findings "+survey) || !strings.Contains(ctx, "--parent "+issues[0].ID) {
		t.Fatalf("survey context does not name its issues:\n%s", ctx)
	}
	if acc := h.read(survey, "acceptance.md"); !strings.Contains(acc, "/overview.md exists as a draft") {
		t.Fatalf("survey criteria missing:\n%s", acc)
	}

	// Not bootstrapped: prime and guide alert, knowledge new waits for the overview.
	if out := h.ok("prime"); !strings.HasPrefix(out, "! The knowledge base is not bootstrapped yet") || !strings.Contains(out, "prep guide "+survey) {
		t.Fatalf("prime does not lead with the bootstrap:\n%s", out)
	}
	if out := h.ok("guide", issues[0].ID); !strings.Contains(out, "not bootstrapped") {
		t.Fatalf("guide lacks the alert:\n%s", out)
	}
	h.fails("G_BOOTSTRAP", "knowledge", "new", "/components/export.md", "--type", "component", "--title", "Export", "--description", "d", "--body", "b", "--json")
	h.fails("bootstrap issues already exist", "knowledge", "bootstrap", "--json")

	// Writing the overview bootstraps the project.
	h.ok("knowledge", "new", "/overview.md", "--type", "overview", "--title", "Demo", "--description", "A demo project.", "--status", "draft", "--body", "What the demo is.")
	if out := h.ok("prime"); strings.Contains(out, "not bootstrapped") {
		t.Fatalf("prime still alerts after the overview:\n%s", out)
	}
	h.ok("knowledge", "new", "/components/export.md", "--type", "component", "--title", "Export", "--description", "d", "--body", "b")
	h.fails("already bootstrapped", "knowledge", "bootstrap", "--json")

	// Without the issues, the alert points at prep knowledge bootstrap.
	g := newHarnessNoBootstrap(t)
	if out := g.ok("prime"); !strings.Contains(out, "prep knowledge bootstrap") {
		t.Fatalf("prime without bootstrap issues:\n%s", out)
	}
	g.ok("knowledge", "bootstrap")
	if got := listIDs(g, "list", "--tag", "bootstrap"); len(got) != 2 {
		t.Fatalf("knowledge bootstrap created %v", got)
	}
}

func newHarnessNoBootstrap(t *testing.T) *harness {
	h := &harness{t: t, dir: t.TempDir(), now: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)}
	clock = func() time.Time { h.now = h.now.Add(time.Second); return h.now }
	h.ok("init", "--no-bootstrap")
	return h
}

func TestUpdateRefusesDevelopmentBuilds(t *testing.T) {
	h := newHarness(t)
	old := Version
	Version = "6449764-dirty"
	t.Cleanup(func() { Version = old })
	h.fails("development build", "update", "--json")
}
