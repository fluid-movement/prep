package cli

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// lockedBuffer lets the test read what prep watch writes from another goroutine.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func TestWatchStreamsChanges(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	watchContext = func() (context.Context, context.CancelFunc) { return ctx, cancel }
	t.Cleanup(func() { watchContext = defaultWatchContext })

	var out lockedBuffer
	done := make(chan int)
	go func() { done <- Main([]string{"watch", "--json", "--root", h.dir}, strings.NewReader(""), &out, &out) }()
	time.Sleep(100 * time.Millisecond) // let the watcher start

	h.newIssue("--title", "Parse CSV", "--kind", "code")
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(out.String(), `{"event":"changed"}`) {
		if time.Now().After(deadline) {
			t.Fatalf("no change line after prep new:\n%s", out.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("watch exited %d:\n%s", code, out.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watch did not stop")
	}
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if l != `{"event":"changed"}` {
			t.Fatalf("unexpected line %q", l)
		}
	}
}

func TestShowIncludesHistory(t *testing.T) {
	h := newHarness(t)
	id := h.newIssue("--title", "Parse CSV", "--kind", "code")
	h.ok("log", id, "tried encoding/csv", "--by", "claude-code/2.0")
	var s struct {
		History string `json:"history"`
	}
	h.jsonOf(&s, "show", id)
	if !strings.Contains(s.History, "tried encoding/csv") {
		t.Fatalf("history missing from show --json: %q", s.History)
	}
}
