package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/fluid-movement/prep/internal/activity"
)

type activityOut struct {
	Events []activity.Event `json:"events"`
	Focus  []activity.Focus `json:"focus"`
}

func (h *harness) activity(args ...string) activityOut {
	h.t.Helper()
	var out activityOut
	h.jsonOf(&out, append([]string{"activity", "--max", "0"}, args...)...)
	return out
}

func (h *harness) runStdin(stdin string, args ...string) (int, string) {
	var out, errb bytes.Buffer
	code := Main(append(args, "--root", h.dir), strings.NewReader(stdin), &out, &errb)
	return code, out.String() + errb.String()
}

func TestActivityRecordsCommands(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "Export", "--kind", "code", "--by", "agent/1")
	h.ok("guide", id, "--by", "agent/1")
	h.ok("criterion", id, "--add", "one", "--add", "two", "--by", "agent/1")
	h.ok("criterion", id, "--check", "1", "--by", "agent/1")
	h.ok("decide", id, "--title", "Use JSON", "--description", "Use JSON", "--body", "why", "--by", "agent/1")
	h.ok("knowledge", "find", "test", "project", "--by", "agent/1")
	other := h.newIssue("--title", "Other", "--kind", "code", "--by", "agent/2")
	h.ok("show", id, "--by", "agent/2") // a read: agent/2 stays on Other
	h.fails("no issue matches", "guide", "NOPE", "--by", "agent/1")
	h.ok("show", id) // an unnamed read (a panel, a person) is not recorded

	got := h.activity()
	var verbs []string
	for _, e := range got.Events {
		if e.Actor == "agent/1" {
			verbs = append(verbs, e.Verb)
		}
	}
	// The failed guide is recorded with its error code, after the commands that succeeded.
	want := []string{"created a code issue", "read the guide", "added 2 criteria", "checked criterion 1", "recorded decision: Use JSON", "searched knowledge", "E_NOT_FOUND"}
	if strings.Join(verbs, "|") != strings.Join(want, "|") {
		t.Fatalf("verbs:\n%v\nwant\n%v", verbs, want)
	}
	for _, e := range got.Events {
		switch e.Op {
		case "guide", "show", "knowledge find":
			if e.Failed {
				continue // a failure records its code, not output
			}
			if e.Chars == 0 {
				t.Fatalf("%s carries no output size: %+v", e.Op, e)
			}
		case "criterion":
			if e.Chars != 0 || e.Issue != id || e.Target != "Export" || e.Area != activity.AreaIssue {
				t.Fatalf("criterion event: %+v", e)
			}
		}
	}
	for _, e := range got.Events {
		if e.Chars > 0 && e.Actor != "agent/1" && e.Actor != "agent/2" {
			t.Fatalf("unnamed read recorded: %+v", e)
		}
	}
	if len(got.Focus) != 2 || got.Focus[0].Actor != "agent/2" || got.Focus[0].Issue != other || got.Focus[1].Issue != id {
		t.Fatalf("focus: %+v", got.Focus)
	}

	// prep focus moves and clears; prime lists the newest foci.
	h.ok("focus", other[len(other)-6:], "--by", "agent/1")
	if out := h.ok("prime"); !strings.Contains(out, "Recently worked on") || !strings.Contains(out, other+" [open] Other — agent/1") {
		t.Fatalf("prime:\n%s", out)
	}
	h.ok("focus", "--clear", "--by", "agent/1")
	if f := h.activity().Focus; len(f) != 1 || f[0].Actor != "agent/2" {
		t.Fatalf("after clear: %+v", f)
	}
	if out := h.ok("activity", "--actor", "agent/2"); !strings.Contains(out, "agent/2") || strings.Contains(out, "agent/1") {
		t.Fatalf("activity --actor:\n%s", out)
	}
}

func TestActivityAdd(t *testing.T) {
	h := newHarness(t)
	lines := `{"kind":"tool","op":"Edit","target":"internal/cli/cli.go","area":"code","chars":120,"session":"s1"}
{"kind":"request","tokens":{"input":10,"output":5,"cache_read":1000,"cache_write":0},"context":{"tokens":40000,"window":200000},"cost_usd":0.02,"session":"s1"}
`
	if code, out := h.runStdin(lines, "activity", "add", "--by", "claude-code/2"); code != 0 {
		t.Fatalf("add: %d\n%s", code, out)
	}
	got := h.activity()
	n := len(got.Events)
	if n < 2 || got.Events[n-2].Op != "Edit" || got.Events[n-2].Actor != "claude-code/2" || got.Events[n-1].Tokens.CacheRead != 1000 || got.Events[n-1].At.IsZero() {
		b, _ := json.Marshal(got.Events)
		t.Fatalf("events: %s", b)
	}
	for _, bad := range []string{`{"kind":"prep","op":"guide"}`, `{"kind":"tool"}`, ``} {
		if code, _ := h.runStdin(bad, "activity", "add"); code == 0 {
			t.Fatalf("accepted %q", bad)
		}
	}
	// The rejected adds fail as commands, which record failed prep events;
	// no harness event may be recorded.
	for _, e := range h.activity().Events[n:] {
		if !e.Failed || e.Kind != activity.KindPrep || e.Op != "activity" {
			t.Fatalf("rejected event recorded as %+v", e)
		}
	}
}

func TestActivityOutsideProject(t *testing.T) {
	var out, errb bytes.Buffer
	dir := t.TempDir()
	if code := Main([]string{"activity", "--root", dir}, strings.NewReader(""), &out, &errb); code == 0 {
		t.Fatalf("activity outside a project: %s", out.String())
	}
}

func TestActivitySessionAndResumedPrime(t *testing.T) {
	h := newHarness(t)
	t.Setenv("PREP_SESSION", "old")
	id := h.newIssue("--title", "Resumed work", "--kind", "code", "--by", "agent/1")
	t.Setenv("PREP_SESSION", "new")
	lines := `{"kind":"session","session":"old","op":"ended"}
{"kind":"session","session":"middle","op":"started","previous":"old"}
{"kind":"session","session":"middle","op":"ended"}
{"kind":"session","session":"new","op":"started","previous":"middle"}`
	if code, out := h.runStdin(lines, "activity", "add"); code != 0 {
		t.Fatalf("add lifecycle: %d %s", code, out)
	}
	got := h.activity()
	if len(got.Focus) != 1 || got.Focus[0].Session != "new" || got.Focus[0].Issue != id {
		t.Fatalf("linked focus: %+v", got.Focus)
	}
	if out := h.ok("prime"); strings.Count(out, id+" [open] Resumed work — agent/1") != 1 {
		t.Fatalf("resumed prime: %s", out)
	}
	for _, bad := range []string{
		`{"kind":"session","session":"new","op":"bad"}`,
		`{"kind":"session","session":"new","op":"ended","previous":"old"}`,
		`{"kind":"session","session":"new","op":"started","previous":"new"}`,
		`{"kind":"session","op":"started"}`,
	} {
		if code, _ := h.runStdin(bad, "activity", "add"); code == 0 {
			t.Errorf("accepted %s", bad)
		}
	}
}
