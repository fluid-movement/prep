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

// writer runs a plan like the CLI does: fresh store, plan, CheckWrite, Apply.
func (p *project) writer() func(func(*domain.Tree) (*domain.Change, error)) (string, error) {
	return func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		st := mdstore.Open(p.dir)
		tr, err := domain.Load(st, &okf.Store{Root: p.dir}, false)
		if err != nil {
			return "", err
		}
		c, err := plan(tr)
		if err != nil {
			return "", err
		}
		if err := tr.CheckWrite(c); err != nil {
			return "", err
		}
		_, err = st.Apply(c)
		return c.IssueID, err
	}
}

// editable opens a model that writes to the fixture and whose editor
// replaces the file's text with the next queued edit.
func editable(t *testing.T, p *project, edits *[]string) *Model {
	t.Helper()
	m := NewModel(testTheme(), Options{Load: p.load, Write: p.writer(), Actor: "human:tester"})
	m.runEditor = func(path string, done func(error) tea.Msg) tea.Cmd {
		return func() tea.Msg {
			if len(*edits) > 0 {
				os.WriteFile(path, []byte((*edits)[0]), 0o644)
				*edits = (*edits)[1:]
			}
			return done(nil)
		}
	}
	m.after = func(time.Duration, func(time.Time) tea.Msg) tea.Cmd { return nil }
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 34})
	return m
}

// run sends a key and runs the commands it returns until they settle,
// skipping timers, the way the Bubble Tea runtime would.
func run(m *Model, k string) {
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
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "space":
		msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	_, cmd := m.Update(msg)
	settle(m, cmd)
}

func settle(m *Model, cmd tea.Cmd) {
	queue := []tea.Cmd{cmd}
	for n := 0; len(queue) > 0 && n < 50; n++ {
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		done := make(chan tea.Msg, 1)
		go func() { done <- c() }()
		var msg tea.Msg
		select {
		case msg = <-done:
		case <-time.After(5 * time.Second):
			panic("a command did not finish")
		}
		switch msg := msg.(type) {
		case nil:
		case tea.BatchMsg:
			queue = append(queue, msg...)
		default:
			_, next := m.Update(msg)
			queue = append(queue, next)
		}
	}
}

func typeIn(m *Model, s string) {
	for _, r := range s {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		settle(m, cmd)
	}
}

func issue(t *testing.T, p *project, id string) *domain.Issue {
	t.Helper()
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	return tr.Issues[id]
}

func TestActionMenu(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["survey"])
	run(m, "a")
	if m.modal == nil || m.modal.kind != modalMenu {
		t.Fatal("a does not open the action menu")
	}
	v := ansi.Strip(m.View())
	for _, want := range []string{"Actions · 090600 Survey export tools", "Define", "the Open questions section in issue.md is not empty", "Complete"} {
		if !strings.Contains(v, want) {
			t.Fatalf("menu lacks %q:\n%s", want, v)
		}
	}
	run(m, "d")
	if m.modal == nil || !strings.Contains(m.notice, "Open questions section") {
		t.Fatalf("unavailable define did not explain itself or closed the menu: %q", m.notice)
	}
	run(m, "esc")

	m.selectInCurrent(ids["csv"])
	run(m, "a")
	for _, it := range m.modal.items {
		if it.key == "f" && !strings.Contains(it.reason, "agent") {
			t.Fatalf("complete on a code issue: reason %q", it.reason)
		}
	}
	run(m, "esc")
}

func TestSetPriority(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	run(m, "a")
	run(m, "i")
	if m.modal == nil || m.modal.heading != "Priority" {
		t.Fatal("i does not open the priority menu")
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "Priority · ") || !strings.Contains(v, "Medium (current)") {
		t.Fatalf("priority menu:\n%s", v)
	}
	run(m, "c")
	if got := issue(t, p, ids["csv"]).Priority; got != domain.PriorityCritical {
		t.Fatalf("priority = %q", got)
	}
	v := ansi.Strip(m.View())
	if !strings.Contains(v, "!crit") || !strings.Contains(v, "priority critical") {
		t.Fatalf("row mark or detail missing:\n%s", v)
	}
	// Medium unsets the priority again.
	run(m, "a")
	run(m, "i")
	run(m, "m")
	if got := issue(t, p, ids["csv"]).Priority; got != "" {
		t.Fatalf("medium did not unset: %q", got)
	}
}

func TestCreateRenameAndEdit(t *testing.T) {
	p, ids := sample(t)
	edits := []string{"Export rows as XML.\n\n## Open questions"}
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["export"])
	run(m, "right") // focus Export: new issues become its children
	run(m, "n")
	typeIn(m, "XML writer")
	run(m, "tab")
	run(m, "right") // kind: manual
	run(m, "enter")
	if m.modal != nil {
		t.Fatalf("create failed: %s", m.modal.err)
	}
	id := m.selected()
	i := issue(t, p, id)
	if i == nil || i.Title != "XML writer" || i.Kind != domain.KindManual || i.Parent != ids["export"] {
		t.Fatalf("created issue = %+v (selected %s)", i, id)
	}

	run(m, "e")
	if got := issue(t, p, id).Prose; got != "Export rows as XML." {
		t.Fatalf("requirement after editor = %q", got)
	}
	edits = append(edits, issue(t, p, id).Body)
	run(m, "e")
	if m.notice != "no changes" {
		t.Fatalf("unchanged editor text: notice %q", m.notice)
	}

	run(m, "a")
	run(m, "t")
	for range "XML writer" {
		m.modal.inputs[0], _ = m.modal.inputs[0].Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	typeIn(m, "XML export")
	run(m, "enter")
	if got := issue(t, p, id).Title; got != "XML export" {
		t.Fatalf("title = %q", got)
	}
	if !strings.Contains(issue(t, p, id).History, "edited by human:tester: title") {
		t.Fatalf("history lacks the human edit:\n%s", issue(t, p, id).History)
	}

	// Context goes through the record command.
	edits = append(edits, "Use `encoding/xml`.")
	run(m, "a")
	run(m, "c")
	if got := issue(t, p, id).Context; got != "Use `encoding/xml`." {
		t.Fatalf("context = %q", got)
	}
}

func TestRejectedEditKeepsText(t *testing.T) {
	p, ids := sample(t)
	edits := []string{"New text"}
	m := editable(t, p, &edits)
	writes := 0
	inner := m.opts.Write
	m.opts.Write = func(plan func(*domain.Tree) (*domain.Change, error)) (string, error) {
		writes++
		if writes == 1 {
			return "", errors.New("disk full")
		}
		return inner(plan)
	}
	run(m, "6")
	m.selectInCurrent(ids["survey"])
	run(m, "e")
	if !strings.Contains(m.notice, "disk full") || m.pending[ids["survey"]+"|requirement"] != "New text" {
		t.Fatalf("rejected edit: notice %q pending %q", m.notice, m.pending)
	}
	var opened string
	m.runEditor = func(path string, done func(error) tea.Msg) tea.Cmd {
		return func() tea.Msg {
			b, _ := os.ReadFile(path)
			opened = string(b)
			return done(nil)
		}
	}
	run(m, "e")
	if strings.TrimSpace(opened) != "New text" {
		t.Fatalf("editor reopened with %q, want the rejected text", opened)
	}
	if got := issue(t, p, ids["survey"]).Prose; got != "New text" {
		t.Fatalf("second save did not write: %q", got)
	}
}

func TestReparentAndCriteria(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["survey"])
	run(m, "a")
	run(m, "m")
	typeIn(m, "Export")
	if len(m.modal.picks) != 2 || m.modal.picks[1] != ids["export"] {
		t.Fatalf("filtered parents = %v", m.modal.picks)
	}
	run(m, "down")
	run(m, "enter")
	if got := issue(t, p, ids["survey"]).Parent; got != ids["export"] {
		t.Fatalf("parent = %q", got)
	}
	// The issue itself and its descendants are never candidates.
	m.selectInCurrent(ids["export"])
	if c := m.parentCandidates(ids["export"], ""); strings.Contains(strings.Join(c, ","), ids["survey"]) || strings.Contains(strings.Join(c, ","), ids["export"]) {
		t.Fatalf("candidates include the issue or its subtree: %v", c)
	}
	m.selectInCurrent(ids["survey"])
	run(m, "a")
	run(m, "m")
	run(m, "enter") // top level
	if got := issue(t, p, ids["survey"]).Parent; got != "" {
		t.Fatalf("not moved to top level: %q", got)
	}

	m.selectInCurrent(ids["csv"])
	run(m, "a")
	run(m, "k")
	run(m, "space")
	run(m, "down")
	run(m, "space")
	run(m, "enter")
	c := issue(t, p, ids["csv"]).Criteria
	if c[0].Checked || !c[1].Checked {
		t.Fatalf("criteria after toggling = %+v", c)
	}
}

func TestTransitionsFromTheTUI(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")

	// A manual issue through its whole lifecycle.
	run(m, "n")
	typeIn(m, "Announce the export")
	run(m, "tab")
	run(m, "right")
	run(m, "enter")
	id := m.selected()
	edits = append(edits, "Post a note in the changelog.")
	run(m, "e")
	run(m, "a")
	run(m, "d")
	if st, _ := stateOf(t, p, id); st != domain.StateDefined {
		t.Fatalf("after define: %s", st)
	}
	p.record(id, domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "add", Text: "note posted"}}})
	m.Update(loadedMsg(func() loadedMsg { tr, err := p.load(); return loadedMsg{tr, err} }()))
	run(m, "a")
	run(m, "r")
	p.op(id, domain.OpClaim, domain.Input{})
	m.Update(loadedMsg(func() loadedMsg { tr, err := p.load(); return loadedMsg{tr, err} }()))
	run(m, "a")
	run(m, "k")
	run(m, "space")
	run(m, "enter")
	run(m, "a")
	run(m, "f")
	run(m, "enter") // no documentation decision yet
	if m.modal == nil || !strings.Contains(m.modal.err, "--docs") {
		t.Fatalf("complete without documentation: modal %+v", m.modal)
	}
	run(m, "tab")
	typeIn(m, "announcement only")
	run(m, "enter")
	if st, _ := stateOf(t, p, id); st != domain.StateDone {
		t.Fatalf("after complete: %s (modal %+v)", st, m.modal)
	}
	if r := issue(t, p, id).Resolution; r.By != "human:tester" {
		t.Fatalf("completed by %q", r.By)
	}

	// Acknowledge a stale change, then drop with a reason.
	p.do(func(tr *domain.Tree, now time.Time) (*domain.Change, error) {
		body := "Write rows as JSON, one object per row."
		return tr.PlanEdit(ids["json"], domain.EditInput{Actor: "test/1", Now: now, Body: &body})
	})
	m.Update(loadedMsg(func() loadedMsg { tr, err := p.load(); return loadedMsg{tr, err} }()))
	m.selectInCurrent(ids["json"])
	run(m, "a")
	run(m, "a")
	if _, stale := stateOf(t, p, ids["json"]); stale {
		t.Fatal("acknowledge did not clear staleness")
	}
	run(m, "a")
	run(m, "x")
	typeIn(m, "not needed")
	run(m, "enter")
	if st, _ := stateOf(t, p, ids["json"]); st != domain.StateDropped {
		t.Fatalf("after drop: %s", st)
	}
	if r := issue(t, p, ids["json"]).Resolution; r.Reason != "not needed" {
		t.Fatalf("drop reason = %q", r.Reason)
	}
}

func stateOf(t *testing.T, p *project, id string) (domain.State, bool) {
	t.Helper()
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	return tr.State(id), tr.Stale(id)
}

func TestChecklist(t *testing.T) {
	c := NewChecklist(testTheme(), "Integrate prep with", []ChecklistItem{{Label: "Claude Code", Detail: "detected", Checked: true}, {Label: "Pi"}})
	if v := ansi.Strip(c.View()); !strings.Contains(v, "[x] Claude Code  detected") || !strings.Contains(v, "[ ] Pi") {
		t.Fatalf("view:\n%s", v)
	}
	for _, k := range []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeySpace, Runes: []rune{' '}}, {Type: tea.KeyUp}, {Type: tea.KeySpace, Runes: []rune{' '}}} {
		c.Update(k)
	}
	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || c.cancelled || c.items[0].Checked || !c.items[1].Checked {
		t.Fatalf("after toggling: %+v cancelled=%v", c.items, c.cancelled)
	}
	c2 := NewChecklist(testTheme(), "x", []ChecklistItem{{Label: "a"}})
	c2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !c2.cancelled {
		t.Fatal("esc does not cancel")
	}
}
