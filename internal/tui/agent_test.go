package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/watch"
)

var agentNow = time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

// agentEvents is a session that finished the format decision and is on the
// CSV writer, with its harness's tool and request events, and an older
// agent on the JSON writer.
func agentEvents(ids map[string]string) []activity.Event {
	at := func(min int) time.Time { return agentNow.Add(time.Duration(min) * time.Minute) }
	s := func(min int, kind, op, issue, verb, target, area string, chars int) activity.Event {
		return activity.Event{At: at(min), Actor: "claude-code/opus", Session: "4f1c9a2e-77", Kind: kind, Op: op, Issue: issue, Verb: verb, Target: target, Area: area, Chars: chars}
	}
	req := func(min, ctx int) activity.Event {
		return activity.Event{At: at(min), Actor: "claude-code/2.1", Session: "4f1c9a2e-77", Kind: activity.KindRequest,
			Tokens: &activity.Tokens{Input: 1200, Output: 800, CacheRead: 42_000, CacheWrite: 3000}, Context: &activity.Context{Tokens: ctx, Window: 200_000}, CostUSD: 0.04}
	}
	return []activity.Event{
		{At: at(-40), Actor: "pi/1.1", Kind: activity.KindPrep, Op: "guide", Issue: ids["json"], Verb: "read the guide", Target: "JSON writer", Area: activity.AreaIssue, Chars: 900},
		s(-30, activity.KindPrep, "prime", "", "read the briefing", "", activity.AreaPrep, 1200),
		s(-29, activity.KindPrep, "guide", ids["format"], "read the guide", "Choose the default format", activity.AreaIssue, 1500),
		req(-29, 31_000),
		s(-25, activity.KindPrep, "decide", ids["format"], "recorded decision D1: CSV", "Choose the default format", activity.AreaIssue, 0),
		s(-24, activity.KindPrep, "complete", ids["format"], "completed the issue", "Choose the default format", activity.AreaIssue, 0),
		s(-10, activity.KindPrep, "guide", ids["csv"], "read the guide", "CSV writer", activity.AreaIssue, 1800),
		s(-9, activity.KindPrep, "knowledge find", "", "searched knowledge", "csv quoting", activity.AreaKnowledge, 2400),
		s(-8, activity.KindPrep, "knowledge show", "", "read knowledge", "/components/cli.md#write", activity.AreaKnowledge, 6800),
		req(-8, 58_000),
		{At: at(-6), Actor: "claude-code/2.1", Session: "4f1c9a2e-77", Kind: activity.KindTool, Op: "Edit", Verb: "edited", Target: "internal/export/csv.go", Area: activity.AreaCode, Chars: 300},
		{At: at(-5), Actor: "claude-code/2.1", Session: "4f1c9a2e-77", Kind: activity.KindTool, Op: "Bash", Verb: "ran", Target: "go test ./internal/export/", Area: activity.AreaCode, Chars: 2100, Failed: true},
		req(-4, 74_000),
		s(-1, activity.KindPrep, "criterion", ids["csv"], "checked criterion 1", "CSV writer", activity.AreaIssue, 0),
	}
}

// agentModel opens the sample project with the activity stream; timers
// are dropped so tests never wait.
func agentModel(t *testing.T, events *[]activity.Event, w, h int) (*Model, map[string]string, *int) {
	t.Helper()
	p, ids := sample(t)
	if *events == nil {
		*events = agentEvents(ids)
	}
	loads := 0
	m := NewModel(testTheme(), Options{
		Load:     func() (*domain.Tree, error) { loads++; return p.load() },
		Activity: func() ([]activity.Event, error) { return *events, nil },
		Now:      func() time.Time { return agentNow },
	})
	m.after = func(time.Duration, func(time.Time) tea.Msg) tea.Cmd { return nil }
	m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	settle(m, m.Init())
	return m, ids, &loads
}

func TestAgentScreenSnapshots(t *testing.T) {
	for _, size := range [][2]int{{110, 28}, {80, 24}} {
		w, h := size[0], size[1]
		var events []activity.Event
		m, _, _ := agentModel(t, &events, w, h)
		run(m, "w")
		view := m.View().Content
		checkSize(t, view, w, h)
		golden(t, fmt.Sprintf("agent-%dx%d", w, h), view)
		for _, want := range []string{"claude-code/opus · 4f1c9a2e", "CSV writer", "1/2 acceptance", "checked criterion 1", "searched knowledge", "done in 3 prep commands", "Choose the default format"} {
			if !strings.Contains(ansi.Strip(view), want) {
				t.Errorf("%dx%d lacks %q", w, h, want)
			}
		}
	}
	none := []activity.Event{}
	m, _, _ := agentModel(t, &none, 110, 28)
	run(m, "w")
	checkSize(t, m.View().Content, 110, 28)
	golden(t, "agent-empty-110x28", m.View().Content)
}

func TestAgentScreenKeys(t *testing.T) {
	var events []activity.Event
	m, ids, _ := agentModel(t, &events, 110, 28)
	run(m, "w")
	if m.screen != screenAgent {
		t.Fatal("w does not open the Agent screen")
	}
	usage := ansi.Strip(m.View().Content)
	for _, want := range []string{"context", "74k/200k", "requests  3 · $0.12", "knowledge", "Largest knowledge reads", "/components/cli.md#write", "This issue", "4 commands · 1 checked · 2.3k knowledge", "active 1m", "agent 1/2"} {
		if !strings.Contains(usage, want) {
			t.Errorf("usage lacks %q:\n%s", want, usage)
		}
	}
	run(m, "tab")
	if got := ansi.Strip(m.View().Content); !strings.Contains(got, "Agent · pi/1.1") || !strings.Contains(got, "JSON writer") {
		t.Fatalf("tab does not show the other agent:\n%s", got)
	}
	run(m, "tab")
	if m.agentS.agent != "" {
		t.Fatalf("back to the first agent should follow again, pinned %q", m.agentS.agent)
	}
	run(m, "enter")
	if m.screen != screenIssues || m.selected() != ids["csv"] || m.focus != focusDetail {
		t.Fatalf("enter: screen %v, selected %s, focus %v", m.screen, m.selected(), m.focus)
	}
	run(m, "w")
	run(m, "esc")
	if m.screen != screenIssues {
		t.Fatal("esc does not leave the Agent screen")
	}
}

func TestAgentScreenFollowsTheStream(t *testing.T) {
	var events []activity.Event
	m, ids, loads := agentModel(t, &events, 110, 28)
	run(m, "w")
	before := *loads
	events = append(events, activity.Event{At: agentNow.Add(time.Second), Actor: "claude-code/opus", Session: "4f1c9a2e-77", Kind: activity.KindPrep, Op: "guide", Issue: ids["json"], Verb: "read the guide", Area: activity.AreaIssue})
	_, cmd := m.Update(changedMsg{watch.Change{Local: true}})
	settle(m, cmd)
	if *loads != before {
		t.Fatalf("an activity change reloaded the issues (%d loads)", *loads-before)
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "JSON writer") || !strings.Contains(view, "▌") {
		t.Fatalf("the new event is not shown fresh with its issue:\n%s", view)
	}
	_, cmd = m.Update(changedMsg{watch.Change{Tree: true}})
	settle(m, cmd)
	if *loads != before+1 {
		t.Fatalf("a project change did not reload the issues")
	}
}

func TestAgentStartScreen(t *testing.T) {
	p, _ := sample(t)
	m := NewModel(testTheme(), Options{Load: p.load, StartAgent: true, Activity: func() ([]activity.Event, error) { return nil, nil }})
	if m.screen != screenAgent {
		t.Fatal("StartAgent does not open the Agent screen")
	}
}
