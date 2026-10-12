package tools

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
)

func init() {
	assert.Register(assert.Tools, "ran", decodeRan)
	assert.Register(assert.Tools, "not_ran", decodeNotRan)
	assert.Register(assert.Tools, "before", decodeBefore)
	assert.Register(assert.Tools, "max_calls", decodeMaxCalls)
	assert.Register(assert.Tools, "no_errors", decodeNoErrors)
	assert.Register(assert.Tools, "args", decodeArgs)
	assert.Register(assert.Tools, "final_mentions", decodeFinal)
	assert.Register(assert.Tools, "no_prep_files", decodeNoPrepFiles)
	assert.Register(assert.Tools, "any_of", decodeAnyOf)
}

// ---- ran

type ran struct {
	sel      Selector
	min, max int // max < 0: unbounded
}

func decodeRan(n *yaml.Node) (assert.Assertion, error) {
	r := &ran{min: 1, max: -1}
	if n.Kind == yaml.ScalarNode {
		sel, err := parseSelector(n)
		r.sel = sel
		return r, err
	}
	var raw struct {
		Selector `yaml:",inline"`
		Min      *int `yaml:"min"`
		Max      *int `yaml:"max"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	r.sel = raw.Selector
	if raw.Min != nil {
		r.min = *raw.Min
	}
	if raw.Max != nil {
		r.max = *raw.Max
		if raw.Min == nil && r.max < r.min { // {max: 0} means "at most 0"
			r.min = r.max
		}
	}
	if r.max >= 0 && r.max < r.min {
		return nil, fmt.Errorf("max %d is below min %d", r.max, r.min)
	}
	return r, r.sel.compile()
}

func (r *ran) Kind() string { return "ran" }
func (r *ran) Describe() string {
	d := "ran " + r.sel.Describe()
	switch {
	case r.max >= 0 && r.min == r.max:
		d += fmt.Sprintf(" exactly %d times", r.min)
	case r.max >= 0:
		d += fmt.Sprintf(" %d to %d times", r.min, r.max)
	case r.min > 1:
		d += fmt.Sprintf(" at least %d times", r.min)
	}
	return d
}
func (r *ran) Check(in *assert.Input) assert.Result {
	first, count := r.sel.First(in.Transcript)
	switch {
	case count < r.min:
		return assert.Fail(r, fmt.Sprintf("expected %s at least %d time(s); it ran %d time(s)", r.sel.Describe(), r.min, count), first)
	case r.max >= 0 && count > r.max:
		return assert.Fail(r, fmt.Sprintf("expected %s at most %d time(s); it ran %d time(s)", r.sel.Describe(), r.max, count), first)
	}
	return assert.Pass(r, fmt.Sprintf("ran %d time(s), first at step %d", count, first), first)
}

// ---- not_ran

type notRan struct{ sel Selector }

func decodeNotRan(n *yaml.Node) (assert.Assertion, error) {
	sel, err := parseSelector(n)
	return &notRan{sel}, err
}

func (a *notRan) Kind() string     { return "not_ran" }
func (a *notRan) Describe() string { return "never ran " + a.sel.Describe() }
func (a *notRan) Check(in *assert.Input) assert.Result {
	if first, count := a.sel.First(in.Transcript); count > 0 {
		st := in.Transcript.Steps[first]
		what := st.Shell
		if what == "" {
			what = st.Path
		}
		return assert.Fail(a, fmt.Sprintf("expected %s never to happen; it did %d time(s), first at step %d (%s %s)", a.sel.Describe(), count, first, st.Tool, clip(what)), first)
	}
	return assert.Pass(a, "never happened", -1)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 100 {
		return s[:100] + "…"
	}
	return s
}

// ---- before

type before struct{ first, then Selector }

func decodeBefore(n *yaml.Node) (assert.Assertion, error) {
	if n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("line %d: before needs {first: ..., then: ...}", n.Line)
	}
	var raw struct {
		First yaml.Node `yaml:"first"`
		Then  yaml.Node `yaml:"then"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	if raw.First.Kind == 0 || raw.Then.Kind == 0 {
		return nil, fmt.Errorf("line %d: before needs both first and then", n.Line)
	}
	f, err := parseSelector(&raw.First)
	if err != nil {
		return nil, fmt.Errorf("first: %w", err)
	}
	t, err := parseSelector(&raw.Then)
	if err != nil {
		return nil, fmt.Errorf("then: %w", err)
	}
	return &before{f, t}, nil
}

func (b *before) Kind() string { return "before" }
func (b *before) Describe() string {
	return b.first.Describe() + " before " + b.then.Describe()
}

// Check passes when the first selector's first match precedes the second's
// first match, or when the second never happens (nothing to order).
func (b *before) Check(in *assert.Input) assert.Result {
	f, _ := b.first.First(in.Transcript)
	t, tc := b.then.First(in.Transcript)
	switch {
	case tc == 0:
		return assert.Pass(b, b.then.Describe()+" never happened, so nothing to order", -1)
	case f < 0:
		return assert.Fail(b, fmt.Sprintf("expected %s before %s; %s first happened at step %d and %s never ran", b.first.Describe(), b.then.Describe(), b.then.Describe(), t, b.first.Describe()), t)
	case f > t:
		return assert.Fail(b, fmt.Sprintf("expected %s before %s; %s first at step %d, %s first at step %d", b.first.Describe(), b.then.Describe(), b.then.Describe(), t, b.first.Describe(), f), t)
	}
	return assert.Pass(b, fmt.Sprintf("%s at step %d, %s at step %d", b.first.Describe(), f, b.then.Describe(), t), f)
}

// ---- max_calls

type maxCalls struct {
	tool string
	max  int
}

func decodeMaxCalls(n *yaml.Node) (assert.Assertion, error) {
	m := &maxCalls{}
	if n.Kind == yaml.ScalarNode {
		if _, err := fmt.Sscanf(n.Value, "%d", &m.max); err != nil {
			return nil, fmt.Errorf("line %d: max_calls wants a number", n.Line)
		}
		return m, nil
	}
	var raw struct {
		Tool string `yaml:"tool"`
		Max  *int   `yaml:"max"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	if raw.Max == nil {
		return nil, fmt.Errorf("line %d: max_calls needs max", n.Line)
	}
	m.tool, m.max = raw.Tool, *raw.Max
	return m, nil
}

func (m *maxCalls) Kind() string { return "max_calls" }
func (m *maxCalls) Describe() string {
	if m.tool != "" {
		return fmt.Sprintf("at most %d %s calls", m.max, m.tool)
	}
	return fmt.Sprintf("at most %d tool calls", m.max)
}
func (m *maxCalls) Check(in *assert.Input) assert.Result {
	n := 0
	for _, s := range in.Transcript.Steps {
		if m.tool == "" || s.Tool == m.tool {
			n++
		}
	}
	if n > m.max {
		return assert.Fail(m, fmt.Sprintf("expected %s; there were %d", m.Describe(), n), -1)
	}
	return assert.Pass(m, fmt.Sprintf("%d call(s)", n), -1)
}

// ---- no_errors

type noErrors struct{ except []Selector }

func decodeNoErrors(n *yaml.Node) (assert.Assertion, error) {
	a := &noErrors{}
	if n.Kind == yaml.ScalarNode {
		return a, nil
	}
	var raw struct {
		Except []yaml.Node `yaml:"except"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	for i := range raw.Except {
		s, err := parseSelector(&raw.Except[i])
		if err != nil {
			return nil, err
		}
		a.except = append(a.except, s)
	}
	return a, nil
}

func (a *noErrors) Kind() string { return "no_errors" }
func (a *noErrors) Describe() string {
	if len(a.except) == 0 {
		return "no tool call failed"
	}
	var e []string
	for _, s := range a.except {
		e = append(e, s.Describe())
	}
	return "no tool call failed except " + strings.Join(e, ", ")
}
func (a *noErrors) Check(in *assert.Input) assert.Result {
	for _, st := range in.Transcript.Steps {
		if !st.Error {
			continue
		}
		excused := false
		for _, s := range a.except {
			if s.Matches(st) > 0 {
				excused = true
			}
		}
		if !excused {
			what := st.Shell
			if what == "" {
				what = st.Path
			}
			return assert.Fail(a, fmt.Sprintf("step %d failed (%s %s): %s", st.Index, st.Tool, clip(what), clip(st.Result)), st.Index)
		}
	}
	return assert.Pass(a, "no failed calls", -1)
}

// ---- args

type argsCheck struct {
	command string
	re      *regexp.Regexp
	source  string
}

func decodeArgs(n *yaml.Node) (assert.Assertion, error) {
	var raw struct {
		Command string `yaml:"command"`
		Matches string `yaml:"matches"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	if raw.Command == "" || raw.Matches == "" {
		return nil, fmt.Errorf("line %d: args needs command and matches", n.Line)
	}
	re, err := regexp.Compile(raw.Matches)
	if err != nil {
		return nil, err
	}
	return &argsCheck{raw.Command, re, raw.Matches}, nil
}

func (a *argsCheck) Kind() string { return "args" }
func (a *argsCheck) Describe() string {
	return fmt.Sprintf("a run of prep %s had arguments matching /%s/", a.command, a.source)
}

// Check passes when some run of the command has matching arguments.
func (a *argsCheck) Check(in *assert.Input) assert.Result {
	seen := 0
	var last string
	for _, st := range in.Transcript.Steps {
		for _, c := range st.Commands {
			if !c.Is(a.command) {
				continue
			}
			seen++
			joined := strings.Join(c.Args, " ")
			last = joined
			if a.re.MatchString(joined) {
				return assert.Pass(a, "arguments: "+clip(joined), st.Index)
			}
		}
	}
	if seen == 0 {
		return assert.Fail(a, fmt.Sprintf("expected a run of prep %s with arguments /%s/; it never ran", a.command, a.source), -1)
	}
	return assert.Fail(a, fmt.Sprintf("expected a run of prep %s with arguments /%s/; %d run(s), the last with %q", a.command, a.source, seen, clip(last)), -1)
}

// ---- final_mentions

type finalMentions struct {
	text string
	re   *regexp.Regexp
}

func decodeFinal(n *yaml.Node) (assert.Assertion, error) {
	f := &finalMentions{}
	if n.Kind == yaml.ScalarNode {
		f.text = n.Value
		return f, nil
	}
	var raw struct {
		Text  string `yaml:"text"`
		Regex string `yaml:"regex"`
	}
	if err := assert.DecodeStrict(n, &raw); err != nil {
		return nil, err
	}
	if (raw.Text == "") == (raw.Regex == "") {
		return nil, fmt.Errorf("line %d: final_mentions needs text or regex", n.Line)
	}
	f.text = raw.Text
	if raw.Regex != "" {
		re, err := regexp.Compile("(?i)" + raw.Regex)
		if err != nil {
			return nil, err
		}
		f.re = re
		f.text = raw.Regex
	}
	return f, nil
}

func (f *finalMentions) Kind() string { return "final_mentions" }
func (f *finalMentions) Describe() string {
	return fmt.Sprintf("the final message mentions %q", f.text)
}
func (f *finalMentions) Check(in *assert.Input) assert.Result {
	msg := in.Transcript.Final.Text
	ok := false
	if f.re != nil {
		ok = f.re.MatchString(msg)
	} else {
		ok = strings.Contains(strings.ToLower(msg), strings.ToLower(f.text))
	}
	if !ok {
		return assert.Fail(f, fmt.Sprintf("expected the final message to mention %q; it said %q", f.text, clip(msg)), -1)
	}
	return assert.Pass(f, "mentioned", -1)
}

// ---- no_prep_files

// noPrepFiles fails when the agent reached into .prep with anything but
// prep: a file tool on a path under .prep, or a shell command that is not a
// prep command and names .prep. prep guides agents to use its commands, so
// this is the rule the skill and hooks exist to protect.
type noPrepFiles struct{}

func decodeNoPrepFiles(n *yaml.Node) (assert.Assertion, error) { return noPrepFiles{}, nil }

// prepRecords matches a path into the records prep keeps and owns: issues,
// knowledge, the local activity stream. Reading project.md or config.yaml is
// not what the skill forbids.
var prepRecords = regexp.MustCompile(`\.prep/(issues|knowledge|local)(/|\s|'|"|$)`)

// prepSweep matches a command that walks the whole .prep directory.
var prepSweep = regexp.MustCompile(`\b(find|grep|rg|ag|tree|du|cat|ls\s+-\S*R)\s[^|;&]*\.prep(\s|/?$|/?['"])`)

// prepExcluded matches the ways a search says to leave .prep out, which is the
// opposite of reading it.
var prepExcluded = regexp.MustCompile(`--(?:exclude(?:-dir)?|ignore|glob)[= ]+['"]?!?\S*\.prep\S*|!\s*-path\s+\S*\.prep\S*|-not\s+-path\s+\S*\.prep\S*|-path\s+\S*\.prep\S*\s+-prune`)

func (noPrepFiles) Kind() string     { return "no_prep_files" }
func (noPrepFiles) Describe() string { return "never touched .prep files directly" }
func (a noPrepFiles) Check(in *assert.Input) assert.Result {
	for _, st := range in.Transcript.Steps {
		switch {
		case st.Tool == "Bash":
			if sh := prepExcluded.ReplaceAllString(st.Shell, ""); len(st.Commands) == 0 && (prepRecords.MatchString(sh) || prepSweep.MatchString(sh)) {
				return assert.Fail(a, fmt.Sprintf("step %d reached into .prep from the shell: %s", st.Index, clip(st.Shell)), st.Index)
			}
		case st.Path != "" && (prepRecords.MatchString(st.Path+"/") || strings.HasSuffix(st.Path, "/.prep")):
			return assert.Fail(a, fmt.Sprintf("step %d used %s on %s", st.Index, st.Tool, st.Path), st.Index)
		}
	}
	return assert.Pass(a, "no direct access", -1)
}

// ---- any_of

// anyOf passes when at least one of its selectors matches a step: several
// right ways to do the same thing.
type anyOf struct{ sels []Selector }

func decodeAnyOf(n *yaml.Node) (assert.Assertion, error) {
	if n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
		return nil, fmt.Errorf("line %d: any_of needs a list of commands or selectors", n.Line)
	}
	a := &anyOf{}
	for _, c := range n.Content {
		s, err := parseSelector(c)
		if err != nil {
			return nil, err
		}
		a.sels = append(a.sels, s)
	}
	return a, nil
}

func (a *anyOf) Kind() string { return "any_of" }
func (a *anyOf) Describe() string {
	var d []string
	for _, s := range a.sels {
		d = append(d, s.Describe())
	}
	return "ran one of: " + strings.Join(d, " | ")
}
func (a *anyOf) Check(in *assert.Input) assert.Result {
	for _, s := range a.sels {
		if first, n := s.First(in.Transcript); n > 0 {
			return assert.Pass(a, fmt.Sprintf("%s at step %d", s.Describe(), first), first)
		}
	}
	return assert.Fail(a, "expected one of "+a.Describe()[len("ran one of: "):]+"; none happened", -1)
}
