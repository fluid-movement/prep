package env

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/eval/proc"
	"github.com/fluid-movement/prep/internal/eval/transcript"
	"github.com/fluid-movement/prep/internal/eval/transcript/claudecode"
)

// ExpectedFromPlugin reads what a plugin directory provides: its name from
// .claude-plugin/plugin.json and the skills under skills/.
func ExpectedFromPlugin(dir string) Expected {
	var want Expected
	if dir == "" {
		return want
	}
	if data, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json")); err == nil {
		var p struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &p) == nil && p.Name != "" {
			want.Plugins = append(want.Plugins, p.Name)
		}
	}
	if entries, err := os.ReadDir(filepath.Join(dir, "skills")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				want.Skills = append(want.Skills, e.Name())
			}
		}
	}
	sort.Strings(want.Skills)
	return want
}

// Probe starts one trivial session in the Env and returns what the harness
// reported seeing in its init event: the skills, plugins, tools and memory
// paths. It costs one short model turn.
func (e *Env) Probe(ctx context.Context, model, pluginDir string) (transcript.Meta, error) {
	args := []string{"-p", "--output-format", "stream-json", "--verbose", "--model", model,
		"--max-turns", "1", "--max-budget-usd", "0.10", "--permission-mode", "bypassPermissions", "--strict-mcp-config"}
	args = append(args, PluginArgs(pluginDir)...)
	cmd := e.Command(e.Workspace, args...)
	cmd.Stdin = strings.NewReader("Reply with the single word ok.")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	proc.Group(cmd)
	if err := cmd.Start(); err != nil {
		return transcript.Meta{}, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-ctx.Done():
		proc.Kill(cmd)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
		return transcript.Meta{}, ctx.Err()
	}
	tr, err := claudecode.Parse(bytes.NewReader(out.Bytes()))
	if err != nil {
		return transcript.Meta{}, err
	}
	if tr.Meta.SessionID == "" {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = strings.TrimSpace(out.String())
		}
		return tr.Meta, fmt.Errorf("the harness reported no session: %s", msg)
	}
	if tr.Final.IsError {
		return tr.Meta, fmt.Errorf("the probe session failed: %s", strings.TrimSpace(tr.Final.Text))
	}
	return tr.Meta, nil
}
