// Package judge scores what the deterministic assertions cannot: whether a
// piece of text is good, whether the guidance was followed in spirit. It is
// kept small and repeatable: a pinned model and prompt, no tools, one
// criterion per call, structured output, and a calibration set that measures
// whether its verdicts can be trusted.
package judge

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/run"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

//go:embed prompt.md
var promptTemplate string

// Caller sends a prompt to the judge model and returns its reply text.
type Caller func(ctx context.Context, prompt string) (string, error)

// Judge grades rubric items.
type Judge struct {
	call Caller
	// MaxTranscript caps the condensed transcript in the prompt.
	MaxTranscript int
	// Trusted is the outcome of the last calibration; nil means not
	// calibrated, and results are not marked.
	Trusted *bool
}

// New returns a judge that sends prompts through call.
func New(call Caller) *Judge { return &Judge{call: call, MaxTranscript: 60000} }

// Config makes a judge that runs Claude Code.
type Config struct {
	// Model is the pinned judge model.
	Model  string
	Claude string
	Auth   env.Auth
	Parent string
}

// NewClaude returns a judge backed by `claude -p` with no tools, no
// persistence and no plugins, in an isolated environment, and a function that
// removes that environment.
func NewClaude(cfg Config) (*Judge, func(), error) {
	if cfg.Model == "" {
		cfg.Model = "sonnet"
	}
	e, err := env.New(env.Config{Claude: cfg.Claude, Auth: cfg.Auth, Parent: cfg.Parent})
	if err != nil {
		return nil, nil, err
	}
	var mu sync.Mutex
	call := func(ctx context.Context, prompt string) (string, error) {
		mu.Lock() // one scratch directory; serialise starting the process only
		cmd := e.Command(e.Workspace, "-p", "--model", cfg.Model, "--output-format", "json", "--tools", "",
			"--no-session-persistence", "--disable-slash-commands", "--strict-mcp-config", "--max-turns", "1")
		mu.Unlock()
		cmd.Stdin = strings.NewReader(prompt)
		out, err := runCmd(ctx, cmd)
		if err != nil {
			return "", err
		}
		var r struct {
			Result  string `json:"result"`
			IsError bool   `json:"is_error"`
		}
		if err := json.Unmarshal(out, &r); err != nil {
			return "", fmt.Errorf("judge output: %w", err)
		}
		if r.IsError {
			return "", fmt.Errorf("judge model error: %s", r.Result)
		}
		return r.Result, nil
	}
	return New(call), e.Cleanup, nil
}

// Verdict is a parsed reply.
type Verdict struct {
	Verdict  string `json:"verdict"`
	Reason   string `json:"reason"`
	Evidence string `json:"evidence"`
}

// Input is what the judge sees.
type Input struct {
	Task       string
	Transcript string
	Diff       string
	Final      string
}

// Grade judges one rubric item, retrying once on an unusable reply. A
// failure to get a verdict is a result with Errored set, never a pass.
func (j *Judge) Grade(ctx context.Context, in Input, rubric string) assert.Result {
	res := assert.Result{Kind: "rubric", Desc: rubric, Step: -1}
	prompt := j.prompt(in, rubric)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		p := prompt
		if attempt == 1 {
			p += "\n\nYour previous reply was not a valid JSON object with a \"verdict\" of \"pass\" or \"fail\". Reply with only that object."
		}
		reply, err := j.call(ctx, p)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				break
			}
			continue
		}
		v, err := ParseVerdict(reply)
		if err != nil {
			lastErr = err
			continue
		}
		res.Pass = v.Verdict == "pass"
		res.Reason = v.Reason
		if v.Evidence != "" {
			res.Reason += " [evidence: " + v.Evidence + "]"
		}
		res.Untrusted = j.Trusted != nil && !*j.Trusted
		return res
	}
	res.Errored = true
	res.Reason = fmt.Sprintf("the judge gave no usable verdict: %v", lastErr)
	return res
}

// Func adapts the judge to the matrix: it judges every rubric item of a run
// from its transcript and state change.
func (j *Judge) Func() func(ctx context.Context, rec *run.Record, rubric []string) []assert.Result {
	return func(ctx context.Context, rec *run.Record, rubric []string) []assert.Result {
		in := Input{Task: rec.Prompt, Transcript: Condense(rec.Transcript, j.MaxTranscript), Final: rec.Transcript.Final.Text}
		if rec.Input != nil && rec.Input.Diff != nil {
			in.Diff = rec.Input.Diff.Summary(rec.Input.After)
		}
		out := make([]assert.Result, len(rubric))
		var wg sync.WaitGroup
		for i, item := range rubric {
			wg.Add(1)
			go func() {
				defer wg.Done()
				out[i] = j.Grade(ctx, in, item)
			}()
		}
		wg.Wait()
		return out
	}
}

func (j *Judge) prompt(in Input, rubric string) string {
	r := strings.NewReplacer(
		"{{rubric}}", rubric,
		"{{task}}", orNone(in.Task),
		"{{transcript}}", orNone(in.Transcript),
		"{{diff}}", orNone(in.Diff),
		"{{final}}", orNone(in.Final),
	)
	return r.Replace(promptTemplate)
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}

// ParseVerdict reads the JSON object out of a reply, tolerating a code fence
// or a sentence around it.
func ParseVerdict(reply string) (Verdict, error) {
	var v Verdict
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start < 0 || end < start {
		return v, fmt.Errorf("no JSON object in the reply")
	}
	if err := json.Unmarshal([]byte(reply[start:end+1]), &v); err != nil {
		return v, fmt.Errorf("bad JSON: %w", err)
	}
	v.Verdict = strings.ToLower(strings.TrimSpace(v.Verdict))
	if v.Verdict != "pass" && v.Verdict != "fail" {
		return v, fmt.Errorf("verdict %q is neither pass nor fail", v.Verdict)
	}
	return v, nil
}

// Condense writes a transcript as text for the prompt: each tool call with
// its result shortened, within max characters.
func Condense(t *transcript.Transcript, max int) string {
	if t == nil {
		return ""
	}
	per := 600
	for {
		var b strings.Builder
		for _, s := range t.Steps {
			what := s.Shell
			if what == "" {
				what = s.Path
			}
			if what == "" {
				what = string(s.Input)
			}
			if s.Before != "" {
				fmt.Fprintf(&b, "agent: %s\n", clipTo(s.Before, per))
			}
			status := ""
			if s.Error {
				status = " (FAILED)"
			}
			fmt.Fprintf(&b, "[%d] %s: %s%s\n    -> %s\n", s.Index, s.Tool, clipTo(what, per), status, clipTo(s.Result, per))
		}
		out := b.String()
		if len(out) <= max || per <= 60 {
			if len(out) > max {
				out = out[:max] + "\n…[transcript cut]"
			}
			return out
		}
		per /= 2
	}
}

func clipTo(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// readFile is a seam for the calibration loader.
var readFile = os.ReadFile
