// Package claudecode turns Claude Code's stream-json output (claude -p
// --output-format stream-json --verbose) into a transcript.Transcript.
package claudecode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// Parse reads a stream and normalizes it. It is tolerant: unknown event
// types and fields are ignored, and a line that does not parse is noted in
// Meta.Warnings, never fatal. The only error is a failing reader.
func Parse(r io.Reader) (*transcript.Transcript, error) {
	p := &parser{t: &transcript.Transcript{Meta: transcript.Meta{Harness: "claude-code"}}, byID: map[string]int{}}
	br := bufio.NewReaderSize(r, 1<<20)
	lineNo := 0
	for {
		line, err := br.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			lineNo++
			if perr := p.line(line); perr != nil {
				p.t.Meta.Warnings = append(p.t.Meta.Warnings, fmt.Sprintf("line %d: %v", lineNo, perr))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return p.t, err
		}
	}
	t := p.t
	t.Counts.ToolCalls = len(t.Steps)
	for _, s := range t.Steps {
		if s.Error {
			t.Counts.Errors++
		}
	}
	if t.Final.Text == "" && len(p.pending) > 0 {
		t.Final.Text = strings.Join(p.pending, "\n")
	}
	return t, nil
}

type parser struct {
	t       *transcript.Transcript
	byID    map[string]int // tool_use id -> index in t.Steps
	pending []string       // assistant text since the last tool call
}

type block struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

func (p *parser) line(line []byte) error {
	var e struct {
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		Parent  string `json:"parent_tool_use_id"`
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(line, &e); err != nil {
		return err
	}
	switch e.Type {
	case "system":
		switch e.Subtype {
		case "init":
			return p.init(line)
		case "hook_response":
			return p.hook(line)
		}
	case "assistant":
		return p.assistant(e.Message.Content, e.Parent)
	case "user":
		return p.user(e.Message.Content)
	case "result":
		return p.result(line)
	}
	return nil
}

func (p *parser) init(line []byte) error {
	var d struct {
		SessionID string          `json:"session_id"`
		Cwd       string          `json:"cwd"`
		Model     string          `json:"model"`
		Version   string          `json:"claude_code_version"`
		Tools     []string        `json:"tools"`
		Skills    []string        `json:"skills"`
		Plugins   json.RawMessage `json:"plugins"`
		Memory    json.RawMessage `json:"memory_paths"`
	}
	if err := json.Unmarshal(line, &d); err != nil {
		return err
	}
	m := &p.t.Meta
	m.SessionID, m.Cwd, m.Model, m.Version = d.SessionID, d.Cwd, d.Model, d.Version
	m.Tools, m.Skills = d.Tools, d.Skills
	m.Plugins, m.BuiltinPlugins = splitPlugins(d.Plugins)
	m.MemoryPaths = names(d.Memory)
	return nil
}

// splitPlugins separates the plugins that ship inside Claude Code (path
// "builtin") from the ones loaded from disk.
func splitPlugins(raw json.RawMessage) (loaded, builtin []string) {
	var objs []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if json.Unmarshal(raw, &objs) != nil {
		return names(raw), nil
	}
	for _, o := range objs {
		name := o.Name
		if name == "" {
			name = o.Path
		}
		if o.Path == "builtin" {
			builtin = append(builtin, name)
		} else {
			loaded = append(loaded, name)
		}
	}
	return loaded, builtin
}

// names reads a list that holds strings or objects with a name or path.
func names(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var strs []string
	if json.Unmarshal(raw, &strs) == nil {
		return strs
	}
	var objs []map[string]any
	if json.Unmarshal(raw, &objs) == nil {
		var out []string
		for _, o := range objs {
			for _, k := range []string{"name", "path", "id"} {
				if s, ok := o[k].(string); ok && s != "" {
					out = append(out, s)
					break
				}
			}
		}
		return out
	}
	var obj map[string]any // memory_paths is an object of named paths
	if json.Unmarshal(raw, &obj) == nil {
		var out []string
		for _, v := range obj {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func (p *parser) hook(line []byte) error {
	var d struct {
		Name     string `json:"hook_name"`
		Event    string `json:"hook_event"`
		Output   string `json:"output"`
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exit_code"`
	}
	if err := json.Unmarshal(line, &d); err != nil {
		return err
	}
	out := d.Output
	if out == "" {
		out = d.Stdout
	}
	p.t.Meta.Hooks = append(p.t.Meta.Hooks, transcript.HookRun{Name: d.Name, Event: d.Event, Output: transcript.Clip(out), ExitCode: d.ExitCode})
	return nil
}

func (p *parser) assistant(content json.RawMessage, parent string) error {
	blocks, err := blocks(content)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		switch b.Type {
		case "text":
			if strings.TrimSpace(b.Text) != "" {
				p.pending = append(p.pending, b.Text)
			}
		case "tool_use":
			s := transcript.Step{Index: len(p.t.Steps), Tool: b.Name, Input: b.Input, Parent: parent, Before: strings.Join(p.pending, "\n")}
			p.pending = nil
			fillTarget(&s)
			p.byID[b.ID] = len(p.t.Steps)
			p.t.Steps = append(p.t.Steps, s)
		}
	}
	return nil
}

// fillTarget sets Path, Shell and Commands from the tool's input.
func fillTarget(s *transcript.Step) {
	var in map[string]any
	if json.Unmarshal(s.Input, &in) != nil {
		return
	}
	str := func(k string) string { v, _ := in[k].(string); return v }
	switch s.Tool {
	case "Bash":
		s.Shell = str("command")
		s.Commands = transcript.ParseShell(s.Shell)
	default:
		for _, k := range []string{"file_path", "path", "notebook_path"} {
			if v := str(k); v != "" {
				s.Path = v
				break
			}
		}
	}
}

func (p *parser) user(content json.RawMessage) error {
	blocks, err := blocks(content)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		if b.Type != "tool_result" {
			continue
		}
		i, ok := p.byID[b.ToolUseID]
		if !ok {
			p.t.Meta.Warnings = append(p.t.Meta.Warnings, "tool_result for unknown call "+b.ToolUseID)
			continue
		}
		text := resultText(b.Content)
		s := &p.t.Steps[i]
		s.Result, s.ResultLen, s.Error = transcript.Clip(text), len(text), b.IsError
	}
	return nil
}

// blocks reads message content, which is a string or a list of blocks.
func blocks(raw json.RawMessage) ([]block, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return []block{{Type: "text", Text: s}}, nil
	}
	var bs []block
	if err := json.Unmarshal(raw, &bs); err != nil {
		return nil, err
	}
	return bs, nil
}

// resultText reads a tool result's content: a string, or blocks of text.
func resultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var bs []block
	if json.Unmarshal(raw, &bs) == nil {
		var parts []string
		for _, b := range bs {
			if b.Type == "text" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return string(raw)
}

func (p *parser) result(line []byte) error {
	var d struct {
		Subtype     string  `json:"subtype"`
		IsError     bool    `json:"is_error"`
		Result      string  `json:"result"`
		StopReason  string  `json:"stop_reason"`
		Terminal    string  `json:"terminal_reason"`
		NumTurns    int     `json:"num_turns"`
		Cost        float64 `json:"total_cost_usd"`
		DurationMS  int     `json:"duration_ms"`
		DurationAPI int     `json:"duration_api_ms"`
		Usage       struct {
			In         int `json:"input_tokens"`
			Out        int `json:"output_tokens"`
			CacheRead  int `json:"cache_read_input_tokens"`
			CacheWrite int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(line, &d); err != nil {
		return err
	}
	t := p.t
	t.Final = transcript.Final{Text: d.Result, StopReason: d.StopReason, Status: d.Subtype, IsError: d.IsError, Terminal: d.Terminal}
	c := &t.Counts
	c.Turns, c.CostUSD = d.NumTurns, d.Cost
	c.InputTokens, c.OutputTokens, c.CacheRead, c.CacheWrite = d.Usage.In, d.Usage.Out, d.Usage.CacheRead, d.Usage.CacheWrite
	c.DurationMS = d.DurationMS
	if c.DurationMS == 0 {
		c.DurationMS = d.DurationAPI
	}
	return nil
}
