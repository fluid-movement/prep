// Package transcript is the harness-neutral form of an agent session: the
// tool calls in order, prep commands parsed out of shell calls, the final
// message and the counts. Assertions and the judge read only this, so a new
// harness needs an adapter and nothing else.
package transcript

import "encoding/json"

// Transcript is one normalized session.
type Transcript struct {
	Meta   Meta   `json:"meta"`
	Steps  []Step `json:"steps"`
	Final  Final  `json:"final"`
	Counts Counts `json:"counts"`
}

// Meta says what ran and what it could see.
type Meta struct {
	Harness     string    `json:"harness"`
	Version     string    `json:"version,omitempty"`
	Model       string    `json:"model,omitempty"` // as resolved by the harness
	SessionID   string    `json:"session_id,omitempty"`
	Cwd         string    `json:"cwd,omitempty"`
	Tools       []string  `json:"tools,omitempty"`
	Skills      []string  `json:"skills,omitempty"`
	Plugins     []string  `json:"plugins,omitempty"`
	MemoryPaths []string  `json:"memory_paths,omitempty"`
	Hooks       []HookRun `json:"hooks,omitempty"`
	Warnings    []string  `json:"warnings,omitempty"`
}

// HookRun is a hook that ran during the session.
type HookRun struct {
	Name     string `json:"name"`
	Event    string `json:"event,omitempty"`
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// Step is one tool call with its result.
type Step struct {
	Index int `json:"index"`
	// Tool is the harness's tool name (Bash, Read, Edit, Grep...).
	Tool  string          `json:"tool"`
	Input json.RawMessage `json:"input,omitempty"`
	// Path is the file or directory the call targets, when it names one.
	Path string `json:"path,omitempty"`
	// Commands are the prep commands a shell call runs, in order.
	Commands []Command `json:"commands,omitempty"`
	// Shell is the shell command text of a Bash call.
	Shell string `json:"shell,omitempty"`
	// Result is the tool result, clipped to ClipLen; ResultLen is its size.
	Result    string `json:"result,omitempty"`
	ResultLen int    `json:"result_len"`
	Error     bool   `json:"error,omitempty"`
	// Before is what the assistant said just before the call.
	Before string `json:"before,omitempty"`
}

// Final is how the session ended.
type Final struct {
	Text       string `json:"text,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	// Status is the harness's result subtype ("success", "error_max_turns"...).
	Status  string `json:"status,omitempty"`
	IsError bool   `json:"is_error,omitempty"`
	// Terminal is the harness's reason the session ended ("max_turns",
	// "completed"...), when it says.
	Terminal string `json:"terminal,omitempty"`
}

// Counts are the size and cost of the session.
type Counts struct {
	ToolCalls    int     `json:"tool_calls"`
	Errors       int     `json:"errors"`
	Turns        int     `json:"turns"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CacheRead    int     `json:"cache_read_tokens"`
	CacheWrite   int     `json:"cache_write_tokens"`
	CostUSD      float64 `json:"cost_usd"`
	DurationMS   int     `json:"duration_ms"`
}

// ClipLen is how much of a tool result the normalized form keeps.
const ClipLen = 4000

// Clip shortens s to ClipLen characters, marking the cut.
func Clip(s string) string {
	if len(s) <= ClipLen {
		return s
	}
	cut := ClipLen
	for cut > 0 && s[cut]&0xC0 == 0x80 { // do not split a UTF-8 sequence
		cut--
	}
	return s[:cut] + "…[clipped]"
}

// PrepSteps returns the steps that ran at least one prep command.
func (t *Transcript) PrepSteps() []Step {
	var out []Step
	for _, s := range t.Steps {
		if len(s.Commands) > 0 {
			out = append(out, s)
		}
	}
	return out
}
