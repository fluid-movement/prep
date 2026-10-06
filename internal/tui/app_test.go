package tui

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

var update = flag.Bool("update", false, "rewrite golden files")

func testTheme() *theme.Theme {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	r.SetHasDarkBackground(true)
	return theme.New(r)
}

// project builds a fixture through the domain and the markdown store with a
// fixed clock, so IDs and dates are stable.
type project struct {
	t   *testing.T
	dir string
	now time.Time
}

func newProject(t *testing.T) *project {
	p := &project{t: t, dir: t.TempDir(), now: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)}
	if _, err := mdstore.Open(p.dir).Init(); err != nil {
		t.Fatal(err)
	}
	return p
}

func (p *project) load() (*domain.Tree, error) {
	return domain.Load(mdstore.Open(p.dir), &okf.Store{Root: p.dir}, false)
}

func (p *project) do(plan func(*domain.Tree, time.Time) (*domain.Change, error)) string {
	p.t.Helper()
	st := mdstore.Open(p.dir)
	tr, err := domain.Load(st, &okf.Store{Root: p.dir}, false)
	if err != nil {
		p.t.Fatal(err)
	}
	p.now = p.now.Add(time.Minute)
	c, err := plan(tr, p.now)
	if err != nil {
		p.t.Fatal(err)
	}
	if _, err := st.Apply(c); err != nil {
		p.t.Fatal(err)
	}
	return c.IssueID
}

func (p *project) issue(title string, kind domain.Kind, body, parent string, deps ...string) string {
	return p.do(func(t *domain.Tree, now time.Time) (*domain.Change, error) {
		return t.PlanNew(domain.NewIssueInput{Title: title, Kind: kind, Body: body, Parent: parent, DependsOn: deps}, now)
	})
}

func (p *project) op(id string, op domain.Op, in domain.Input) {
	p.do(func(t *domain.Tree, now time.Time) (*domain.Change, error) {
		in.Actor, in.Now = "test/1", now
		return t.Plan(id, op, in)
	})
}

func (p *project) record(id string, op domain.Op, in domain.RecordInput) {
	p.do(func(t *domain.Tree, now time.Time) (*domain.Change, error) {
		in.Actor, in.Now = "test/1", now
		return t.PlanRecord(id, op, in)
	})
}

// sample is a small project with every state and a parent.
func sample(t *testing.T) (*project, map[string]string) {
	p := newProject(t)
	ids := map[string]string{}
	ids["export"] = p.issue("Export", domain.KindCode, "Export data in several formats.", "")
	ids["csv"] = p.issue("CSV writer", domain.KindCode, "Write rows as CSV with a header line.", ids["export"])
	ids["json"] = p.issue("JSON writer", domain.KindCode, "Write rows as JSON.", ids["export"], ids["csv"])
	ids["format"] = p.issue("Choose the default format", domain.KindDecision, "Pick the default export format.", "")
	ids["survey"] = p.issue("Survey export tools", domain.KindResearch, "Which tools do users export to?\n\n## Open questions\n\n- Which tools matter most?", "")

	enrich := func(id string) {
		p.op(id, domain.OpDefine, domain.Input{})
		p.record(id, domain.OpContext, domain.RecordInput{Text: "Change `internal/export/csv.go`; see [CLI](/components/cli.md)."})
		p.record(id, domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "add", Text: "writes a header"}, {Op: "add", Text: "quotes fields"}}})
	}
	enrich(ids["csv"])
	p.record(ids["csv"], domain.OpDecide, domain.RecordInput{Title: "Use encoding/csv", Text: "Standard library, handles quoting."})
	p.op(ids["csv"], domain.OpReady, domain.Input{})
	p.op(ids["csv"], domain.OpClaim, domain.Input{})
	p.record(ids["csv"], domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "check", Index: 1}}})
	p.record(ids["csv"], domain.OpLog, domain.RecordInput{Text: "header done"})
	enrich(ids["json"])
	p.op(ids["format"], domain.OpDefine, domain.Input{})
	p.record(ids["format"], domain.OpContext, domain.RecordInput{Text: "CSV or JSON."})
	p.record(ids["format"], domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "add", Text: "decided"}}})
	p.op(ids["format"], domain.OpReady, domain.Input{})
	p.op(ids["format"], domain.OpClaim, domain.Input{})
	p.record(ids["format"], domain.OpDecide, domain.RecordInput{Title: "CSV", Outcome: true})
	p.record(ids["format"], domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "check", Index: 1}}})
	p.op(ids["format"], domain.OpComplete, domain.Input{NoImpact: "fixture"})
	return p, ids
}

func keys(m *Model, ks ...string) {
	for _, k := range ks {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "pgdown":
			msg = tea.KeyMsg{Type: tea.KeyPgDown}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m.Update(msg)
	}
}

func screen(t *testing.T, p *project, w, h int) *Model {
	t.Helper()
	m := NewModel(testTheme(), Options{Load: p.load})
	m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m
}

func checkSize(t *testing.T, view string, w, h int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != h {
		t.Errorf("view has %d lines, want %d", len(lines), h)
	}
	for k, l := range lines {
		if lw := lipgloss.Width(l); lw > w {
			t.Errorf("line %d is %d wide, limit %d", k+1, lw, w)
		}
	}
}

func TestScreenSnapshots(t *testing.T) {
	p, _ := sample(t)

	wide := screen(t, p, 110, 28)
	keys(wide, "6", "down")
	checkSize(t, wide.View(), 110, 28)
	golden(t, "screen-110x28", wide.View())

	narrow := screen(t, p, 80, 24)
	keys(narrow, "6", "down", "enter")
	checkSize(t, narrow.View(), 80, 24)
	golden(t, "screen-80x24-detail", narrow.View())
}

func TestTabsAndSelection(t *testing.T) {
	p, ids := sample(t)
	m := screen(t, p, 110, 28)

	var names []string
	for _, tb := range m.tabs {
		names = append(names, tb.name)
	}
	if got := strings.Join(names, ","); got != "Attention,Actionable,In progress,To enrich,To define,All" {
		t.Fatalf("tabs in config order = %s", got)
	}
	keys(m, "3")
	if m.current().name != "In progress" || m.selected() != ids["csv"] {
		t.Fatalf("In progress tab: %s selected %s", m.current().name, m.selected())
	}
	keys(m, "tab", "tab", "tab")
	if m.current().name != "All" {
		t.Fatalf("tab cycling reached %s", m.current().name)
	}
	keys(m, "down", "down", "down", "down", "down", "down", "down")
	if m.selected() != ids["survey"] {
		t.Fatalf("cursor not clamped to the last issue: %s", m.selected())
	}
	keys(m, "g", "enter")
	if v := ansi.Strip(m.View()); m.focus != focusDetail || !strings.Contains(v, "Export data in several") {
		t.Fatalf("enter does not show the detail:\n%s", v)
	}
	keys(m, "esc")
	if m.focus != focusList {
		t.Fatal("esc does not return to the list")
	}

	// The selection follows the issue across a reload that shifts the list.
	keys(m, "down", "down")
	sel := m.selected()
	p.issue("Another early issue", domain.KindManual, "x", "")
	m.Update(loadedMsg(func() loadedMsg { tr, err := p.load(); return loadedMsg{tr, err} }()))
	if m.selected() != sel {
		t.Fatalf("selection moved from %s to %s after reload", sel, m.selected())
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "All 6") {
		t.Fatalf("tab count not updated:\n%s", v)
	}
}

func TestDetailContent(t *testing.T) {
	p, ids := sample(t)
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	doc := detailMarkdown(tr, ids["csv"])
	for _, want := range []string{
		"# CSV writer", "**in progress** · code · `" + ids["csv"] + "`",
		"- Parent: `" + ids["export"] + "` Export · open",
		"- Blocks: `" + ids["json"] + "` JSON writer · defined",
		"## Requirement", "## Context", "see CLI (`/components/cli.md`)",
		"### D1: Use encoding/csv", "- [x] **1.** writes a header", "- [ ] **2.** quotes fields",
		"## Definition of Done", "## History", "test/1: header done",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("detail lacks %q:\n%s", want, doc)
		}
	}
	if doc := detailMarkdown(tr, ids["survey"]); !strings.Contains(doc, "## Open questions\n\n- Which tools matter most?") {
		t.Errorf("open questions missing:\n%s", doc)
	}
	if doc := detailMarkdown(tr, ids["format"]); !strings.Contains(doc, "**outcome**") || !strings.Contains(doc, "No knowledge impact: fixture") {
		t.Errorf("decision outcome or resolution missing:\n%s", doc)
	}
}

func TestLoadErrorKeepsData(t *testing.T) {
	p, _ := sample(t)
	m := screen(t, p, 110, 28)
	m.Update(loadedMsg{nil, errors.New("issue.md: bad frontmatter")})
	v := ansi.Strip(m.View())
	if !strings.Contains(v, "bad frontmatter") || !strings.Contains(v, "Attention") {
		t.Fatalf("error not shown with the last data:\n%s", v)
	}
}

func TestWatchReportsChanges(t *testing.T) {
	dir := t.TempDir()
	ch, stop, err := watch(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	expect := func(what string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("no change reported after %s", what)
		}
	}
	sub := filepath.Join(dir, "issues", "20260102-090000")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	expect("creating directories")
	time.Sleep(50 * time.Millisecond) // let the watcher add the new directory
	if err := os.WriteFile(filepath.Join(sub, "issue.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	expect("writing a file in a new directory")
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/tui/... -update to create it)", err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the golden file; run prep tui to check and go test ./internal/tui/... -update if intended", name)
	}
}
