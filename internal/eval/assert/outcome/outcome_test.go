package outcome

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/fixture"
)

var prepBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "outcome-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	prepBin = filepath.Join(dir, "prep")
	if out, err := exec.Command("go", "build", "-o", prepBin, "../../../../cmd/prep").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build prep: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

const spec = `
name: shop
repo:
  files: {main.go: "package main\n", docs/a.md: "a\n"}
issues:
  - {alias: epic, title: Checkout epic, kind: manual, state: ready, criteria: [Shipped]}
  - {alias: cart, title: Cart page, parent: epic, state: ready}
  - {alias: pay, title: Payment, parent: epic, state: ready}
  - {alias: tax, title: Tax rules}
knowledge:
  - {path: /components/cart.md, type: component, title: Cart, description: The cart, body: Cart text.}
`

type world struct {
	ws *fixture.Workspace
	p  *fixture.Project
	in *assert.Input
}

// run builds the fixture, lets act change it through prep as an agent
// would, and prepares the assertion input.
func run(t *testing.T, act func(p *fixture.Project)) *world {
	t.Helper()
	s, err := fixture.LoadWorkspace(strings.NewReader("projects:\n  shop:\n" + indent(spec)))
	if err != nil {
		t.Fatal(err)
	}
	ws, err := fixture.BuildWorkspace(s, t.TempDir(), fixture.Options{Binary: prepBin})
	if err != nil {
		t.Fatal(err)
	}
	p := ws.Projects["shop"]
	before, err := assert.Capture(ws)
	if err != nil {
		t.Fatal(err)
	}
	act(p)
	after, err := assert.Capture(ws)
	if err != nil {
		t.Fatal(err)
	}
	return &world{ws, p, &assert.Input{Workspace: ws, Project: p, Before: before, After: after, Diff: assert.Compare(ws, before, after)}}
}

func indent(s string) string {
	var out []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		out = append(out, "    "+l)
	}
	return strings.Join(out, "\n") + "\n"
}

func (w *world) check(t *testing.T, src string) assert.Result {
	t.Helper()
	src, err := w.p.Expand(src)
	if err != nil {
		t.Fatal(err)
	}
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(src), &n); err != nil {
		t.Fatal(err)
	}
	a, err := assert.Decode(assert.Outcome, n.Content[0])
	if err != nil {
		t.Fatalf("%s: %v", src, err)
	}
	return a.Check(w.in)
}

func must(t *testing.T, _ string, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestOutcomeKinds(t *testing.T) {
	w := run(t, func(p *fixture.Project) {
		out, err := p.Prep("new", "--json", "--title", "Refund flow", "--kind", "code", "--parent", p.IDs["epic"], "--depends-on", p.IDs["pay"], "--tag", "payments", "--priority", "high", "--body", "Refunds for orders.")
		must(t, out, err)
		out, err = p.Prep("criterion", p.IDs["pay"], "--add", "Refunds work")
		must(t, out, err)
		// Criteria are contract requirements. Reconcile the intentional change
		// before asserting that the resulting project has no warnings.
		out, err = p.Prep("policy", "adopt", p.IDs["pay"])
		must(t, out, err)
		out, err = p.Prep("claim", p.IDs["cart"])
		must(t, out, err)
		out, err = p.Prep("knowledge", "new", "/components/refunds.md", "--type", "component", "--title", "Refunds", "--description", "Refund rules", "--body", "Refunds take 5 days.")
		must(t, out, err)
		must(t, "", os.WriteFile(filepath.Join(p.Dir, "main.go"), []byte("package main // changed\n"), 0o644))
		out, err = p.Git("commit", "-qam", "Change main")
		must(t, out, err)
	})
	cases := []struct {
		src  string
		pass bool
		hint string
	}{
		{"issue: {title: refund}", true, ""},
		{"issue: {title: refund, state: open, kind: code, parent: '{{epic}}', depends_on: ['{{pay}}'], tags: [payments], priority: high, requirement: refunds for}", true, ""},
		{"issue: {title: refund, state: ready}", false, `state is "open"`},
		{"issue: {title: refund, tags: [ui]}", false, `lacks tag "ui"`},
		{"issue: {title: refund, parent: '{{cart}}'}", false, "parent is"},
		{"issue: {title: refund, depends_on: ['{{tax}}']}", false, "does not depend on"},
		{"issue: {title: nonsense}", false, "new issues:"},
		{"issue: {title: nonsense, exists: false}", true, ""},
		{"issue: {title: refund, exists: false}", false, "expected no new issue"},
		{"issue: {ref: '{{cart}}', state: in-progress}", true, ""},
		{"issue: {ref: '{{pay}}', criteria: {min: 2, contains: [refunds work]}}", true, ""},
		{"issue: {ref: '{{pay}}', criteria: {count: 5}}", false, "want 5"},
		{"issue: {ref: '01XXXXXXXXXXXXXXXXXXXXXXXX'}", false, "does not exist"},
		{"claimed: '{{cart}}'", true, ""},
		{"claimed: '{{pay}}'", false, "not claimed"},
		{"claimed: {ref: '{{pay}}', claimed: false}", true, ""},
		{"claimed: {ref: '{{cart}}', by: '^nobody$'}", false, "claimed by"},
		{"unchanged: ['{{tax}}', '{{epic}}']", true, ""},
		{"unchanged: ['{{pay}}']", false, "criteria"},
		{"unchanged: {all: true}", false, "created"},
		{"unchanged: {all: true, except: ['{{pay}}', '{{cart}}']}", false, "was created"},
		{"knowledge: /components/refunds.md", true, ""},
		{"knowledge: {path: /components/refunds.md, contains: ['5 days'], type: component, created: true}", true, ""},
		{"knowledge: {path: /components/cart.md, created: true}", false, "created=true"},
		{"knowledge: {path: /components/nope.md}", false, "entries:"},
		{"knowledge: {path: /components/nope.md, exists: false}", true, ""},
		{"knowledge: {path: /components/refunds.md, not_contains: ['5 days']}", false, "contains"},
		{"check: clean", true, ""},
		{"created: {issues: {min: 1, max: 1}, entries: {min: 1}}", true, ""},
		{"created: {issues: {min: 2}}", false, "1 new issues, want at least 2"},
		{"created: {entries: {max: 0}}", false, "1 new knowledge entries, want at most 0"},
		{"created: {issues: {max: 0}}", false, "at most 0"},
		{"issue: {ref: '{{cart}}', state: 'ready|in-progress'}", true, ""},
		{"issue: {ref: '{{cart}}', state: 'ready|done'}", false, "state is"},
		{"issue: {ref: '{{pay}}', stale: false}", true, ""},
		{"issue: {ref: '{{pay}}', stale: true}", false, "stale is false"},
		{"check: {warnings: 0}", true, ""},
		{"repo: {commit: true, changed: ['main.go'], untouched: ['docs/**']}", true, ""},
		{"repo: {commit: false}", false, "new commit"},
		{"repo: {changed: ['docs/**']}", false, "no change to docs"},
	}
	for _, c := range cases {
		r := w.check(t, c.src)
		if r.Pass != c.pass {
			t.Errorf("%s: pass=%v, want %v (%s)", c.src, r.Pass, c.pass, r.Reason)
		}
		if c.hint != "" && !strings.Contains(r.Reason, c.hint) {
			t.Errorf("%s: reason %q lacks %q", c.src, r.Reason, c.hint)
		}
	}
	if det := w.in.Diff.Detail(w.in.After, 200); !strings.Contains(det, "Refunds for orders.") || !strings.Contains(det, "Refunds take 5 days.") || !strings.Contains(det, "Refunds work") {
		t.Errorf("detail lacks the text the agent wrote:\n%s", det)
	}
	sum := w.in.Diff.Summary(w.in.After)
	for _, want := range []string{`created `, `"Refund flow"`, "changed ", "criteria", "created knowledge entry /components/refunds.md", "new commits", "file changed"} {
		if !strings.Contains(sum, want) {
			t.Errorf("summary lacks %q:\n%s", want, sum)
		}
	}
}

func TestAmbiguousTitle(t *testing.T) {
	w := run(t, func(p *fixture.Project) {
		for _, title := range []string{"Refund flow", "Refund email"} {
			out, err := p.Prep("new", "--title", title, "--kind", "code")
			must(t, out, err)
		}
	})
	r := w.check(t, "issue: {title: refund}")
	if r.Pass || !strings.Contains(r.Reason, "ambiguous") {
		t.Errorf("%+v", r)
	}
	if r := w.check(t, "issue: {title: 'refund flow'}"); !r.Pass {
		t.Errorf("%+v", r)
	}
}

func TestNothingChanged(t *testing.T) {
	w := run(t, func(p *fixture.Project) {})
	if r := w.check(t, "unchanged: {all: true, knowledge: true}"); !r.Pass {
		t.Errorf("%+v", r)
	}
	if !w.in.Diff.Projects["shop"].Empty() {
		t.Errorf("diff not empty: %s", w.in.Diff.Summary(w.in.After))
	}
	if w.in.Diff.Summary(w.in.After) != "nothing changed\n" {
		t.Error("summary of an empty diff")
	}
}

func TestDecodeErrors(t *testing.T) {
	for _, src := range []string{
		"issue: {}",
		"issue: {ref: a, title: b}",
		"issue: {ref: a, exists: false}",
		"issue: {title: a, bogus: 1}",
		"unchanged: {}",
		"knowledge: {}",
		"claimed: {}",
		"created: {}",
		"claimed: {ref: a, by: '['}",
		"repo: {bogus: 1}",
	} {
		var n yaml.Node
		if err := yaml.Unmarshal([]byte(src), &n); err != nil {
			t.Fatal(err)
		}
		if _, err := assert.Decode(assert.Outcome, n.Content[0]); err == nil {
			t.Errorf("%s: no error", src)
		}
	}
}
