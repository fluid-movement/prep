package claudecode

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

func TestParsePrepSession(t *testing.T) {
	f, err := os.Open("testdata/prep-session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	tr, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(tr)
	got := buf.Bytes()
	if *update {
		os.WriteFile("testdata/prep-session.golden.json", got, 0o644)
	}
	want, err := os.ReadFile("testdata/prep-session.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("transcript differs from golden; run with -update\n%s", got)
	}
	// The things the golden file is for.
	if len(tr.Steps) != 3 || tr.Steps[0].Commands[0].Name != "guide" || !tr.Steps[2].Error {
		t.Errorf("steps: %+v", tr.Steps)
	}
	if len(tr.Steps[2].Commands) != 2 || tr.Steps[2].Commands[0].Project != "api" || tr.Steps[2].Commands[1].Name != "context" {
		t.Errorf("chained commands: %+v", tr.Steps[2].Commands)
	}
	if tr.Steps[1].Path != "/work/shop/.prep/issues/01ABC/issue.md" || tr.Steps[1].Result != "# Cart page" {
		t.Errorf("read step: %+v", tr.Steps[1])
	}
	if tr.Steps[0].Before != "I'll look at the guide first." {
		t.Errorf("before: %q", tr.Steps[0].Before)
	}
	m := tr.Meta
	if m.Model != "claude-haiku-5-5" || m.Version != "2.1.296" || len(m.Plugins) != 1 || m.Plugins[0] != "prep" || len(m.Hooks) != 1 || len(m.Skills) != 1 {
		t.Errorf("meta: %+v", m)
	}
	c := tr.Counts
	if c.ToolCalls != 3 || c.Errors != 1 || c.Turns != 3 || c.InputTokens != 12 || c.OutputTokens != 55 || c.CacheRead != 200 || c.CacheWrite != 100 || c.CostUSD != 0.0123 || c.DurationMS != 4200 {
		t.Errorf("counts: %+v", c)
	}
	if tr.Final.Text != "Could not find it." || tr.Final.Status != "success" {
		t.Errorf("final: %+v", tr.Final)
	}
	if len(m.Warnings) != 1 || !strings.Contains(m.Warnings[0], "line 13") {
		t.Errorf("warnings: %v", m.Warnings)
	}
}

func TestParseTolerant(t *testing.T) {
	// no result event, a result for an unknown call, a very long line, an empty stream
	long := `{"type":"assistant","message":{"content":[{"type":"text","text":"` + strings.Repeat("x", 200000) + `"}]}}` + "\n"
	in := long + `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"zz","content":"x"}]}}` + "\n"
	tr, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(tr.Final.Text) != 200000 || len(tr.Meta.Warnings) != 1 {
		t.Errorf("final %d warnings %v", len(tr.Final.Text), tr.Meta.Warnings)
	}
	empty, err := Parse(strings.NewReader(""))
	if err != nil || len(empty.Steps) != 0 {
		t.Errorf("empty: %v %+v", err, empty)
	}
}

func TestSubagentSteps(t *testing.T) {
	in := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"a1","name":"Agent","input":{"prompt":"look"}}]}}
{"type":"assistant","parent_tool_use_id":"a1","message":{"content":[{"type":"tool_use","id":"b1","name":"Bash","input":{"command":"cat .prep/x"}}]}}
{"type":"user","parent_tool_use_id":"a1","message":{"content":[{"type":"tool_result","tool_use_id":"b1","content":"x"}]}}
`
	tr, err := Parse(strings.NewReader(in))
	if err != nil || len(tr.Steps) != 2 || tr.Steps[0].Parent != "" || tr.Steps[1].Parent != "a1" || tr.Steps[1].Result != "x" {
		t.Errorf("%v %+v", err, tr.Steps)
	}
}
