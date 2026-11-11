package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/mdstore"
)

func TestHookModelRequests(t *testing.T) {
	h := newHarness(t)
	for _, attempt := range []struct{ agent, status, usage string }{
		{"", "completed", `{"input_tokens":20,"cached_input_tokens":15,"output_tokens":3}`},
		{"child", "completed", `{"input_tokens":10,"cached_input_tokens":2,"output_tokens":4}`},
		{"child", "failed", "null"},
		{"child", "cancelled", "null"},
		{"", "completed", "null"},
	} {
		payload := fmt.Sprintf(`{"cwd":%q,"session_id":"s","hook_event_name":"PostModelRequest","model":"gpt","turn_id":"t","agent_id":%q,"agent_type":"worker","status":%q,"token_usage":%s}`, h.dir, attempt.agent, attempt.status, attempt.usage)
		out := runHook(t, "post-model-request", "codex", json.RawMessage(payload))
		if out != "" {
			t.Fatalf("model hook printed %q", out)
		}
	}
	events, err := activity.Read(filepath.Join(h.dir, mdstore.Dir), 0)
	if err != nil {
		t.Fatal(err)
	}
	var requests []activity.Event
	for _, e := range events {
		if e.Kind == activity.KindRequest || e.Kind == activity.KindGap {
			requests = append(requests, e)
		}
	}
	if len(requests) != 5 {
		t.Fatalf("events: %+v", requests)
	}
	if e := requests[1]; e.AgentID != "child" || e.AgentType != "worker" || e.TurnID != "t" || e.Status != "completed" || e.Op != "gpt" || e.Tokens.Input != 8 || e.Tokens.CacheRead != 2 {
		t.Fatalf("request: %+v", e)
	}
	r := activity.BuildReport(requests, time.Time{})
	if r.Cost.Requests != 2 || r.Cost.Gaps != 3 || r.Cost.Coverage != "partial" || len(r.Cost.ByAgent) != 2 {
		t.Fatalf("cost: %+v", r.Cost)
	}
}
