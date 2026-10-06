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
	ids["schema"] = p.issue("JSON schema", domain.KindCode, "Publish a JSON schema for the output.", ids["json"])
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
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "left":
			msg = tea.KeyMsg{Type: tea.KeyLeft}
		case "end":
			msg = tea.KeyMsg{Type: tea.KeyEnd}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m.Update(msg)
	}
}

func openModel(t *testing.T, p *project, w, h int) *Model {
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

	wide := openModel(t, p, 110, 28)
	keys(wide, "6", "down")
	checkSize(t, wide.View(), 110, 28)
	golden(t, "screen-110x28", wide.View())

	narrow := openModel(t, p, 80, 24)
	keys(narrow, "6", "down", "enter")
	checkSize(t, narrow.View(), 80, 24)
	golden(t, "screen-80x24-detail", narrow.View())
}

func TestTabsAndSelection(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 110, 28)

	var names []string
	for _, tb := range m.tabs {
		names = append(names, tb.name)
	}
	if got := strings.Join(names, ","); got != "Attention,Actionable,In progress,To enrich,To define,All" {
		t.Fatalf("tabs in config order = %s", got)
	}
	keys(m, "3", "down")
	if m.current().name != "In progress" || m.selected() != ids["csv"] {
		t.Fatalf("In progress tab: %s selected %s", m.current().name, m.selected())
	}
	keys(m, "tab", "tab", "tab")
	if m.current().name != "All" {
		t.Fatalf("tab cycling reached %s", m.current().name)
	}
	keys(m, "down", "down", "down", "down", "down", "down", "down", "down")
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
	if v := ansi.Strip(m.View()); !strings.Contains(v, "All 7") {
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
	m := openModel(t, p, 110, 28)
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

func rowIDs(m *Model) []string {
	var out []string
	for _, r := range m.current().rows {
		out = append(out, r.id)
	}
	return out
}

func TestHierarchy(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6")

	// Tree mode by default: children under their parent, with tree lines.
	want := []string{ids["export"], ids["csv"], ids["json"], ids["schema"], ids["format"], ids["survey"]}
	if got := rowIDs(m); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("tree order = %v, want %v", got, want)
	}
	var prefixes []string
	for _, r := range m.current().rows {
		prefixes = append(prefixes, r.prefix)
	}
	if got := strings.Join(prefixes, "|"); got != "|├─ |└─ |   └─ ||" {
		t.Fatalf("tree prefixes = %q", got)
	}

	// A view matching a child shows its parent as dimmed context, uncounted.
	keys(m, "3")
	tb := m.current()
	if tb.count != 1 || len(tb.rows) != 2 || !tb.rows[0].context || tb.rows[0].id != ids["export"] || tb.rows[1].context {
		t.Fatalf("In progress rows = %+v, count %d", tb.rows, tb.count)
	}

	// Focusing parents narrows the list and shows a breadcrumb; left goes up.
	keys(m, "6", "g", "right")
	if got := rowIDs(m); strings.Join(got, ",") != strings.Join([]string{ids["export"], ids["csv"], ids["json"], ids["schema"]}, ",") {
		t.Fatalf("focused Export rows = %v", got)
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "All › Export") {
		t.Fatalf("breadcrumb missing:\n%s", v)
	}
	keys(m, "down", "down", "right")
	if got := rowIDs(m); len(got) != 2 || got[0] != ids["json"] || got[1] != ids["schema"] {
		t.Fatalf("nested focus rows = %v", got)
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "All › Export › JSON writer") {
		t.Fatalf("nested breadcrumb missing:\n%s", v)
	}
	keys(m, "left")
	if m.selected() != ids["json"] || len(m.current().rows) != 4 {
		t.Fatalf("left did not return to Export with JSON writer selected: %s", m.selected())
	}
	keys(m, "esc")
	if m.selected() != ids["export"] || len(m.current().rows) != 6 {
		t.Fatalf("esc did not return to All with Export selected: %s", m.selected())
	}
	// right on an issue without children opens the detail.
	keys(m, "end", "right")
	if m.focus != focusDetail {
		t.Fatal("right on a leaf does not open the detail")
	}
	keys(m, "esc")

	// t toggles flat: issues in ID order, no tree lines.
	keys(m, "t")
	if got := rowIDs(m); got[2] != ids["json"] || m.current().rows[1].prefix != "" {
		t.Fatalf("flat rows = %v", got)
	}
	keys(m, "t")

	// p jumps to the parent; outside the current view it switches to All.
	keys(m, "g", "down", "down", "down", "p")
	if m.selected() != ids["json"] {
		t.Fatalf("p selected %s, want JSON writer", m.selected())
	}
	keys(m, "t", "3", "p")
	if m.current().name != "All" || m.selected() != ids["export"] {
		t.Fatalf("p outside the view: tab %s selected %s", m.current().name, m.selected())
	}
	keys(m, "t", "p")
	if m.notice == "" {
		t.Fatal("p on a top-level issue gives no notice")
	}

	// Relations in the detail: tab moves, enter opens, backspace returns.
	keys(m, "6", "g", "down", "enter")
	var got []string
	for _, r := range m.relations(m.selected()) {
		got = append(got, r.label+":"+r.id)
	}
	wantRels := []string{"parent:" + ids["export"], "blocks:" + ids["json"]}
	if strings.Join(got, ",") != strings.Join(wantRels, ",") {
		t.Fatalf("relations = %v, want %v", got, wantRels)
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "Related 1/2") {
		t.Fatalf("relations block missing:\n%s", v)
	}
	keys(m, "tab", "enter")
	if m.selected() != ids["json"] || m.focus != focusDetail {
		t.Fatalf("enter on a relation selected %s", m.selected())
	}
	keys(m, "backspace")
	if m.selected() != ids["csv"] {
		t.Fatalf("backspace returned to %s", m.selected())
	}
}

func typeText(m *Model, s string) {
	for _, r := range s {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestFilterBar(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6", "t", "/")
	if !m.filtering {
		t.Fatal("/ does not open the filter bar")
	}
	typeText(m, "--kind decision")
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 1 || got[0] != ids["format"] {
		t.Fatalf("--kind decision rows = %v", got)
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "/ --kind decision") {
		t.Fatalf("filter missing from the pane title:\n%s", v)
	}

	// Bare words match titles; the filter narrows the tab, never widens it.
	keys(m, "/")
	m.input.SetValue("")
	typeText(m, "writer")
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 2 || got[0] != ids["csv"] || got[1] != ids["json"] {
		t.Fatalf("text filter rows = %v", got)
	}
	keys(m, "3")
	keys(m, "/")
	typeText(m, "--state defined")
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 0 {
		t.Fatalf("filter widened In progress: %v", got)
	}

	// A parse error keeps the bar open; esc clears the filter.
	keys(m, "/")
	m.input.SetValue("--bogus")
	keys(m, "enter")
	if !m.filtering || m.filterErr == "" {
		t.Fatal("invalid filter accepted")
	}
	keys(m, "esc")
	if m.filtering || m.filters["In progress"] != "" || len(rowIDs(m)) != 1 {
		t.Fatalf("esc did not clear the filter: %v", rowIDs(m))
	}
}

func TestStaleDiff(t *testing.T) {
	p, ids := sample(t)
	p.do(func(tr *domain.Tree, now time.Time) (*domain.Change, error) {
		body := "Write rows as JSON lines."
		kind := domain.KindManual
		return tr.PlanEdit(ids["json"], domain.EditInput{Actor: "test/1", Now: now, Body: &body, Kind: &kind})
	})
	m := openModel(t, p, 120, 40)
	keys(m, "6")
	m.selectInCurrent(ids["json"])
	v := ansi.Strip(m.View())
	for _, want := range []string{"Changed since baseline", "- kind: code", "+ kind: manual", "- Write rows as JSON.", "+ Write rows as JSON lines."} {
		if !strings.Contains(v, want) {
			t.Fatalf("stale diff lacks %q:\n%s", want, v)
		}
	}
}

func TestCheckAndSettingsScreens(t *testing.T) {
	p, ids := sample(t)
	if err := os.WriteFile(filepath.Join(p.dir, ".prep", "issues", ids["survey"], "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls := 0
	m := NewModel(testTheme(), Options{Load: p.load, Check: func() ([]domain.Diagnostic, error) {
		calls++
		tr, err := p.load()
		if err != nil {
			return nil, err
		}
		return domain.Validate(tr), nil
	}})
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	cmd := m.key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if m.screen != screenCheck || cmd == nil {
		t.Fatal("c does not open the check screen")
	}
	m.Update(cmd())
	v := ansi.Strip(m.View())
	if calls != 1 || !strings.Contains(v, "Errors (1)") || !strings.Contains(v, "notes.txt") || !strings.Contains(v, "✕ 1") {
		t.Fatalf("check screen (calls %d):\n%s", calls, v)
	}
	keys(m, "esc")
	if m.screen != screenIssues {
		t.Fatal("esc does not leave the check screen")
	}

	keys(m, "s")
	v = ansi.Strip(m.View())
	for _, want := range []string{"Settings", "Commit mode", "Attention", "--stale", "Definition of Done"} {
		if !strings.Contains(v, want) {
			t.Fatalf("settings lacks %q:\n%s", want, v)
		}
	}
	keys(m, "s")
	if m.screen != screenIssues {
		t.Fatal("s does not toggle the settings screen")
	}
}
