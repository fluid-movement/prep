// Package run is the session runner: it builds the fixture workspace in an
// isolated environment, runs Claude Code headless on a case's prompt and
// returns what happened. It judges nothing.
package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
	"github.com/fluid-movement/prep/internal/eval/fixture"
	"github.com/fluid-movement/prep/internal/eval/transcript"
	"github.com/fluid-movement/prep/internal/eval/transcript/claudecode"
)

// Status says how a run ended.
type Status string

const (
	Completed Status = "completed"
	Timeout   Status = "timeout"
	Budget    Status = "budget"
	Turns     Status = "turns"
	// Errored means the harness failed (missing binary or auth, a crash);
	// it says nothing about the agent.
	Errored Status = "error"
)

// Variant is the harness configuration a run uses, already laid out on disk.
type Variant struct {
	Name string
	Hash string
	// PluginDir is loaded with --plugin-dir; empty means no plugin.
	PluginDir string
	// AppendSystemPrompt is added to the system prompt.
	AppendSystemPrompt string
}

// Request is one run.
type Request struct {
	Case    *evalcase.Case
	Variant Variant
	// Model is passed to --model and is recorded as requested.
	Model string
	// Binary is the prep binary under test; Claude the claude executable
	// ("" finds it on PATH).
	Binary string
	Claude string
	Auth   env.Auth
	// Parent is the directory the run's temp directory is created in.
	Parent string
	// OutDir, when set, receives stream.jsonl, stderr.txt, transcript.json
	// and record.json.
	OutDir string
	// Keep leaves the environment on disk after Release.
	Keep bool
}

// Record is the result of a run.
type Record struct {
	Case    string `json:"case"`
	Variant string `json:"variant"`
	Hash    string `json:"variant_hash,omitempty"`
	Model   string `json:"model"`
	// ModelResolved is the model id the harness reports.
	ModelResolved  string `json:"model_resolved,omitempty"`
	HarnessVersion string `json:"harness_version,omitempty"`
	Prompt         string `json:"prompt"`

	Status   Status        `json:"status"`
	Err      string        `json:"error,omitempty"`
	ExitCode int           `json:"exit_code"`
	Started  time.Time     `json:"started"`
	Duration time.Duration `json:"duration"`

	Transcript *transcript.Transcript `json:"-"`
	Compiled   *evalcase.Compiled     `json:"-"`
	// Input is what the assertions read.
	Input *assert.Input `json:"-"`
	// Stream is the path of the raw stream, when OutDir was set.
	Stream string `json:"stream,omitempty"`
	// Dir is the environment directory; it exists until Release.
	Dir string `json:"dir,omitempty"`

	env  *env.Env
	keep bool
}

// Release removes the environment unless the run is to be kept: Keep was
// asked for, or pass is false (a failing run keeps its workspace for
// inspection). It reports where a kept one is.
func (r *Record) Release(pass bool) (kept string) {
	if r.env == nil {
		return ""
	}
	if r.keep || !pass {
		return r.env.Dir
	}
	r.env.Cleanup()
	r.env = nil
	return ""
}

// Run executes a request. A failure of the harness is returned in the
// Record (Status Errored), not as an error; the error is for problems that
// stop a record from existing at all.
func Run(ctx context.Context, req Request) (*Record, error) {
	if req.Model == "" {
		return nil, errors.New("run: no model")
	}
	rec := &Record{Case: req.Case.Name, Variant: req.Variant.Name, Hash: req.Variant.Hash, Model: req.Model, Started: time.Now(), keep: req.Keep}
	e, err := env.New(env.Config{Binary: req.Binary, Claude: req.Claude, Auth: req.Auth, Parent: req.Parent})
	if err != nil {
		return nil, err
	}
	rec.env, rec.Dir = e, e.Dir
	fail := func(format string, a ...any) (*Record, error) {
		rec.Status, rec.Err = Errored, fmt.Sprintf(format, a...)
		rec.Duration = time.Since(rec.Started)
		req.write(rec)
		return rec, nil
	}

	ws, err := fixture.BuildWorkspace(req.Case.Resolved(), e.Workspace, fixture.Options{Binary: filepath.Join(e.Bin, "prep")})
	if err != nil {
		return fail("build fixture: %v", err)
	}
	e.StateDir = ws.State // the agent's prep sees the workspace's registry
	primary := ws.PrimaryProject()
	comp, err := req.Case.Compile(primary.Expand)
	if err != nil {
		return fail("compile case: %v", err)
	}
	rec.Compiled, rec.Prompt = comp, comp.Prompt
	before, err := assert.Capture(ws)
	if err != nil {
		return fail("snapshot before: %v", err)
	}
	if v, err := e.Command(e.Workspace, "--version").Output(); err == nil {
		rec.HarnessVersion = strings.TrimSpace(string(v))
	}

	lim := req.Case.Limits
	rctx, cancel := context.WithTimeout(ctx, time.Duration(lim.Timeout))
	defer cancel()
	cmd := e.Command(primary.Dir, claudeArgs(req, lim)...)
	cmd.Stdin = strings.NewReader(comp.Prompt)
	var stream, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stream, &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	rec.Started = time.Now()
	runErr := runCommand(rctx, cmd)
	rec.Duration = time.Since(rec.Started)

	if req.OutDir != "" {
		os.MkdirAll(req.OutDir, 0o755)
		rec.Stream = filepath.Join(req.OutDir, "stream.jsonl")
		os.WriteFile(rec.Stream, stream.Bytes(), 0o644)
		os.WriteFile(filepath.Join(req.OutDir, "stderr.txt"), stderr.Bytes(), 0o644)
	}
	tr, perr := claudecode.Parse(bytes.NewReader(stream.Bytes()))
	if perr != nil {
		return fail("read stream: %v", perr)
	}
	rec.Transcript = tr
	rec.ModelResolved = tr.Meta.Model
	if rec.HarnessVersion == "" {
		rec.HarnessVersion = tr.Meta.Version
	}
	if exit := (&exec.ExitError{}); errors.As(runErr, &exit) {
		rec.ExitCode = exit.ExitCode()
	}
	rec.Status, rec.Err = classify(rctx, runErr, tr, stderr.String())

	// The state the agent left, even when it was cut off.
	after, aerr := assert.Capture(ws)
	if aerr != nil {
		return fail("snapshot after: %v", aerr)
	}
	rec.Input = &assert.Input{Transcript: tr, Workspace: ws, Project: primary, Before: before, After: after, Diff: assert.Compare(ws, before, after)}
	req.write(rec)
	return rec, nil
}

func runCommand(ctx context.Context, cmd *exec.Cmd) error {
	// The command was built by env without a context; wire cancellation by
	// hand and kill the whole process group.
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // the whole process group
		select {
		case <-done:
		case <-time.After(5 * time.Second): // a grandchild still holds the pipes
		}
		return ctx.Err()
	}
}

func claudeArgs(req Request, lim evalcase.Limits) []string {
	args := []string{
		"-p",
		"--output-format", "stream-json", "--verbose",
		"--model", req.Model,
		"--max-budget-usd", strconv.FormatFloat(lim.MaxBudgetUSD, 'f', -1, 64),
		"--max-turns", strconv.Itoa(lim.MaxTurns),
		"--permission-mode", "bypassPermissions",
		"--disallowedTools", "WebFetch,WebSearch",
		"--strict-mcp-config",
	}
	args = append(args, env.PluginArgs(req.Variant.PluginDir)...)
	if req.Variant.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", req.Variant.AppendSystemPrompt)
	}
	return args
}

// classify decides how the session ended from the process outcome and the
// result event.
func classify(ctx context.Context, runErr error, tr *transcript.Transcript, stderr string) (Status, string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Timeout, "timed out"
	}
	if ctx.Err() != nil {
		return Errored, "cancelled"
	}
	status := tr.Final.Status
	switch {
	case strings.Contains(status, "max_turns") || tr.Final.Terminal == "max_turns":
		return Turns, "turn limit reached"
	case strings.Contains(status, "budget") || strings.Contains(tr.Final.Terminal, "budget"):
		return Budget, "budget limit reached"
	case tr.Final.IsError:
		// The harness reports a failure of its own (a session limit, an
		// auth error) even when the subtype says success.
		msg := strings.TrimSpace(tr.Final.Text)
		if msg == "" {
			msg = harnessMessage(stderr, tr.Final.Text, runErr)
		}
		return Errored, clip(msg)
	case status == "success":
		return Completed, ""
	case status == "" && runErr == nil && len(tr.Steps) > 0:
		return Completed, ""
	}
	return Errored, clip(harnessMessage(stderr, tr.Final.Text, runErr))
}

// harnessMessage says why the harness produced no usable result: its stderr,
// else its result text, else the process error.
func harnessMessage(stderr, text string, runErr error) string {
	msg := strings.TrimSpace(stderr)
	if msg == "" {
		msg = text
	}
	if msg == "" && runErr != nil {
		msg = runErr.Error()
	}
	if msg == "" {
		msg = "the harness produced no result"
	}
	return msg
}

func clip(msg string) string {
	if len(msg) > 500 {
		return msg[:500] + "…"
	}
	return msg
}

// write stores transcript.json and record.json.
func (req Request) write(rec *Record) {
	if req.OutDir == "" {
		return
	}
	os.MkdirAll(req.OutDir, 0o755)
	if rec.Transcript != nil {
		writeJSON(filepath.Join(req.OutDir, "transcript.json"), rec.Transcript)
	}
	writeJSON(filepath.Join(req.OutDir, "record.json"), rec)
}

func writeJSON(path string, v any) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(io.Writer(f))
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
