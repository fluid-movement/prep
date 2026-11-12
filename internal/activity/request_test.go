package activity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRequestAttributionParsing(t *testing.T) {
	events, err := Parse(strings.NewReader(`{"kind":"request","session":"s","tokens":{"input":2}}
{"kind":"request","session":"s","agent_id":"child","agent_type":"worker","request_id":"r","turn_id":"t","status":"completed","tokens":{"input":3}}
{"kind":"gap","session":"s","agent_id":"child"}
{"kind":"summary","session":"s","cost_usd":50,"context":{"tokens":100}}
{"kind":"request","session":"s","status":"failed"}
`), 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		if err := e.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	if events[1].RequestID != "r" || events[1].AgentType != "worker" {
		t.Fatalf("payload: %+v", events[1])
	}
	r := BuildReport(events, time.Time{})
	c := r.Cost
	if c.Requests != 2 || c.Tokens.Input != 5 || c.Gaps != 2 || c.Coverage != "partial" || c.CostUSD != 50 || c.SubagentsWithRequests != 1 {
		t.Fatalf("cost: %+v", c)
	}
	if len(c.ByAgent) != 2 || c.ByAgent[0].Role != "main" || c.ByAgent[0].Tokens.Input != 2 || c.ByAgent[1].Role != "subagent" || c.ByAgent[1].Tokens.Input != 3 || c.ByAgent[1].CostUSD != 0 {
		t.Fatalf("agents: %+v", c.ByAgent)
	}
	if got := BuildReport(nil, time.Time{}).Cost.Coverage; got != "unknown" {
		t.Fatalf("empty coverage: %s", got)
	}
}

func TestReadDurableBridgeGap(t *testing.T) {
	dir := t.TempDir()
	if err := Append(dir, Event{At: reportDay.Add(time.Minute), Kind: KindRequest, Session: "s", Tokens: &Tokens{Input: 1}}); err != nil {
		t.Fatal(err)
	}
	marker := `{"at":"2026-11-10T09:00:00Z","kind":"gap","session":"s","op":"bridge-disabled"}`
	if err := os.WriteFile(filepath.Join(dir, Dir, "request-gap-s.json"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, all := range []bool{false, true} {
		var events []Event
		var err error
		if all {
			events, err = ReadAll(dir)
		} else {
			events, err = Read(dir, 0)
		}
		if err != nil {
			t.Fatal(err)
		}
		r := BuildReport(events, time.Time{})
		if len(events) != 2 || r.Cost.Gaps != 1 || r.Cost.Coverage != "partial" {
			t.Fatalf("all=%v: %+v", all, r.Cost)
		}
		if r := BuildReport(events, reportDay.Add(time.Minute)); r.Cost.Gaps != 1 || r.Cost.Coverage != "partial" || len(r.Cost.ByAgent) != 1 || r.Cost.ByAgent[0].Gaps != 1 {
			t.Fatalf("ongoing session lost its disable marker: %+v", r.Cost)
		}
	}
}

func TestReportHealthyRequestCoverage(t *testing.T) {
	dir := t.TempDir()
	events := []Event{
		{Kind: KindSession, Op: "started"},
		{Kind: KindHook, Op: "prompt", Verb: "sent"},
		{Kind: KindRequest, Tokens: &Tokens{Input: 2}},
		{Kind: KindTool, Op: "Read"},
		{Kind: KindRequest, AgentID: "child", Tokens: &Tokens{Input: 3}},
		{Kind: KindSummary, Context: &Context{Tokens: 100}, CostUSD: 0.5},
		{Kind: KindRequest, Tokens: &Tokens{Input: 2}},
		{Kind: KindRequest, AgentID: "child", Tokens: &Tokens{Input: 3}},
		{Kind: KindHook, Op: "finish", Verb: "skipped"},
		{Kind: KindRequest, Tokens: &Tokens{Input: 2}},
		{Kind: KindSession, Op: "ended"},
	}
	for _, e := range events {
		e.Session, e.Actor, e.At = "s", "claude-code", reportDay
		if err := Append(dir, e); err != nil {
			t.Fatal(err)
		}
	}
	for _, read := range []func(string) ([]Event, error){
		func(dir string) ([]Event, error) { return Read(dir, 0) }, ReadAll,
	} {
		stream, err := read(dir)
		if err != nil {
			t.Fatal(err)
		}
		c := BuildReport(stream, time.Time{}).Cost
		if c.Coverage != "full" || c.Gaps != 0 || c.Requests != 5 || c.Tokens.Input != 12 || c.SessionsWithRequests != 1 || c.SubagentsWithRequests != 1 {
			t.Fatalf("healthy coverage: %+v", c)
		}
		if len(c.ByAgent) != 2 || c.ByAgent[0].Requests != 3 || c.ByAgent[1].Requests != 2 || c.ByAgent[1].AgentID != "child" {
			t.Fatalf("agent requests: %+v", c.ByAgent)
		}
	}
}

func TestReadInterruptedRequestGap(t *testing.T) {
	dir := t.TempDir()
	// This is the per-request failure payload sent by the Claude bridge,
	// unlike the separate durable file used when the whole bridge disables.
	stream := `{"at":"2026-11-10T09:00:00Z","actor":"claude-code","session":"s","kind":"request","tokens":{"input":2}}
{"at":"2026-11-10T09:01:00Z","actor":"claude-code","session":"s","kind":"gap","op":"request-interrupted","agent_id":"child","turn_id":"t"}
`
	if err := os.MkdirAll(filepath.Join(dir, Dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(dir), []byte(stream), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, read := range []func(string) ([]Event, error){
		func(dir string) ([]Event, error) { return Read(dir, 0) }, ReadAll,
	} {
		events, err := read(dir)
		if err != nil {
			t.Fatal(err)
		}
		c := BuildReport(events, time.Time{}).Cost
		if c.Coverage != "partial" || c.Gaps != 1 || c.Requests != 1 || len(c.ByAgent) != 2 || c.ByAgent[1].AgentID != "child" || c.ByAgent[1].Gaps != 1 {
			t.Fatalf("interrupted request: %+v", c)
		}
	}
}

func TestReportMissingRequestCoverage(t *testing.T) {
	for _, tc := range []struct {
		name   string
		events []Event
		gaps   int
	}{
		{"session without requests", []Event{{Kind: KindHook, Session: "s", Op: "prompt"}}, 0},
		{"another session without requests", []Event{
			{Kind: KindRequest, Session: "s", Tokens: &Tokens{Input: 1}},
			{Kind: KindSummary, Session: "other", CostUSD: 0.5},
		}, 0},
		{"request without usage", []Event{
			{Kind: KindRequest, Session: "s", Tokens: &Tokens{Input: 1}},
			{Kind: KindRequest, Session: "s"},
		}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := BuildReport(tc.events, time.Time{}).Cost
			if c.Coverage != "partial" || c.Gaps != tc.gaps {
				t.Fatalf("missing coverage: %+v", c)
			}
		})
	}
}

func TestReportSessionlessBridgeGap(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, Dir), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := `{"at":"2026-11-10T09:00:00Z","kind":"gap","op":"bridge-disabled"}`
	if err := os.WriteFile(filepath.Join(dir, Dir, "request-gap-unknown.json"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, read := range []func(string) ([]Event, error){
		func(dir string) ([]Event, error) { return Read(dir, 0) }, ReadAll,
	} {
		events, err := read(dir)
		if err != nil {
			t.Fatal(err)
		}
		r := BuildReport(events, time.Time{})
		if r.Cost.Gaps != 1 || r.Cost.Coverage != "partial" || len(r.Cost.ByAgent) != 0 || r.Sessions != 0 || r.Unattributed != 1 {
			t.Fatalf("sessionless gap: %+v", r)
		}
	}
}

func TestReportSinceBridgeDisable(t *testing.T) {
	since := reportDay.Add(time.Hour)
	events := []Event{
		{At: reportDay, Kind: KindGap, Op: "bridge-disabled", Session: "ongoing"},
		{At: reportDay, Kind: KindGap, Op: "bridge-disabled", Session: "finished"},
		{At: reportDay, Kind: KindGap, Op: "bridge-disabled"},
		{At: reportDay, Kind: KindGap, Op: "request-interrupted", Session: "ongoing"},
		{At: reportDay, Kind: KindRequest, Session: "ongoing", Tokens: &Tokens{Input: 100}},
		{At: since, Kind: KindTool, Session: "ongoing", Op: "Read"},
	}
	r := BuildReport(events, since)
	if r.Events != 1 || r.Sessions != 1 || r.Cost.Gaps != 1 || r.Cost.Coverage != "partial" || r.Cost.Requests != 0 || r.Cost.Tokens.Input != 0 {
		t.Fatalf("since report: %+v", r)
	}
	if len(r.Cost.ByAgent) != 1 || r.Cost.ByAgent[0].Session != "ongoing" || r.Cost.ByAgent[0].Gaps != 1 {
		t.Fatalf("gap attribution: %+v", r.Cost.ByAgent)
	}
}
