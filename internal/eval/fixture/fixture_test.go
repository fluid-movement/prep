package fixture

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var prepBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "fixture-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	prepBin = filepath.Join(dir, "prep")
	if out, err := exec.Command("go", "build", "-o", prepBin, "../../../cmd/prep").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build prep: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

const everyState = `
name: shop
issues:
  - {alias: epic, title: Checkout epic, kind: manual, state: ready, criteria: [Shipped]}
  - {alias: cart, title: Cart page, parent: epic, state: done, tags: [ui], priority: high}
  - {alias: pay, title: Payment, parent: epic, state: in-progress, depends_on: [cart]}
  - {alias: tax, title: Tax rules, parent: epic, state: ready, context: 'See {{cart}}'}
  - {alias: ship, title: Shipping, parent: epic, state: defined}
  - {alias: idea, title: Gift wrap}
  - {alias: old, title: Wishlist, state: dropped}
  - {alias: why, title: Why Postgres, kind: research, state: done}
  - {alias: db, title: Pick a database, kind: decision, state: done}
knowledge:
  - {path: /components/cart.md, type: component, title: Cart, description: The cart, body: "See {{cart}}."}
`

func load(t *testing.T, src string) *Spec {
	t.Helper()
	s, err := LoadSpec(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func build(t *testing.T, s *Spec) *Project {
	t.Helper()
	p, err := Build(s, t.TempDir(), Options{Binary: prepBin})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEveryState(t *testing.T) {
	p := build(t, load(t, everyState))
	want := map[string]string{
		"epic": "ready", "cart": "done", "pay": "in-progress", "tax": "ready",
		"ship": "defined", "idea": "open", "old": "dropped", "why": "done", "db": "done",
	}
	for alias, state := range want {
		out, err := p.Prep("show", p.IDs[alias], "--json")
		if err != nil {
			t.Fatal(err)
		}
		var s struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal([]byte(out), &s); err != nil {
			t.Fatal(err)
		}
		if s.State != state {
			t.Errorf("%s: state %q, want %q", alias, s.State, state)
		}
	}
	if out, err := p.Prep("check"); err != nil || !strings.Contains(out, "0 errors") {
		t.Fatalf("check: %v\n%s", err, out)
	}
	// Aliases expanded in context and knowledge.
	if out, _ := p.Prep("show", p.IDs["tax"]); !strings.Contains(out, p.IDs["cart"]) {
		t.Errorf("context alias not expanded:\n%s", out)
	}
	if out, _ := p.Prep("knowledge", "show", "/components/cart.md"); !strings.Contains(out, p.IDs["cart"]) {
		t.Errorf("knowledge alias not expanded:\n%s", out)
	}
	// Tags, priority, parent.
	if out, _ := p.Prep("show", p.IDs["cart"]); !strings.Contains(out, "ui") || !strings.Contains(out, "high") || !strings.Contains(out, p.IDs["epic"]) {
		t.Errorf("tags, priority or parent missing:\n%s", out)
	}
}

func TestSetupCallsCarryFixtureActor(t *testing.T) {
	p := build(t, load(t, everyState))
	out, err := p.Prep("activity", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"actor": "fixture"`) {
		t.Errorf("no setup event by fixture:\n%.400s", out)
	}
}

// treeHash hashes every file under dir except git internals and the
// machine-local activity stream.
func treeHash(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	var paths []string
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() && (rel == ".git" || rel == filepath.Join(".prep", "local")) {
			return fs.SkipDir
		}
		if !d.IsDir() {
			paths = append(paths, rel)
		}
		return nil
	})
	sort.Strings(paths)
	for _, rel := range paths {
		data, _ := os.ReadFile(filepath.Join(dir, rel))
		fmt.Fprintf(h, "%s\x00%s\x00", rel, data)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestDeterministic(t *testing.T) {
	a := build(t, load(t, everyState))
	b := build(t, load(t, everyState))
	if a.IDs["cart"] != b.IDs["cart"] || a.IDs["db"] != b.IDs["db"] {
		t.Errorf("IDs differ: %v vs %v", a.IDs, b.IDs)
	}
	// project.md carries the machine's directory-derived identity? Compare the whole tree.
	if ha, hb := treeHash(t, a.Dir), treeHash(t, b.Dir); ha != hb {
		t.Errorf("trees differ")
	}
}

func TestSeamsDoNothingUnset(t *testing.T) {
	newID := func(env ...string) string {
		dir := t.TempDir()
		base := append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + dir, "XDG_STATE_HOME=" + dir}, env...)
		for _, args := range [][]string{{"init"}, {"new", "--title", "x", "--kind", "code", "--json"}} {
			c := exec.Command(prepBin, append([]string{"--root", dir}, args...)...)
			c.Env = base
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			if args[0] == "new" {
				var r writeOut
				json.Unmarshal(out, &r)
				return r.ID
			}
		}
		return ""
	}
	if a, b := newID(), newID(); a == b || a == "" {
		t.Errorf("unset seams: IDs %q %q", a, b)
	}
	if a, b := newID("PREP_TEST_NOW=2026-01-05T09:00:00Z", "PREP_TEST_SEED=1"), newID("PREP_TEST_NOW=2026-01-05T09:00:00Z", "PREP_TEST_SEED=1"); a != b {
		t.Errorf("set seams: IDs differ %q %q", a, b)
	}
}

func TestRepoSection(t *testing.T) {
	s := load(t, `
name: app
repo:
  files:
    main.go: "package main\n"
    pkg/util.go: "package pkg\n"
issues:
  - {title: One}
`)
	p := build(t, s)
	if _, err := os.Stat(filepath.Join(p.Dir, "pkg", "util.go")); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("git", "-C", p.Dir, "log", "--format=%s").CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Initial commit") {
		t.Errorf("log: %s", out)
	}
	// no repo section, no git repo
	q := build(t, load(t, "name: bare\nissues: [{title: One}]\n"))
	if _, err := os.Stat(filepath.Join(q.Dir, ".git")); err == nil {
		t.Error("bare fixture has a git repo")
	}
}

func TestSpecErrors(t *testing.T) {
	for name, src := range map[string]string{
		"unknown key":    "name: x\nbogus: 1\n",
		"no title":       "issues: [{kind: code}]\n",
		"bad state":      "issues: [{title: a, state: finished}]\n",
		"bad parent":     "issues: [{title: a, parent: nope}]\n",
		"duplicate":      "issues: [{alias: a, title: a}, {alias: a, title: b}]\n",
		"bad dependency": "issues: [{title: a, depends_on: [nope]}]\n",
	} {
		if _, err := LoadSpec(strings.NewReader(src)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	p, err := Build(load(t, "issues: [{alias: a, title: a, body: 'see {{ghost}}'}]\n"), t.TempDir(), Options{Binary: prepBin})
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("unknown alias: %v %v", p, err)
	}
	_, err = Build(load(t, "issues: [{alias: p, title: p, state: in-progress}, {title: c, parent: p}]\n"), t.TempDir(), Options{Binary: prepBin})
	if err == nil {
		t.Error("in-progress parent accepted")
	}
	_, err = Build(load(t, "issues: [{alias: a, title: a, parent: b}, {alias: b, title: b, parent: a}]\n"), t.TempDir(), Options{Binary: prepBin})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("cycle: %v", err)
	}
}

func TestScopeEmptyKnowledgeAndGenerate(t *testing.T) {
	const src = `
name: big
repo:
  files: {src/a.go: "package a\n", test.sh: "#!/bin/sh\nexit 1\n"}
  executable: [test.sh]
generate: {issues: 21, knowledge: 6}
knowledge:
  - {path: /components/a.md, title: A, scope: [src/a.go], body: About a.}
`
	p, q := build(t, load(t, src)), build(t, load(t, src))
	// the same spec twice gives the same tree (compared before anything else touches either)
	if treeHash(t, p.Dir) != treeHash(t, q.Dir) {
		t.Error("generated fixture is not deterministic")
	}
	out, err := p.Prep("list", "--json")
	if err != nil || strings.Count(out, `"id"`) != 21 {
		t.Fatalf("generated issues: %v\n%.200s", err, out)
	}
	if out, _ := p.Prep("knowledge", "list"); strings.Count(out, "/components/") != 7 {
		t.Errorf("entries:\n%s", out)
	}
	if st, err := os.Stat(filepath.Join(p.Dir, "test.sh")); err != nil || st.Mode()&0o100 == 0 {
		t.Errorf("executable bit: %v %v", err, st)
	}
	if out, err := p.Prep("check"); err != nil || !strings.Contains(out, "0 errors") {
		t.Errorf("check: %v\n%s", err, out)
	}
	states := map[string]int{}
	var l struct {
		Issues []struct{ State string } `json:"issues"`
	}
	json.Unmarshal([]byte(func() string { o, _ := p.Prep("list", "--json"); return o }()), &l)
	for _, i := range l.Issues {
		states[i.State]++
	}
	for _, want := range []string{"open", "defined", "ready", "done", "dropped", "in-progress"} {
		if states[want] == 0 {
			t.Errorf("no generated issue in state %s: %v", want, states)
		}
	}
	// Compatibility fixture option leaves knowledge empty; init creates no issues.
	b := build(t, load(t, "name: fresh\nbootstrap: true\nrepo: {files: {main.go: \"package main\\n\"}}\n"))
	if out, _ := b.Prep("list"); !strings.Contains(out, "no issues") {
		t.Errorf("init created unexpected issues:\n%s", out)
	}
	if out, _ := b.Prep("knowledge", "list"); strings.Contains(out, "/overview.md") {
		t.Errorf("a bootstrapped fixture should have no overview yet:\n%s", out)
	}
}
