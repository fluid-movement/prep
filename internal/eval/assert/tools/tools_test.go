package tools

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// session builds a transcript from steps written as "Tool: argument"; Bash
// steps are parsed for prep commands. A step starting with "!" failed.
func session(steps ...string) *transcript.Transcript {
	t := &transcript.Transcript{}
	for i, s := range steps {
		st := transcript.Step{Index: i}
		if strings.HasPrefix(s, "!") {
			st.Error, st.Result, s = true, "it failed", s[1:]
		}
		tool, arg, _ := strings.Cut(s, ": ")
		st.Tool = tool
		if tool == "Bash" {
			st.Shell = arg
			st.Commands = transcript.ParseShell(arg)
		} else {
			st.Path = arg
		}
		t.Steps = append(t.Steps, st)
	}
	return t
}

func check(t *testing.T, tr *transcript.Transcript, src string) assert.Result {
	t.Helper()
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(src), &n); err != nil {
		t.Fatal(err)
	}
	a, err := assert.Decode(assert.Tools, n.Content[0])
	if err != nil {
		t.Fatalf("%s: %v", src, err)
	}
	return a.Check(&assert.Input{Transcript: tr})
}

func TestKinds(t *testing.T) {
	tr := session(
		"Bash: prep prime",
		"Bash: cd /w && prep guide 01A",
		"Read: /w/.prep/issues/01A/issue.md",
		"Bash: cat /w/.prep/issues/01A/context.md",
		"Bash: prep --project api list --state open",
		"!Bash: prep show nope",
		"Bash: prep edit 01A --parent 01B",
		"Edit: /w/src/main.go",
	)
	tr.Final.Text = "I defined the Cart page issue."
	cases := []struct {
		src  string
		pass bool
		hint string // text the reason must contain
	}{
		{"ran: guide", true, ""},
		{"ran: prep", false, "0 time"}, // `prep` is not a prep command name
		{"ran: {command: list, project: api}", true, ""},
		{"ran: {command: list, project: web}", false, "0 time"},
		{"ran: {command: list, args: '--state open'}", true, ""},
		{"ran: {command: guide, min: 2}", false, "at least 2"},
		{"ran: {command: guide, max: 0}", false, "at most 0"},
		{"ran: {command: complete, max: 0}", true, ""},
		{"ran: {tool: Edit, path: 'src/*.go'}", true, ""},
		{"ran: {tool: Edit, path: 'other/*.go'}", false, ""},
		{"not_ran: {tool: Read, path: '.prep/**'}", false, "step 2"},
		{"not_ran: {path: '.prep/**'}", false, "step 2"},
		{"not_ran: {tool: Bash, shell: 'cat .*\\.prep'}", false, "step 3"},
		{"not_ran: {tool: Grep}", true, ""},
		{"not_ran: knowledge", true, ""},
		{"before: {first: guide, then: edit}", true, ""},
		{"before: {first: edit, then: guide}", false, "first at step 1"},
		{"before: {first: {command: knowledge}, then: edit}", false, "never ran"},
		{"before: {first: guide, then: {command: complete}}", true, "nothing to order"},
		{"max_calls: 8", true, ""},
		{"max_calls: 7", false, "there were 8"},
		{"max_calls: {tool: Bash, max: 5}", false, "there were 6"},
		{"no_errors: true", false, "step 5"},
		{"no_errors: {except: [show]}", true, ""},
		{"args: {command: edit, matches: '--parent \\w+'}", true, ""},
		{"args: {command: edit, matches: '--tag'}", false, "the last with"},
		{"args: {command: complete, matches: 'x'}", false, "never ran"},
		{"no_prep_files: true", false, "step 2"},
		{"any_of: [{command: show}, {command: list, project: api}]", true, "at step"},
		{"any_of: [knowledge, {command: complete}]", false, "none happened"},
		{"final_mentions: cart page", true, ""},
		{"final_mentions: {regex: 'defin(ed|ing)'}", true, ""},
		{"final_mentions: 'finished'", false, "it said"},
	}
	for _, c := range cases {
		r := check(t, tr, c.src)
		if r.Pass != c.pass {
			t.Errorf("%s: pass=%v, want %v (%s)", c.src, r.Pass, c.pass, r.Reason)
		}
		if c.hint != "" && !strings.Contains(r.Reason, c.hint) {
			t.Errorf("%s: reason %q lacks %q", c.src, r.Reason, c.hint)
		}
		if r.Kind == "" || r.Desc == "" {
			t.Errorf("%s: result without kind or description: %+v", c.src, r)
		}
	}
}

func TestFailureNamesTheStep(t *testing.T) {
	tr := session("Bash: prep list", "Read: /w/.prep/x.md")
	r := check(t, tr, "not_ran: {tool: Read, path: '.prep/**'}")
	if r.Pass || r.Step != 1 || !strings.Contains(r.Reason, "step 1") {
		t.Errorf("%+v", r)
	}
}

func TestDecodeErrors(t *testing.T) {
	for _, src := range []string{
		"ran: {}",
		"ran: {command: x, bogus: 1}",
		"ran: {command: x, min: 3, max: 1}",
		"ran: {command: x, args: '['}",
		"before: {first: guide}",
		"before: guide",
		"any_of: []",
		"any_of: guide",
		"max_calls: many",
		"max_calls: {tool: Bash}",
		"args: {command: x}",
		"final_mentions: {text: a, regex: b}",
		"not_ran: [a]",
	} {
		var n yaml.Node
		if err := yaml.Unmarshal([]byte(src), &n); err != nil {
			t.Fatal(err)
		}
		if _, err := assert.Decode(assert.Tools, n.Content[0]); err == nil {
			t.Errorf("%s: no error", src)
		}
	}
}

func TestGlob(t *testing.T) {
	re, _ := globRegexp(".prep/**")
	for p, want := range map[string]bool{
		"/w/.prep/issues/a/issue.md": true, ".prep/x": true, "/w/prep/x": false, "/w/x.prep/a": false,
	} {
		if re.MatchString(p) != want {
			t.Errorf("%q: %v", p, !want)
		}
	}
}

func TestNoPrepFiles(t *testing.T) {
	for _, c := range []struct {
		steps []string
		pass  bool
	}{
		{[]string{"Bash: prep show 01A", "Read: /w/src/main.go", "Bash: ls /w"}, true},
		{[]string{"Bash: cat /w/.prep/issues/a/issue.md"}, false},
		{[]string{"Bash: ls .prep"}, false},
		{[]string{"Bash: grep -r todo .prep/"}, false},
		{[]string{"Edit: /w/.prep/issues/a/issue.md"}, false},
		{[]string{"Grep: /w/.prep"}, false},
		{[]string{"Read: /w/.prepare/x"}, true},
		{[]string{"Bash: prep edit 01A --body 'see .prep/ docs'"}, true},
		{[]string{"Bash: echo hi | tee x.prep"}, true},
		{[]string{"Bash: grep -rniE headline . -l --exclude-dir=.git --exclude-dir=.prep | head"}, true},
		{[]string{"Bash: rg headline --glob '!.prep'"}, true},
		{[]string{"Bash: find . -path ./.prep -prune -o -type f -print"}, true},
		{[]string{"Bash: grep -r headline --exclude-dir=.prep . .prep/knowledge"}, false},
	} {
		r := check(t, session(c.steps...), "no_prep_files: true")
		if r.Pass != c.pass {
			t.Errorf("%v: pass=%v (%s)", c.steps, r.Pass, r.Reason)
		}
	}
}

func TestSubagentSelector(t *testing.T) {
	tr := session("Agent: look around", "Bash: cat /w/.prep/issues/a/issue.md", "Bash: prep show 01A")
	tr.Steps[1].Parent = "a1"
	for src, pass := range map[string]bool{
		"not_ran: {subagent: true, path: '.prep/**'}":  false,
		"not_ran: {subagent: false, path: '.prep/**'}": true,
		"ran: {tool: Bash, subagent: true}":            true,
		"ran: {command: show, subagent: false}":        true,
		"no_prep_files: true":                          false,
	} {
		if r := check(t, tr, src); r.Pass != pass {
			t.Errorf("%s: %v (%s)", src, r.Pass, r.Reason)
		}
	}
}
