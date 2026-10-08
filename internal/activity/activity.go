// Package activity keeps the local stream of what agents do with prep:
// the commands they run, and the tool calls and requests their harnesses
// report. It is per-machine UI state under .prep/local, never a record:
// the TUI's Agent view reads it, and an agent's current issue is derived
// from it.
package activity

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Dir is the local directory under .prep, and File the stream in it.
const (
	Dir  = "local"
	File = "activity.jsonl"
)

// Event kinds.
const (
	KindPrep    = "prep"    // a prep command, written by the binary
	KindFocus   = "focus"   // prep focus: an explicit current issue, or none
	KindTool    = "tool"    // a harness tool call
	KindRequest = "request" // a harness model request
)

// Areas group what was read or changed.
const (
	AreaIssue     = "issue"
	AreaKnowledge = "knowledge"
	AreaCode      = "code"
	AreaPrep      = "prep"
	AreaOther     = "other"
)

// Event is one line of the stream. Kind decides which fields are set.
type Event struct {
	At      time.Time `json:"at"`
	Actor   string    `json:"actor"`
	Kind    string    `json:"kind"`
	Session string    `json:"session,omitempty"` // the harness session, when one reports it
	// Op is the prep command (guide, criterion, knowledge find, ...) or the
	// harness tool (Bash, Edit, read, ...).
	Op      string   `json:"op,omitempty"`
	Issue   string   `json:"issue,omitempty"`
	Verb    string   `json:"verb,omitempty"` // what happened, in plain words
	Target  string   `json:"target,omitempty"`
	Area    string   `json:"area,omitempty"`
	Chars   int      `json:"chars,omitempty"` // output read back, in characters
	Failed  bool     `json:"failed,omitempty"`
	Tokens  *Tokens  `json:"tokens,omitempty"`
	Context *Context `json:"context,omitempty"`
	CostUSD float64  `json:"cost_usd,omitempty"`
}

// Tokens are one request's token counts.
type Tokens struct {
	Input      int `json:"input"`
	Output     int `json:"output"`
	CacheRead  int `json:"cache_read"`
	CacheWrite int `json:"cache_write"`
}

// Context is how full the model's context window was after a request.
type Context struct {
	Tokens int `json:"tokens"`
	Window int `json:"window,omitempty"`
}

// MovesFocus reports whether the event sets its actor's current issue:
// prep focus, and every prep command naming an issue except plain reads.
func (e Event) MovesFocus() bool {
	switch e.Kind {
	case KindFocus:
		return true
	case KindPrep:
		return e.Issue != "" && e.Op != "show"
	}
	return false
}

// Validate checks an event a harness adds.
func (e Event) Validate() error {
	var why []string
	if e.Kind != KindTool && e.Kind != KindRequest {
		why = append(why, fmt.Sprintf("kind must be %s or %s, not %q", KindTool, KindRequest, e.Kind))
	}
	if e.Kind == KindTool && e.Op == "" {
		why = append(why, "a tool event needs op (the tool's name)")
	}
	if e.Kind == KindRequest && e.Tokens == nil && e.Context == nil && e.CostUSD == 0 {
		why = append(why, "a request event needs tokens, context or cost_usd")
	}
	if e.Chars < 0 || e.CostUSD < 0 {
		why = append(why, "chars and cost_usd cannot be negative")
	}
	if len(why) > 0 {
		return errors.New(strings.Join(why, "; "))
	}
	return nil
}

// Limits: lines stay below the size a single append writes atomically
// (PIPE_BUF is 4 KiB on Linux and macOS), so concurrent writers never
// interleave; past maxBytes the stream keeps its newest keepLines, at
// most half of maxBytes.
const (
	maxLine   = 4000
	maxField  = 300
	maxBytes  = 1 << 20
	keepLines = 2000
)

// Path is the stream's file under the .prep directory.
func Path(prepDir string) string { return filepath.Join(prepDir, Dir, File) }

// Append adds one event to the stream, creating .prep/local when needed.
func Append(prepDir string, e Event) error {
	e.Verb, e.Target = clip(e.Verb), clip(e.Target)
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if len(line) >= maxLine {
		e.Target = ""
		if line, err = json.Marshal(e); err != nil {
			return err
		}
	}
	line = append(line, '\n')
	path := Path(prepDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(line)
	info, statErr := f.Stat()
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && statErr == nil && info.Size() > maxBytes {
		err = prune(path)
	}
	return err
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxField {
		return string(r[:maxField-1]) + "…"
	}
	return s
}

// prune rewrites the stream with its newest lines. Two writers pruning at
// once both write a valid file; an append racing the rename may be lost,
// which a UI stream tolerates.
func prune(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	// Keep the newest lines that fit in half the limit, so pruning runs
	// once per many appends rather than on each.
	keep, size := 0, 0
	for k := len(lines) - 1; k >= 0 && keep < keepLines && size+len(lines[k])+1 <= maxBytes/2; k-- {
		keep, size = keep+1, size+len(lines[k])+1
	}
	lines = lines[len(lines)-keep:]
	tmp, err := os.CreateTemp(filepath.Dir(path), ".activity-*")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(tmp)
	for _, l := range lines {
		w.Write(l)
		w.WriteByte('\n')
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func readLines(path string) ([][]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	for l := range bytes.SplitSeq(b, []byte("\n")) {
		if len(bytes.TrimSpace(l)) > 0 {
			out = append(out, l)
		}
	}
	return out, nil
}

// Read returns the stream's events, oldest first, at most max of the
// newest (0: all). A missing stream is empty; lines that do not parse are
// skipped, so a torn or foreign line never hides the rest.
func Read(prepDir string, max int) ([]Event, error) {
	f, err := os.Open(Path(prepDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, max)
}

// Parse reads events from JSON lines; see Read.
func Parse(r io.Reader, max int) ([]Event, error) {
	var out []Event
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Kind == "" {
			continue
		}
		out = append(out, e)
		if max > 0 && len(out) > 2*max {
			out = append(out[:0], out[len(out)-max:]...)
		}
	}
	if max > 0 && len(out) > max {
		out = out[len(out)-max:]
	}
	return out, sc.Err()
}

// Focus is an actor's current issue and when it was set.
type Focus struct {
	Actor string    `json:"actor"`
	Issue string    `json:"issue"`
	At    time.Time `json:"at"`
}

// Foci returns each actor's current issue from events (oldest first);
// actors whose latest focus event cleared it are left out.
func Foci(events []Event) map[string]Focus {
	out := map[string]Focus{}
	for _, e := range events {
		if !e.MovesFocus() {
			continue
		}
		if e.Issue == "" {
			delete(out, e.Actor)
			continue
		}
		out[e.Actor] = Focus{Actor: e.Actor, Issue: e.Issue, At: e.At}
	}
	return out
}

// Actors lists the actors in events, the most recently active first.
func Actors(events []Event) []string {
	seen := map[string]bool{}
	var out []string
	for k := len(events) - 1; k >= 0; k-- {
		if a := events[k].Actor; !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}
