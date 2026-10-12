package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/eval/assert"
	_ "github.com/fluid-movement/prep/internal/eval/assert/outcome"
	_ "github.com/fluid-movement/prep/internal/eval/assert/tools"
	"github.com/fluid-movement/prep/internal/eval/env"
	"github.com/fluid-movement/prep/internal/eval/evalcase"
)

var prepBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "run-test")
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

const caseSrc = `
name: make-refund-issue
prompt: "File an issue about {{cart}}"
limits: {timeout: 3s, max_turns: 7, max_budget_usd: 0.25}
spec:
  name: shop
  issues:
    - {alias: cart, title: Cart page}
expect:
  tools:
    - ran: new
  outcome:
    - issue: {title: refund, parent: '{{cart}}'}
`

func parse(t *testing.T) *evalcase.Case {
	t.Helper()
	c, err := evalcase.Parse(strings.NewReader(caseSrc), "case.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// fakeClaude writes a claude stand-in: it records its arguments and stdin,
// runs body (shell, in the primary project) and prints the stream.
func fakeClaude(t *testing.T, body, stream string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "claude")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.9.9 (Claude Code)"; exit 0; fi
printf '%s\n' "$@" > "$CLAUDE_CONFIG_DIR/args.txt"
cat > "$CLAUDE_CONFIG_DIR/stdin.txt"
` + body + `
cat <<'STREAM'
` + stream + `
STREAM
`
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

const goodStream = `{"type":"system","subtype":"init","cwd":"/x","session_id":"s","model":"claude-haiku-5-5","claude_code_version":"9.9.9","skills":[],"plugins":[]}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"prep new --title Refund --kind code"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"ok"}]}}
{"type":"result","subtype":"success","is_error":false,"result":"done","num_turns":2,"total_cost_usd":0.01}`

func request(t *testing.T, claude string) Request {
	return Request{Case: parse(t), Variant: Variant{Name: "current", Hash: "h1", PluginDir: "/plugin"}, Model: "haiku",
		Binary: prepBin, Claude: claude, Auth: env.Auth{OAuthToken: "tok"}, Parent: t.TempDir(), OutDir: t.TempDir()}
}

func TestRunCompletes(t *testing.T) {
	claude := fakeClaude(t, `prep new --title "Refund flow" --kind code --parent "$(prep list --json | sed -n 's/.*"id": "\([A-Z0-9]*\)".*/\1/p' | head -1)" >/dev/null`, goodStream)
	req := request(t, claude)
	rec, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Release(true)
	if rec.Status != Completed || rec.Err != "" {
		t.Fatalf("status %s: %s", rec.Status, rec.Err)
	}
	if rec.ModelResolved != "claude-haiku-5-5" || rec.HarnessVersion != "9.9.9 (Claude Code)" || rec.Model != "haiku" {
		t.Errorf("model/version: %+v", rec)
	}
	args, _ := os.ReadFile(filepath.Join(rec.Dir, "claude", "args.txt"))
	for _, want := range []string{"-p", "stream-json", "--verbose", "--model\nhaiku", "--max-budget-usd\n0.25", "--max-turns\n7", "bypassPermissions", "--plugin-dir\n/plugin", "--strict-mcp-config"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("argument %q missing:\n%s", want, args)
		}
	}
	if stdin, _ := os.ReadFile(filepath.Join(rec.Dir, "claude", "stdin.txt")); !strings.HasPrefix(string(stdin), "File an issue about 01") {
		t.Errorf("prompt: %q", stdin)
	}
	// Assertions run against what the fake agent did.
	in := rec.Input
	if len(in.Diff.Projects["shop"].Created) != 1 {
		t.Fatalf("diff: %s", in.Diff.Summary(in.After))
	}
	for _, a := range append(append([]assert.Assertion{}, rec.Compiled.Tools...), rec.Compiled.Outcome...) {
		if r := a.Check(in); !r.Pass {
			t.Errorf("%s failed: %s", a.Describe(), r.Reason)
		}
	}
	for _, f := range []string{"stream.jsonl", "transcript.json", "record.json", "stderr.txt"} {
		if _, err := os.Stat(filepath.Join(req.OutDir, f)); err != nil {
			t.Errorf("%s not written: %v", f, err)
		}
	}
}

func TestReleaseKeepsFailedRuns(t *testing.T) {
	rec, err := Run(context.Background(), request(t, fakeClaude(t, "true", goodStream)))
	if err != nil {
		t.Fatal(err)
	}
	dir := rec.Dir
	if kept := rec.Release(false); kept != dir {
		t.Errorf("failing run not kept: %q", kept)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("kept directory gone: %v", err)
	}
	os.RemoveAll(dir)
	rec2, _ := Run(context.Background(), request(t, fakeClaude(t, "true", goodStream)))
	if kept := rec2.Release(true); kept != "" {
		t.Errorf("passing run kept: %q", kept)
	}
	if _, err := os.Stat(rec2.Dir); !os.IsNotExist(err) {
		t.Errorf("directory remains: %v", err)
	}
	req := request(t, fakeClaude(t, "true", goodStream))
	req.Keep = true
	rec3, _ := Run(context.Background(), req)
	if kept := rec3.Release(true); kept == "" {
		t.Error("Keep ignored")
	}
	os.RemoveAll(rec3.Dir)
}

func TestStatuses(t *testing.T) {
	init := `{"type":"system","subtype":"init","model":"m","session_id":"s"}` + "\n"
	for name, tc := range map[string]struct {
		body, stream string
		want         Status
		errHas       string
	}{
		"turns":   {"true", init + `{"type":"result","subtype":"error_max_turns","is_error":true,"terminal_reason":"max_turns","result":""}`, Turns, "turn limit"},
		"budget":  {"true", init + `{"type":"result","subtype":"error_max_budget_usd","is_error":true,"result":""}`, Budget, "budget"},
		"crash":   {"echo 'boom: no network' >&2; exit 3", "", Errored, "boom"},
		"empty":   {"true", "", Errored, "no result"},
		"timeout": {"sleep 30", "", Timeout, "timed out"},
	} {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			rec, err := Run(context.Background(), request(t, fakeClaude(t, tc.body, tc.stream)))
			if err != nil {
				t.Fatal(err)
			}
			defer rec.Release(true)
			if rec.Status != tc.want || !strings.Contains(rec.Err, tc.errHas) {
				t.Errorf("status %s err %q, want %s containing %q", rec.Status, rec.Err, tc.want, tc.errHas)
			}
			if rec.Input == nil {
				t.Error("state not captured after a stopped run")
			}
			if name == "timeout" && time.Since(start) > 15*time.Second {
				t.Errorf("timeout took %v", time.Since(start))
			}
		})
	}
}

func TestMissingAuthIsAnError(t *testing.T) {
	req := request(t, fakeClaude(t, "true", goodStream))
	req.Auth = env.Auth{}
	if _, err := Run(context.Background(), req); err != env.ErrNoAuth {
		t.Errorf("error: %v", err)
	}
	req = request(t, fakeClaude(t, "true", goodStream))
	req.Model = ""
	if _, err := Run(context.Background(), req); err == nil {
		t.Error("no model accepted")
	}
}

func TestHarnessMissing(t *testing.T) {
	req := request(t, "/nonexistent/claude")
	rec, err := Run(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Release(true)
	if rec.Status != Errored {
		t.Errorf("status %s %s", rec.Status, rec.Err)
	}
}
