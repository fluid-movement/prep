package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexSkillUsesPrepOnlyAttribution(t *testing.T) {
	root := filepath.Join("..", "..")
	path := filepath.Join(root, "plugins", "codex", "skills", "prep", "SKILL.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), skillText) {
		t.Fatal("Codex skill base differs from embedded skill; run just sync-skill")
	}
	for _, text := range []string{string(b), skillText} {
		if !strings.Contains(text, "For prep CLI invocations only") || !strings.Contains(text, "Leave unrelated shell commands unchanged") {
			t.Fatal("skill requests attribution beyond prep CLI invocations")
		}
		if strings.Contains(text, "for shell commands so prep CLI activity") {
			t.Fatal("legacy blanket attribution guidance remains")
		}
	}
	// Keep the generator scoped too, so a later sync cannot restore the bug.
	recipe, err := os.ReadFile(filepath.Join(root, "justfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(recipe), "only to prep CLI invocations") || strings.Contains(string(recipe), "for shell commands so prep CLI activity") {
		t.Fatal("skill generator would restore blanket attribution guidance")
	}
}

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
