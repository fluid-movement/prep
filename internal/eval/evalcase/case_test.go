package evalcase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
)

type dummy struct{ text string }

func (d dummy) Kind() string                         { return "dummy" }
func (d dummy) Describe() string                     { return d.text }
func (d dummy) Check(in *assert.Input) assert.Result { return assert.Pass(d, "", -1) }

func init() {
	dec := func(n *yaml.Node) (assert.Assertion, error) { return dummy{n.Value}, nil }
	assert.Register(assert.Tools, "dummy", dec)
	assert.Register(assert.Outcome, "dummy", dec)
}

const shopFixture = `
name: shop
issues:
  - {alias: cart, title: Cart page}
`

const good = `
name: define-cart
area: define
tags: [smoke]
prompt: "Define {{cart}} please."
fixture: shop
limits: {timeout: 2m}
expect:
  tools:
    - dummy: guide
  outcome:
    - dummy: "{{cart}} is defined"
  rubric:
    - "The requirement is clear"
`

func fixtures(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "shop.yaml"), []byte(shopFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestParseGood(t *testing.T) {
	c, err := Parse(strings.NewReader(good), "good.yaml", fixtures(t))
	if err != nil {
		t.Fatal(err)
	}
	if c.Limits.Timeout.String() != "2m0s" || c.Limits.MaxTurns != Default.MaxTurns || c.Limits.MaxBudgetUSD != Default.MaxBudgetUSD {
		t.Errorf("limits: %+v", c.Limits)
	}
	comp, err := c.Compile(func(s string) (string, error) { return strings.ReplaceAll(s, "{{cart}}", "01XYZ"), nil })
	if err != nil {
		t.Fatal(err)
	}
	if comp.Prompt != "Define 01XYZ please." || len(comp.Tools) != 1 || comp.Outcome[0].Describe() != "01XYZ is defined" || len(comp.Rubric) != 1 {
		t.Errorf("compiled: %+v", comp)
	}
}

func TestParseErrors(t *testing.T) {
	fx := fixtures(t)
	for name, tc := range map[string]struct{ src, want string }{
		"unknown key":     {"name: a\nprompt: p\nfixture: shop\nbogus: 1\n", "bogus"},
		"no name":         {"prompt: p\nfixture: shop\n", "no name"},
		"no prompt":       {"name: a\nfixture: shop\n", "no prompt"},
		"no fixture":      {"name: a\nprompt: p\n", "exactly one"},
		"two fixtures":    {"name: a\nprompt: p\nfixture: shop\nspec: {issues: []}\n", "exactly one"},
		"missing fixture": {"name: a\nprompt: p\nfixture: nope\n", "nope"},
		"unknown alias":   {"name: a\nprompt: 'do {{ghost}}'\nfixture: shop\n", "ghost"},
		"alias in assert": {"name: a\nprompt: p\nfixture: shop\nexpect:\n  outcome:\n    - dummy: '{{ghost}}'\n", "ghost"},
		"unknown kind":    {"name: a\nprompt: p\nfixture: shop\nexpect:\n  tools:\n    - nonsense: x\n", "nonsense"},
		"two keys":        {"name: a\nprompt: p\nfixture: shop\nexpect:\n  tools:\n    - {dummy: a, dummy2: b}\n", "one key"},
		"not a list":      {"name: a\nprompt: p\nfixture: shop\nexpect:\n  tools: dummy\n", "list"},
		"bad primary":     {"name: a\nprompt: p\nfixture: shop\nprimary: zzz\n", "primary"},
		"bad inline spec": {"name: a\nprompt: p\nspec: {issues: [{kind: code}]}\n", "no title"},
		"bad duration":    {"name: a\nprompt: p\nfixture: shop\nlimits: {timeout: soon}\n", "duration"},
	} {
		_, err := Parse(strings.NewReader(tc.src), "x.yaml", fx)
		if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "x.yaml") {
			t.Errorf("%s: error %v, want mention of %q and the file", name, err, tc.want)
		}
	}
}

func TestInlineSpecAndWorkspaceAliases(t *testing.T) {
	src := `
name: ws
prompt: "Look at {{api/auth}} and {{login}}"
workspace_spec:
  primary: web
  projects:
    api:
      issues: [{alias: auth, title: Auth}]
    web:
      links: [api]
      issues: [{alias: login, title: Login}]
`
	if _, err := Parse(strings.NewReader(src), "ws.yaml", ""); err != nil {
		t.Fatal(err)
	}
	bad := strings.Replace(src, "{{login}}", "{{auth}}", 1) // auth is not in the primary project
	if _, err := Parse(strings.NewReader(bad), "ws.yaml", ""); err == nil {
		t.Error("alias of a non-primary project accepted unqualified")
	}
}

func TestLoadAndSelect(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "fixtures"), 0o755)
	os.WriteFile(filepath.Join(root, "fixtures", "shop.yaml"), []byte(shopFixture), 0o644)
	write := func(rel, src string) {
		p := filepath.Join(root, "cases", rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(src), 0o644)
	}
	write("define/a.yaml", "name: a\nprompt: p\nfixture: shop\ntags: [smoke]\n")
	write("define/b.yaml", "name: b-long\nprompt: p\nfixture: shop\n")
	write("enrich/c.yaml", "name: c\narea: enrich\nprompt: p\nfixture: shop\n")
	cases, err := Load(root)
	if err != nil || len(cases) != 3 {
		t.Fatalf("load: %v %d", err, len(cases))
	}
	if cases[0].Area != "define" {
		t.Errorf("area from directory: %q", cases[0].Area)
	}
	names := func(cs []*Case) string {
		var n []string
		for _, c := range cs {
			n = append(n, c.Name)
		}
		return strings.Join(n, ",")
	}
	for _, tc := range []struct {
		f    Filter
		want string
	}{
		{Filter{}, "a,b-long,c"},
		{Filter{Areas: []string{"define"}}, "a,b-long"},
		{Filter{Names: []string{"b-*"}}, "b-long"},
		{Filter{Tags: []string{"smoke"}}, "a"},
		{Filter{Areas: []string{"enrich"}, Tags: []string{"smoke"}}, ""},
	} {
		if got := names(Select(cases, tc.f)); got != tc.want {
			t.Errorf("select %+v = %q, want %q", tc.f, got, tc.want)
		}
	}
	write("enrich/dup.yaml", "name: a\nprompt: p\nfixture: shop\n")
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "also used") {
		t.Errorf("duplicate name: %v", err)
	}
}
