package judge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

func replies(rs ...string) (Caller, *[]string) {
	var prompts []string
	i := 0
	return func(ctx context.Context, p string) (string, error) {
		prompts = append(prompts, p)
		if i >= len(rs) {
			return "", errors.New("no more replies")
		}
		r := rs[i]
		i++
		if r == "ERR" {
			return "", errors.New("model down")
		}
		return r, nil
	}, &prompts
}

func TestParseVerdict(t *testing.T) {
	for in, want := range map[string]string{
		`{"verdict":"pass","reason":"ok","evidence":"step 1"}`:            "pass",
		"```json\n{\"verdict\": \"FAIL\", \"reason\": \"x\"}\n```":        "fail",
		`Here you go: {"verdict":"pass","reason":"fine"} hope that helps`: "pass",
	} {
		v, err := ParseVerdict(in)
		if err != nil || v.Verdict != want {
			t.Errorf("%q: %+v %v", in, v, err)
		}
	}
	for _, in := range []string{"", "no json", `{"verdict":"maybe"}`, `{"verdict":`, `{"reason":"x"}`} {
		if _, err := ParseVerdict(in); err == nil {
			t.Errorf("%q: no error", in)
		}
	}
}

func TestGradeRetriesOnceThenErrors(t *testing.T) {
	call, prompts := replies("not json", `{"verdict":"pass","reason":"clear","evidence":"step 2"}`)
	res := New(call).Grade(context.Background(), Input{Task: "t", Transcript: "tr", Diff: "d", Final: "f"}, "the text is clear")
	if !res.Pass || res.Errored || !strings.Contains(res.Reason, "clear") || !strings.Contains(res.Reason, "step 2") {
		t.Errorf("%+v", res)
	}
	if len(*prompts) != 2 || !strings.Contains((*prompts)[1], "previous reply was not") {
		t.Errorf("retry prompt: %v", *prompts)
	}
	p := (*prompts)[0]
	for _, want := range []string{"the text is clear", "tr", "## Change to the project", "f"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	call, _ = replies("garbage", "more garbage")
	res = New(call).Grade(context.Background(), Input{}, "x")
	if res.Pass || !res.Errored || !strings.Contains(res.Reason, "no usable verdict") {
		t.Errorf("two bad replies: %+v", res)
	}
	call, _ = replies("ERR", "ERR")
	if res := New(call).Grade(context.Background(), Input{}, "x"); res.Pass || !res.Errored {
		t.Errorf("model errors: %+v", res)
	}
	call, _ = replies(`{"verdict":"fail","reason":"vague"}`)
	if res := New(call).Grade(context.Background(), Input{}, "x"); res.Pass || res.Errored {
		t.Errorf("fail verdict: %+v", res)
	}
}

func TestItemsAreJudgedSeparately(t *testing.T) {
	var seen []string
	j := New(func(ctx context.Context, p string) (string, error) {
		if strings.Contains(p, "## Criterion\n\nfirst") {
			seen = append(seen, "first")
			return `{"verdict":"pass","reason":"a"}`, nil
		}
		seen = append(seen, "second")
		return `{"verdict":"fail","reason":"b"}`, nil
	})
	out := make([]bool, 0)
	for _, item := range []string{"first", "second"} {
		out = append(out, j.Grade(context.Background(), Input{}, item).Pass)
	}
	if len(seen) != 2 || !out[0] || out[1] {
		t.Errorf("%v %v", seen, out)
	}
}

func TestCalibration(t *testing.T) {
	samples, err := LoadSamples("../../../eval/judge/calibration")
	if err != nil || len(samples) != 6 {
		t.Fatalf("seed samples: %v %d", err, len(samples))
	}
	// A judge that follows the label is trusted; one that always passes is not.
	byRubricLabel := func(ctx context.Context, p string) (string, error) {
		for _, s := range samples {
			if strings.Contains(p, s.Rubric) && strings.Contains(p, s.Transcript) {
				return `{"verdict":"` + s.Verdict + `","reason":"label"}`, nil
			}
		}
		return "", errors.New("unknown sample")
	}
	good := New(byRubricLabel)
	rep, err := good.Calibrate(context.Background(), samples, 2, 0)
	if err != nil || rep.Agreement != 1 || rep.Consistency != 1 || !rep.Trusted || rep.Judgements != 12 {
		t.Errorf("good judge: %+v %v", rep, err)
	}
	if good.Trusted == nil || !*good.Trusted {
		t.Error("trusted flag not set")
	}
	always := New(func(ctx context.Context, p string) (string, error) { return `{"verdict":"pass","reason":"fine"}`, nil })
	rep, _ = always.Calibrate(context.Background(), samples, 1, 0)
	if rep.Agreement != 0.5 || rep.Trusted || len(rep.Disagreements) != 3 {
		t.Errorf("always-pass judge: %+v", rep)
	}
	if res := always.Grade(context.Background(), Input{}, "x"); !res.Untrusted {
		t.Error("results of an untrusted judge are not marked")
	}
	if res := New(always.call).Grade(context.Background(), Input{}, "x"); res.Untrusted {
		t.Error("uncalibrated judge marked untrusted")
	}
	flaky := 0
	f := New(func(ctx context.Context, p string) (string, error) {
		flaky++
		return map[bool]string{true: `{"verdict":"pass"}`, false: `{"verdict":"fail"}`}[flaky%2 == 0], nil
	})
	rep, _ = f.Calibrate(context.Background(), samples[:1], 4, 0)
	if rep.Consistency != 0 {
		t.Errorf("inconsistent judge: %+v", rep)
	}
	if _, err := good.Calibrate(context.Background(), nil, 1, 0); err == nil {
		t.Error("no samples accepted")
	}
}

func TestLoadSamplesErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("- {name: x, rubric: r, verdict: maybe}\n"), 0o644)
	if _, err := LoadSamples(dir); err == nil {
		t.Error("bad verdict accepted")
	}
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: x\nrubric: r\nverdict: pass\nbogus: 1\n"), 0o644)
	if _, err := LoadSamples(dir); err == nil {
		t.Error("unknown key accepted")
	}
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: x\nrubric: r\nverdict: pass\n"), 0o644)
	if s, err := LoadSamples(dir); err != nil || len(s) != 1 {
		t.Errorf("single sample file: %v %v", s, err)
	}
}

func TestCondense(t *testing.T) {
	tr := &transcript.Transcript{}
	for i := 0; i < 50; i++ {
		tr.Steps = append(tr.Steps, transcript.Step{Index: i, Tool: "Bash", Shell: "prep show " + strings.Repeat("x", 200), Result: strings.Repeat("y", 2000), Error: i == 3, Before: "thinking"})
	}
	out := Condense(tr, 5000)
	if len(out) > 5100 || !strings.Contains(out, "(FAILED)") || !strings.Contains(out, "[0] Bash") {
		t.Errorf("condensed %d chars", len(out))
	}
	if Condense(nil, 100) != "" {
		t.Error("nil transcript")
	}
}

func TestClaudeBackedJudge(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, "claude")
	os.WriteFile(claude, []byte(`#!/bin/sh
printf '%s\n' "$@" > "$CLAUDE_CONFIG_DIR/args.txt"
cat > "$CLAUDE_CONFIG_DIR/stdin.txt"
echo '{"type":"result","is_error":false,"result":"{\"verdict\":\"pass\",\"reason\":\"ok\"}"}'
`), 0o755)
	j, cleanup, err := NewClaude(Config{Model: "sonnet", Claude: claude, Auth: env.Auth{APIKey: "k"}, Parent: root})
	if err != nil {
		t.Fatal(err)
	}
	res := j.Grade(context.Background(), Input{Final: "done"}, "the message is accurate")
	if !res.Pass || res.Errored {
		t.Errorf("%+v", res)
	}
	var args []byte
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && d.Name() == "args.txt" {
			args, _ = os.ReadFile(p)
		}
		return nil
	})
	for _, want := range []string{"--model\nsonnet", "--output-format\njson", "--tools\n\n", "--no-session-persistence", "--disable-slash-commands"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("argument %q missing:\n%s", want, args)
		}
	}
	cleanup()
	if _, _, err := NewClaude(Config{Claude: claude}); err != env.ErrNoAuth {
		t.Errorf("no auth: %v", err)
	}
}
