// Package env builds the isolated environment a test run happens in: a
// private home, Claude Code configuration directory and project registry,
// the prep binary under test pinned first on PATH, and nothing of the real
// user's configuration. Auth is the one thing passed through.
package env

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// Auth is how the harness logs in. A fresh configuration directory has no
// login, so one of the two must come from the caller's environment.
type Auth struct {
	OAuthToken string // CLAUDE_CODE_OAUTH_TOKEN, from `claude setup-token`
	APIKey     string // ANTHROPIC_API_KEY
}

// ErrNoAuth is returned when neither credential is available.
var ErrNoAuth = errors.New("no credentials for the test runs: a private Claude Code configuration has no login, so " +
	"run `just eval-login` once (it runs `claude setup-token` and keeps the token in ~/.config/prep/eval-token; uses your subscription), " +
	"or export CLAUDE_CODE_OAUTH_TOKEN, or ANTHROPIC_API_KEY (billed per token)")

// TokenFile is where `just eval-login` stores the token: eval-token in
// prep's user configuration directory. "" when the home is unknown.
func TokenFile(getenv func(string) string) string {
	if getenv == nil {
		getenv = os.Getenv
	}
	dir := getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home := getenv("HOME")
		if home == "" {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "prep", "eval-token")
}

// AuthFromEnv reads the credentials with getenv (os.Getenv if nil): the
// CLAUDE_CODE_OAUTH_TOKEN and ANTHROPIC_API_KEY variables, else the token
// file.
func AuthFromEnv(getenv func(string) string) (Auth, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	a := Auth{OAuthToken: getenv("CLAUDE_CODE_OAUTH_TOKEN"), APIKey: getenv("ANTHROPIC_API_KEY")}
	if a.OAuthToken == "" && a.APIKey == "" {
		if f := TokenFile(getenv); f != "" {
			if data, err := os.ReadFile(f); err == nil {
				a.OAuthToken = strings.TrimSpace(string(data))
			}
		}
	}
	if a.OAuthToken == "" && a.APIKey == "" {
		return a, ErrNoAuth
	}
	return a, nil
}

func (a Auth) environ() []string {
	var out []string
	if a.OAuthToken != "" {
		out = append(out, "CLAUDE_CODE_OAUTH_TOKEN="+a.OAuthToken)
	}
	if a.APIKey != "" {
		out = append(out, "ANTHROPIC_API_KEY="+a.APIKey)
	}
	return out
}

// Config says how to build an Env.
type Config struct {
	// Binary is the prep binary under test; it is copied into the Env.
	Binary string
	// Claude is the claude executable (default: found on the caller's PATH).
	Claude string
	Auth   Auth
	// Parent is where the Env directory is created ("" for the system
	// temp directory).
	Parent string
}

// Env is one run's isolated environment.
type Env struct {
	// Dir holds everything below.
	Dir       string
	Home      string
	ConfigDir string // CLAUDE_CONFIG_DIR
	StateDir  string // XDG_STATE_HOME: the project registry
	Bin       string // directory holding the pinned prep
	// Workspace is where the fixture projects are built.
	Workspace string
	// Claude is the executable run.
	Claude string

	auth Auth
	path string
}

// New creates the directories, pins the binary and resolves claude. The
// caller removes it with Cleanup.
func New(cfg Config) (*Env, error) {
	if cfg.Auth.OAuthToken == "" && cfg.Auth.APIKey == "" {
		return nil, ErrNoAuth
	}
	claude := cfg.Claude
	if claude == "" {
		p, err := exec.LookPath("claude")
		if err != nil {
			return nil, fmt.Errorf("claude not found on PATH: %w", err)
		}
		claude = p
	}
	claude, err := filepath.Abs(claude)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(cfg.Parent, "prep-eval-")
	if err != nil {
		return nil, err
	}
	e := &Env{
		Dir:       dir,
		Home:      filepath.Join(dir, "home"),
		ConfigDir: filepath.Join(dir, "claude"),
		StateDir:  filepath.Join(dir, "state"),
		Bin:       filepath.Join(dir, "bin"),
		Workspace: filepath.Join(dir, "work"),
		Claude:    claude,
		auth:      cfg.Auth,
	}
	for _, d := range []string{e.Home, e.ConfigDir, e.StateDir, e.Bin, e.Workspace} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			e.Cleanup()
			return nil, err
		}
	}
	if cfg.Binary != "" {
		if err := copyExecutable(cfg.Binary, filepath.Join(e.Bin, "prep")); err != nil {
			e.Cleanup()
			return nil, fmt.Errorf("pin prep binary: %w", err)
		}
	}
	e.path = strings.Join([]string{e.Bin, filepath.Dir(claude), "/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin"}, string(os.PathListSeparator))
	return e, nil
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// Cleanup removes the Env's directory.
func (e *Env) Cleanup() { os.RemoveAll(e.Dir) }

// Environ is the harness process's environment, built from scratch: nothing
// of the caller's reaches it but the credentials, TERM-free defaults and the
// temp directory.
func (e *Env) Environ() []string {
	env := []string{
		"HOME=" + e.Home,
		"CLAUDE_CONFIG_DIR=" + e.ConfigDir,
		"XDG_STATE_HOME=" + e.StateDir,
		"XDG_CONFIG_HOME=" + filepath.Join(e.Home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(e.Home, ".cache"),
		"PATH=" + e.path,
		"TERM=dumb",
		"LANG=en_US.UTF-8",
		"TMPDIR=" + os.TempDir(),
		"CLAUDE_CODE_DISABLE_AUTO_MEMORY=1",
		"DISABLE_AUTOUPDATER=1",
		"DISABLE_TELEMETRY=1",
	}
	return append(env, e.auth.environ()...)
}

// Command builds a claude command running in dir with the isolated
// environment.
func (e *Env) Command(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(e.Claude, args...)
	cmd.Dir = dir
	cmd.Env = e.Environ()
	return cmd
}

// Expected is what a run is allowed to see.
type Expected struct {
	// Plugins are the plugin names the variant loads.
	Plugins []string
	// Skills are the skill names those plugins provide (matched as a suffix
	// after any "plugin:" prefix).
	Skills []string
}

// Leaks says what a session could see that does not come from the variant,
// given the metadata of its init event. An empty result means isolated.
func (e *Env) Leaks(m transcript.Meta, want Expected) []string {
	var out []string
	allowed := func(list []string, name string) bool {
		base := name
		if i := strings.LastIndex(name, ":"); i >= 0 {
			base = name[i+1:]
		}
		for _, w := range list {
			if w == name || w == base {
				return true
			}
		}
		return false
	}
	for _, p := range m.Plugins {
		if !allowed(want.Plugins, filepath.Base(p)) && !allowed(want.Plugins, p) {
			out = append(out, "plugin "+p)
		}
	}
	for _, s := range m.Skills {
		if !allowed(want.Skills, s) && !builtinSkill(s) {
			out = append(out, "skill "+s)
		}
	}
	for _, p := range m.MemoryPaths {
		if !strings.HasPrefix(p, e.Dir) {
			out = append(out, "memory path "+p)
		}
	}
	if m.Cwd != "" && !strings.HasPrefix(resolve(m.Cwd), resolve(e.Dir)) {
		out = append(out, "working directory "+m.Cwd)
	}
	return out
}

// BuiltinSkills are skills Claude Code ships itself, which every session
// has and no configuration removes. The doctor command can extend the list
// when a Claude Code release adds one.
var BuiltinSkills = []string{
	"update-config", "keybindings-help", "simplify", "loop", "schedule", "claude-api",
	"init", "review", "security-review", "code-review", "run", "fewer-permission-prompts",
	// seen in the init event of Claude Code 2.1.296 with an empty configuration
	"deep-research", "dataviz", "verify", "debug", "batch", "doctor", "workflow-authoring",
	"run-skill-generator", "plugin-authoring",
}

func builtinSkill(name string) bool {
	for _, b := range BuiltinSkills {
		if b == name {
			return true
		}
	}
	return false
}

func resolve(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}

// PluginArgs are the arguments that load a variant's plugin for one session
// only. Plugins are never installed with `prep setup` or `claude plugin
// install`, so the real user's installation is untouched.
func PluginArgs(pluginDir string) []string {
	if pluginDir == "" {
		return nil
	}
	return []string{"--plugin-dir", pluginDir}
}
