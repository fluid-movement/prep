package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexLegacyHookLauncher(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("legacy launcher requires Python")
	}
	script, err := filepath.Abs("../../plugins/codex/hooks/prep_hook.py")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	stub := filepath.Join(bin, "prep")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(stub, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	run := func(mode string) string {
		t.Helper()
		cmd := exec.Command(python, script, mode)
		cmd.Env = append(os.Environ(), "PATH="+bin)
		cmd.Stdin = strings.NewReader(`{"prompt":"test"}`)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("legacy hook blocked: %v %s", err, b)
		}
		return string(b)
	}
	// A missing binary must stay silent instead of exiting 2, which Codex
	// interprets as an intentional denial of the pending tool call.
	if out := run("before-tool"); out != "" {
		t.Fatalf("missing binary: %s", out)
	}
	write("printf '%s\\n' \"$*\"\n/bin/cat\n")
	for old, current := range map[string]string{"session": "session-start", "prompt": "prompt", "before-tool": "before-tool", "after-tool": "after-tool"} {
		out := run(old)
		if !strings.Contains(out, "hook "+current+" --harness codex") || !strings.Contains(out, `{"prompt":"test"}`) {
			t.Fatalf("forwarding: %s", out)
		}
	}
	if out := run("unknown"); out != "" {
		t.Fatalf("unknown mode: %s", out)
	}
	write("echo unexpected-output\nexit 2\n")
	if out := run("before-tool"); out != "" {
		t.Fatalf("failed binary output leaked: %s", out)
	}
}
