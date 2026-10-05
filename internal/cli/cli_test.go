package cli

import (
	"bytes"
	"encoding/json"
	"os"
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
	h.ok("init")
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
