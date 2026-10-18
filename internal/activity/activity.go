// Package activity keeps the local stream of what agents do with prep:
// the commands they run, and the tool calls and requests their harnesses
// report. It is per-machine UI state under .prep/local, never a record:
// the TUI's Agent view reads it, and an agent's current issue is derived
// from it.
package activity

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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
	KindSession = "session" // a harness session started or ended
	KindHook    = "hook"    // a prep hook's outcome, written by the binary
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
	At       time.Time `json:"at"`
	Actor    string    `json:"actor"`
	Kind     string    `json:"kind"`
	Previous string    `json:"previous,omitempty"` // predecessor session, on started events
	Session  string    `json:"session,omitempty"`  // the harness session, when one reports it
	// Op is the prep command (guide, criterion, knowledge find, ...) or the
	// harness tool (Bash, Edit, read, ...).
	Op     string `json:"op,omitempty"`
	Issue  string `json:"issue,omitempty"`
	Verb   string `json:"verb,omitempty"` // what happened, in plain words
	Target string `json:"target,omitempty"`
	Area   string `json:"area,omitempty"`
	// Project names the other project a prep command ran in (--project),
	// on the copy of its event in the stream of the project it was started
	// from; such events never move focus.
	Project string   `json:"project,omitempty"`
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
// prep focus, and every prep command naming an issue except plain reads,
// unless it ran in another project.
func (e Event) MovesFocus() bool {
	if e.Project != "" || e.Failed {
		return false
	}
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
	if e.Kind != KindTool && e.Kind != KindRequest && e.Kind != KindSession {
		why = append(why, fmt.Sprintf("kind must be %s, %s or %s, not %q", KindTool, KindRequest, KindSession, e.Kind))
	}
	if e.Kind == KindTool && e.Op == "" {
		why = append(why, "a tool event needs op (the tool's name)")
	}
	if e.Kind == KindRequest && e.Tokens == nil && e.Context == nil && e.CostUSD == 0 {
		why = append(why, "a request event needs tokens, context or cost_usd")
	}
	if e.Kind == KindSession {
		if strings.TrimSpace(e.Session) == "" {
			why = append(why, "a session event needs session")
		}
		if e.Op != "started" && e.Op != "ended" {
			why = append(why, "a session event needs op started or ended")
		}
	}
	if e.Previous != "" && (e.Kind != KindSession || e.Op != "started" || strings.TrimSpace(e.Previous) == "" || e.Previous == e.Session) {
		why = append(why, "previous must name a different session on a session started event")
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
// interleave; past maxBytes the stream is archived and a new one starts.
const (
	maxLine  = 4000
	maxField = 300
	maxBytes = 1 << 20
)

// Path is the stream's file under the .prep directory.
func Path(prepDir string) string { return filepath.Join(prepDir, Dir, File) }

// Append adds one event to the stream, creating .prep/local when needed.
func Append(prepDir string, e Event) error {
	for _, f := range []*string{&e.Actor, &e.Kind, &e.Session, &e.Previous, &e.Op, &e.Issue, &e.Verb, &e.Target, &e.Area, &e.Project} {
		*f = clip(*f)
	}
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
	if len(line) >= maxLine {
		return fmt.Errorf("activity event of %d bytes is too large for one atomic append", len(line))
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
		err = archive(path, time.Now())
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

// archive moves a full stream aside as activity-<local date>.jsonl (with
// -<n> added when that name exists) so the next append starts a new stream.
// Rename is atomic: two writers archiving at once leave one archive, and a
// writer that finds the stream already moved does nothing. An append racing
// the rename may land in the archive, which a UI stream tolerates. Archives
// are never read or pruned by prep.
func archive(path string, now time.Time) error {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size() <= maxBytes {
		return nil // another writer archived it already
	}
	dir, stamp := filepath.Dir(path), now.Format("2006-01-02")
	dest := filepath.Join(dir, "activity-"+stamp+".jsonl")
	for n := 2; ; n++ {
		if _, err := os.Stat(dest); errors.Is(err, fs.ErrNotExist) {
			break
		}
		dest = filepath.Join(dir, fmt.Sprintf("activity-%s-%d.jsonl", stamp, n))
	}
	if err := os.Rename(path, dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// ValidateHook checks an event the binary records for a hook outcome.
// Harnesses cannot add these: prep activity add rejects the hook kind.
func (e Event) ValidateHook() error {
	verbs, ok := hookVerbs[e.Op]
	if !ok {
		return fmt.Errorf("hook op must be prompt, finish, before-tool or session-start, not %q", e.Op)
	}
	if !slices.Contains(verbs, e.Verb) {
		return fmt.Errorf("%s verb must be one of %v, not %q", e.Op, verbs, e.Verb)
	}
	return nil
}

// hookVerbs are the outcomes each hook records.
var hookVerbs = map[string][]string{
	"prompt":        {"sent", "skipped"},
	"finish":        {"asked"},
	"before-tool":   {"denied read", "denied write"},
	"session-start": {"primed"},
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

// Agent names who an event belongs to: its harness session when it has
// one (PREP_SESSION, set by the harness for the agent's shell and sent by
// its bridge, so both sides meet whatever actor name the agent passes),
// else its actor.
func (e Event) Agent() string {
	if e.Session != "" {
		return "session:" + e.Session
	}
	return "actor:" + e.Actor
}

// Focus is an agent's current issue and when it was set.
type Focus struct {
	Agent   string    `json:"agent"`
	Actor   string    `json:"actor"`
	Session string    `json:"session,omitempty"`
	Issue   string    `json:"issue"`
	At      time.Time `json:"at"`
}

// Foci returns each agent's current issue from events (oldest first);
// agents whose latest focus event cleared it are left out.
func Foci(events []Event) map[string]Focus {
	out := map[string]Focus{}
	keys := AgentKeys(events)
	for _, e := range events {
		if !e.MovesFocus() {
			continue
		}
		if e.Issue == "" {
			delete(out, keys[e.Agent()])
			continue
		}
		key := keys[e.Agent()]
		session := ""
		if strings.HasPrefix(key, "session:") {
			session = strings.TrimPrefix(key, "session:")
		}
		out[key] = Focus{Agent: key, Actor: e.Actor, Session: session, Issue: e.Issue, At: e.At}
	}
	return out
}

// AgentKeys maps every event's agent key to the newest linked session key.
// Links are only taken from started events; actor-only events stay separate.
func AgentKeys(events []Event) map[string]string {
	parent := map[string]string{}
	var root func(string) string
	root = func(key string) string {
		if next, ok := parent[key]; ok && next != key {
			parent[key] = root(next)
			return parent[key]
		}
		parent[key] = key
		return key
	}
	for _, e := range events {
		root(e.Agent())
		if e.Kind == KindSession && e.Op == "started" && e.Previous != "" && e.Session != "" {
			old, current := root("session:"+e.Previous), root(e.Agent())
			if old != current {
				parent[old] = current
			}
		}
	}
	for key := range parent {
		root(key)
	}
	return parent
}

const quietAfter = 30 * time.Minute

// Agents lists live agents, most recently active first. Ended sessions and
// agents with no events for 30 minutes are hidden; linked IDs count once.
func Agents(events []Event, now time.Time) []string {
	keys := AgentKeys(events)
	ended := map[string]bool{}
	for _, e := range events {
		if e.Kind == KindSession {
			ended[e.Agent()] = e.Op == "ended"
		}
	}
	seen := map[string]bool{}
	var out []string
	for k := len(events) - 1; k >= 0; k-- {
		e := events[k]
		a := keys[e.Agent()]
		if !seen[a] {
			seen[a] = true
			if !ended[a] && now.Sub(e.At) < quietAfter {
				out = append(out, a)
			}
		}
	}
	return out
}
