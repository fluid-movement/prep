package activity

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

func ev(actor, op, issue string, min int) Event {
	return Event{At: t0.Add(time.Duration(min) * time.Minute), Actor: actor, Kind: KindPrep, Op: op, Issue: issue}
}

func TestAppendAndRead(t *testing.T) {
	dir := t.TempDir()
	if events, err := Read(dir, 0); err != nil || events != nil {
		t.Fatalf("missing stream: %v %v", events, err)
	}
	for k := range 5 {
		if err := Append(dir, ev("a", "guide", fmt.Sprint("I", k), k)); err != nil {
			t.Fatal(err)
		}
	}
	all, err := Read(dir, 0)
	if err != nil || len(all) != 5 || all[0].Issue != "I0" || all[4].Issue != "I4" {
		t.Fatalf("all: %+v %v", all, err)
	}
	tail, _ := Read(dir, 2)
	if len(tail) != 2 || tail[0].Issue != "I3" || tail[1].Issue != "I4" {
		t.Fatalf("tail: %+v", tail)
	}
}

func TestReadSkipsBadLines(t *testing.T) {
	dir := t.TempDir()
	Append(dir, ev("a", "guide", "I1", 0))
	f, _ := os.OpenFile(Path(dir), os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("{torn\n{}\n")
	f.Close()
	Append(dir, ev("a", "guide", "I2", 1))
	got, err := Read(dir, 0)
	if err != nil || len(got) != 2 || got[1].Issue != "I2" {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestConcurrentAppendsStayWhole(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	for w := range 8 {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for k := range 50 {
				e := ev(fmt.Sprint("actor", w), "log", "I", k)
				e.Verb = strings.Repeat("x", 250)
				if err := Append(dir, e); err != nil {
					t.Error(err)
				}
			}
		}(w)
	}
	wg.Wait()
	raw, _ := os.ReadFile(Path(dir))
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	got, _ := Read(dir, 0)
	if len(lines) != 400 || len(got) != 400 {
		t.Fatalf("lines %d, parsed %d", len(lines), len(got))
	}
}

func TestArchiveAtSizeLimit(t *testing.T) {
	dir := t.TempDir()
	e := ev("a", "log", "I", 0)
	e.Verb = strings.Repeat("y", 280)
	e.Target = strings.Repeat("z", 280)
	// ~700 bytes per line: each archive holds about 1500 lines, so 4000
	// appends archive twice on the same day and leave the rest in the stream.
	const total = 4000
	for k := range total {
		e.At = t0.Add(time.Duration(k) * time.Second)
		if err := Append(dir, e); err != nil {
			t.Fatal(err)
		}
	}
	archives, _ := filepath.Glob(filepath.Join(dir, Dir, "activity-*.jsonl"))
	if len(archives) != 2 {
		t.Fatalf("archives: %v", archives)
	}
	stamp := time.Now().Format("2006-01-02")
	want := []string{filepath.Join(dir, Dir, "activity-"+stamp+".jsonl"), filepath.Join(dir, Dir, "activity-"+stamp+"-2.jsonl")}
	if !slices.Equal(archives, want) && !slices.Equal(archives, []string{want[1], want[0]}) {
		t.Fatalf("archive names %v, want %v", archives, want)
	}
	archived := 0
	for _, path := range archives {
		info, _ := os.Stat(path)
		if info.Size() <= maxBytes {
			t.Fatalf("%s archived at %d bytes", path, info.Size())
		}
		f, _ := os.Open(path)
		events, _ := Parse(f, 0)
		f.Close()
		archived += len(events)
	}
	// Readers see only the current stream: the newest events, after the archives.
	got, _ := Read(dir, 0)
	if archived+len(got) != total || !got[len(got)-1].At.Equal(t0.Add((total-1)*time.Second)) {
		t.Fatalf("archived %d, current %d", archived, len(got))
	}
	if len(got) >= total/2 {
		t.Fatalf("current stream kept %d events; it should start empty at each archive", len(got))
	}
	if tail, _ := Read(dir, 1); len(tail) != 1 || !tail[0].At.Equal(t0.Add((total-1)*time.Second)) {
		t.Fatalf("tail: %+v", tail)
	}
	info, _ := os.Stat(Path(dir))
	if info.Size() > maxBytes+maxLine {
		t.Fatalf("size %d", info.Size())
	}
}

func TestArchiveLosesRaceQuietly(t *testing.T) {
	dir := t.TempDir()
	// Another writer already archived the full stream: the stream is gone
	// or small again, and archiving does nothing.
	if err := archive(Path(dir), time.Now()); err != nil {
		t.Fatalf("missing stream: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, Dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(dir), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := archive(Path(dir), time.Now()); err != nil {
		t.Fatalf("small stream: %v", err)
	}
	if _, err := os.Stat(Path(dir)); err != nil {
		t.Fatal("small stream was archived")
	}
}

func TestValidateHook(t *testing.T) {
	for _, e := range []Event{
		{Op: "prompt", Verb: "sent"},
		{Op: "prompt", Verb: "skipped"},
		{Op: "finish", Verb: "asked"},
		{Op: "before-tool", Verb: "denied read"},
		{Op: "before-tool", Verb: "denied write"},
	} {
		if err := e.ValidateHook(); err != nil {
			t.Errorf("%+v: %v", e, err)
		}
	}
	for _, e := range []Event{
		{Op: "stop", Verb: "sent"},
		{Op: "prompt", Verb: "denied read"},
		{Op: "before-tool", Verb: ""},
	} {
		if e.ValidateHook() == nil {
			t.Errorf("accepted %+v", e)
		}
	}
	if (Event{Kind: KindHook, Op: "prompt", Verb: "sent"}).Validate() == nil {
		t.Error("harnesses may add hook events")
	}
}

func TestLongFieldsAreClipped(t *testing.T) {
	dir := t.TempDir()
	e := ev("a", "knowledge find", "", 0)
	e.Target = strings.Repeat("w ", 5000)
	Append(dir, e)
	got, _ := Read(dir, 0)
	if len([]rune(got[0].Target)) != maxField || !strings.HasSuffix(got[0].Target, "…") {
		t.Fatalf("target %d runes", len([]rune(got[0].Target)))
	}

	// Every field a harness sends is clipped, and no line reaches maxLine.
	e = ev(strings.Repeat("ä", 5000), strings.Repeat("ö", 5000), strings.Repeat("ü", 5000), 0)
	e.Session, e.Area, e.Target = strings.Repeat("s", 5000), strings.Repeat("\u2028", 900), strings.Repeat("t", 5000)
	if err := Append(dir, e); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(Path(dir))
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if len(l)+1 >= maxLine {
			t.Fatalf("line of %d bytes", len(l)+1)
		}
	}
}

func TestFoci(t *testing.T) {
	events := []Event{
		ev("a", "guide", "I1", 0),
		ev("b", "criterion", "I2", 1),
		ev("a", "show", "I9", 2),                                   // reads do not move focus
		{At: t0.Add(3 * time.Minute), Actor: "b", Kind: KindFocus}, // cleared
		ev("a", "context", "I3", 4),
		{At: t0.Add(5 * time.Minute), Actor: "a", Kind: KindTool, Op: "Edit", Issue: "I8"},
		{At: t0.Add(6 * time.Minute), Actor: "c", Kind: KindFocus, Issue: "I5"},
	}
	f := Foci(events)
	if len(f) != 2 || f["actor:a"].Issue != "I3" || f["actor:c"].Issue != "I5" {
		t.Fatalf("foci: %+v", f)
	}
	if a := Agents(events, t0.Add(6*time.Minute)); strings.Join(a, ",") != "actor:c,actor:a,actor:b" {
		t.Fatalf("agents: %v", a)
	}
	// A session joins the agent's prep commands and its harness's events,
	// whatever actor names they carry.
	s1 := []Event{
		{At: t0, Actor: "claude-code/opus", Session: "s1", Kind: KindPrep, Op: "guide", Issue: "I1"},
		{At: t0.Add(time.Minute), Actor: "claude-code/2.1", Session: "s1", Kind: KindTool, Op: "Edit"},
	}
	if a := Agents(s1, t0.Add(time.Minute)); len(a) != 1 || a[0] != "session:s1" || Foci(s1)["session:s1"].Issue != "I1" {
		t.Fatalf("session agent: %v %+v", a, Foci(s1))
	}
}

func TestValidate(t *testing.T) {
	ok := []Event{
		{Kind: KindTool, Op: "Edit", Area: AreaCode, Chars: 10},
		{Kind: KindRequest, Tokens: &Tokens{Input: 1}},
	}
	for _, e := range ok {
		if err := e.Validate(); err != nil {
			t.Fatalf("%+v: %v", e, err)
		}
	}
	bad := []Event{{Kind: KindPrep}, {Kind: KindTool}, {Kind: KindRequest}, {Kind: KindTool, Op: "x", Chars: -1}}
	for _, e := range bad {
		if e.Validate() == nil {
			t.Fatalf("%+v should fail", e)
		}
	}
}

func TestSessionLifecycle(t *testing.T) {
	events := []Event{
		{At: t0, Session: "s1", Kind: KindPrep, Op: "guide", Issue: "I1"},
		{At: t0.Add(time.Minute), Session: "s1", Kind: KindSession, Op: "ended"},
		{At: t0.Add(time.Minute), Session: "s2", Previous: "s1", Kind: KindSession, Op: "started"},
		{At: t0.Add(2 * time.Minute), Session: "s2", Kind: KindSession, Op: "ended"},
		{At: t0.Add(2 * time.Minute), Session: "s3", Previous: "s2", Kind: KindSession, Op: "started"},
	}
	if got := Agents(events, t0.Add(3*time.Minute)); strings.Join(got, ",") != "session:s3" {
		t.Fatalf("linked live agents: %v", got)
	}
	f := Foci(events)
	if len(f) != 1 || f["session:s3"].Issue != "I1" || f["session:s3"].Session != "s3" {
		t.Fatalf("inherited focus: %+v", f)
	}
	// A delayed predecessor event must not end the successor.
	events = append(events, Event{At: t0.Add(3 * time.Minute), Session: "s1", Kind: KindSession, Op: "ended"})
	if got := Agents(events, t0.Add(3*time.Minute)); len(got) != 1 {
		t.Fatalf("predecessor ended successor: %v", got)
	}
	events = append(events, Event{At: t0.Add(4 * time.Minute), Session: "s3", Kind: KindSession, Op: "ended"})
	// Even a late tool event cannot revive an ended session.
	events = append(events, Event{At: t0.Add(5 * time.Minute), Session: "s3", Kind: KindTool, Op: "Edit"})
	if got := Agents(events, t0.Add(5*time.Minute)); len(got) != 0 {
		t.Fatalf("ended agents: %v", got)
	}
	// Focus remains in the stream, and clearing in a successor clears the chain.
	events = append(events, Event{At: t0.Add(6 * time.Minute), Session: "s3", Kind: KindFocus})
	if got := Foci(events); len(got) != 0 {
		t.Fatalf("cleared focus: %+v", got)
	}
}

func TestQuietAgents(t *testing.T) {
	events := []Event{
		{At: t0, Actor: "legacy", Kind: KindTool, Op: "Edit"},
		{At: t0, Session: "crashed", Kind: KindSession, Op: "started"},
	}
	if got := Agents(events, t0.Add(30*time.Minute-time.Nanosecond)); len(got) != 2 {
		t.Fatalf("before cutoff: %v", got)
	}
	if got := Agents(events, t0.Add(30*time.Minute)); len(got) != 0 {
		t.Fatalf("at cutoff: %v", got)
	}
}

func TestValidateSession(t *testing.T) {
	for _, e := range []Event{
		{Kind: KindSession, Session: "s1", Op: "started"},
		{Kind: KindSession, Session: "s1", Op: "ended"},
		{Kind: KindSession, Session: "s2", Op: "started", Previous: "s1"},
	} {
		if err := e.Validate(); err != nil {
			t.Errorf("%+v: %v", e, err)
		}
	}
	for _, e := range []Event{
		{Kind: KindSession, Op: "started"},
		{Kind: KindSession, Session: "s1"},
		{Kind: KindSession, Session: "s1", Op: "stopped"},
		{Kind: KindSession, Session: "s1", Op: "ended", Previous: "s0"},
		{Kind: KindSession, Session: "s1", Op: "started", Previous: "s1"},
		{Kind: KindTool, Op: "Edit", Previous: "s1"},
	} {
		if e.Validate() == nil {
			t.Errorf("accepted %+v", e)
		}
	}
}
