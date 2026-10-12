package env

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/eval/transcript"
)

func fakeClaude(t *testing.T, script string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func fakePrep(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "prep-src")
	if err := os.WriteFile(p, []byte("#!/bin/sh\necho pinned\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAuthRequired(t *testing.T) {
	_, err := AuthFromEnv(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "claude setup-token") || !strings.Contains(err.Error(), "CLAUDE_CODE_OAUTH_TOKEN") {
		t.Errorf("error: %v", err)
	}
	if _, err := New(Config{Claude: fakeClaude(t, "true")}); err != ErrNoAuth {
		t.Errorf("New without auth: %v", err)
	}
	a, err := AuthFromEnv(func(k string) string {
		if k == "ANTHROPIC_API_KEY" {
			return "k"
		}
		return ""
	})
	if err != nil || a.APIKey != "k" {
		t.Errorf("api key: %v %+v", err, a)
	}
}

func TestEnvironIsBuiltFromScratch(t *testing.T) {
	t.Setenv("AWS_SECRET_ACCESS_KEY", "leak")
	t.Setenv("ANTHROPIC_BASE_URL", "http://leak")
	t.Setenv("CLAUDE_CONFIG_DIR", "/real/config")
	t.Setenv("PREP_ACTOR", "someone")
	t.Setenv("XDG_STATE_HOME", "/real/state")
	out := filepath.Join(t.TempDir(), "env.txt")
	claude := fakeClaude(t, "env > '"+out+"'\ncommand -v prep >> '"+out+"'\npwd >> '"+out+"'\n")
	e, err := New(Config{Binary: fakePrep(t), Claude: claude, Auth: Auth{OAuthToken: "tok"}})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Cleanup()
	cmd := e.Command(e.Workspace, "-p", "hi")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	data, _ := os.ReadFile(out)
	got := string(data)
	for _, bad := range []string{"AWS_SECRET", "ANTHROPIC_BASE_URL", "/real/", "PREP_ACTOR"} {
		if strings.Contains(got, bad) {
			t.Errorf("%q leaked into the harness environment:\n%s", bad, got)
		}
	}
	for _, want := range []string{"CLAUDE_CONFIG_DIR=" + e.ConfigDir, "HOME=" + e.Home, "XDG_STATE_HOME=" + e.StateDir, "CLAUDE_CODE_OAUTH_TOKEN=tok", filepath.Join(e.Bin, "prep")} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	if !strings.HasPrefix(e.Environ()[5], "PATH="+e.Bin+":") {
		t.Errorf("pinned binary is not first on PATH: %v", e.Environ())
	}
	// The pinned copy is independent of the source.
	if b, err := os.ReadFile(filepath.Join(e.Bin, "prep")); err != nil || !strings.Contains(string(b), "pinned") {
		t.Errorf("pinned binary: %v %s", err, b)
	}
}

func TestLeaks(t *testing.T) {
	e, err := New(Config{Claude: fakeClaude(t, "true"), Auth: Auth{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Cleanup()
	want := Expected{Plugins: []string{"prep"}, Skills: []string{"prep"}}
	clean := transcript.Meta{
		Plugins: []string{"prep"}, Skills: []string{"prep", "simplify"}, Cwd: e.Workspace,
		MemoryPaths: []string{filepath.Join(e.ConfigDir, "memory")},
	}
	if l := e.Leaks(clean, want); len(l) != 0 {
		t.Errorf("clean session reported leaks: %v", l)
	}
	dirty := transcript.Meta{
		Plugins: []string{"prep", "other-plugin"}, Skills: []string{"prep", "my-private-skill", "prep:prep"},
		Cwd: "/Users/someone/project", MemoryPaths: []string{"/Users/someone/.claude/memory"},
	}
	l := strings.Join(e.Leaks(dirty, want), "|")
	for _, w := range []string{"plugin other-plugin", "skill my-private-skill", "memory path /Users/someone", "working directory"} {
		if !strings.Contains(l, w) {
			t.Errorf("leak %q not reported in %q", w, l)
		}
	}
	if strings.Contains(l, "skill prep") {
		t.Errorf("the variant's own skill reported: %q", l)
	}
}

func TestCleanup(t *testing.T) {
	e, err := New(Config{Claude: fakeClaude(t, "true"), Auth: Auth{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	e.Cleanup()
	if _, err := os.Stat(e.Dir); !os.IsNotExist(err) {
		t.Errorf("directory remains: %v", err)
	}
}

func TestPluginArgs(t *testing.T) {
	if got := strings.Join(PluginArgs("/p"), " "); got != "--plugin-dir /p" {
		t.Errorf("args: %q", got)
	}
	if PluginArgs("") != nil {
		t.Error("no plugin should give no arguments")
	}
}

func TestProbeAndExpectedFromPlugin(t *testing.T) {
	plugin := t.TempDir()
	os.MkdirAll(filepath.Join(plugin, ".claude-plugin"), 0o755)
	os.WriteFile(filepath.Join(plugin, ".claude-plugin", "plugin.json"), []byte(`{"name":"prep","version":"1"}`), 0o644)
	os.MkdirAll(filepath.Join(plugin, "skills", "prep"), 0o755)
	want := ExpectedFromPlugin(plugin)
	if len(want.Plugins) != 1 || want.Plugins[0] != "prep" || len(want.Skills) != 1 || want.Skills[0] != "prep" {
		t.Fatalf("expected: %+v", want)
	}
	if e := ExpectedFromPlugin(""); len(e.Plugins)+len(e.Skills) != 0 {
		t.Errorf("no plugin: %+v", e)
	}
	claude := fakeClaude(t, `cat >/dev/null
printf '{"type":"system","subtype":"init","cwd":"%s","session_id":"s","model":"m","skills":["prep","stray-skill"],"plugins":[{"name":"prep"},{"name":"intruder"}],"memory_paths":{"auto":"/real/home/.claude/memory"}}\n' "$PWD"
`)
	e, err := New(Config{Claude: claude, Auth: Auth{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Cleanup()
	meta, err := e.Probe(context.Background(), "haiku", plugin)
	if err != nil {
		t.Fatal(err)
	}
	leaks := strings.Join(e.Leaks(meta, want), "|")
	for _, w := range []string{"plugin intruder", "skill stray-skill", "memory path /real/home"} {
		if !strings.Contains(leaks, w) {
			t.Errorf("leak %q not found in %q", w, leaks)
		}
	}
	bad := fakeClaude(t, "echo 'not logged in' >&2; exit 1\n")
	e2, _ := New(Config{Claude: bad, Auth: Auth{APIKey: "k"}})
	defer e2.Cleanup()
	if _, err := e2.Probe(context.Background(), "haiku", ""); err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("failing harness: %v", err)
	}
}
