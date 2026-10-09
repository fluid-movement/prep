package tui

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/okf"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

var update = flag.Bool("update", false, "rewrite golden files")

func testTheme() *theme.Theme { return theme.New(true, colorprofile.TrueColor) }

// project builds a fixture through the domain and the markdown store with a
// fixed clock and seeded entropy, so IDs and dates are stable.
type project struct {
	t    *testing.T
	dir  string
	now  time.Time
	rand io.Reader
}

func newProject(t *testing.T) *project {
	p := &project{t: t, dir: t.TempDir(), now: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC), rand: rand.NewChaCha8([32]byte{})}
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
		return t.PlanNew(domain.NewIssueInput{Title: title, Kind: kind, Body: body, Parent: parent, DependsOn: deps, Entropy: p.rand}, now)
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

// sampleViews are more views than prep init writes, so tests can switch,
// filter and reorder tabs.
var sampleViews = []struct{ name, query string }{
	{"Attention", "--stale"},
	{"Actionable", "--actionable"},
	{"In progress", "--state in-progress"},
	{"To enrich", "--state defined"},
	{"To define", "--state open"},
	{"All", ""},
}

// sample is a small project with every state and a parent, and the
// sampleViews.
func sample(t *testing.T) (*project, map[string]string) {
	p := newProject(t)
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	cfg := domain.Config{Views: map[string]string{}}
	for _, v := range sampleViews {
		cfg.Views[v.name] = v.query
		cfg.ViewOrder = append(cfg.ViewOrder, v.name)
	}
	c, err := tr.PlanConfig(cfg)
	if err == nil {
		_, err = mdstore.Open(p.dir).Apply(c)
	}
	if err != nil {
		t.Fatal(err)
	}
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

// knowledgeSample is sample with a knowledge base: the CLI entry has a
// broken link (the check flags it), the format decision is documented by a
// completed issue (a backlink).
func knowledgeSample(t *testing.T) (*project, map[string]string) {
	p, ids := sample(t)
	write := func(rel, content string) {
		path := filepath.Join(p.dir, ".prep", "knowledge", rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("components/cli.md", "---\ntype: component\ntitle: CLI\ndescription: Commands, flags and output.\nstatus: stable\nscope:\n  - internal/cli\n---\n\n# CLI\n\nCommands live in `cli.go`. See [Gone](/components/gone.md).\n")
	write("decisions/format.md", "---\ntype: decision\ntitle: Default export format\ndescription: Why CSV is the default.\nstatus: draft\n---\n\n# Default export format\n\nCSV, because spreadsheets open it.\n")
	ids["doc"] = p.issue("Document the format", domain.KindManual, "Write down why CSV is the default.", "")
	p.op(ids["doc"], domain.OpDefine, domain.Input{})
	p.record(ids["doc"], domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "add", Text: "written"}}})
	p.op(ids["doc"], domain.OpReady, domain.Input{})
	p.op(ids["doc"], domain.OpClaim, domain.Input{})
	p.record(ids["doc"], domain.OpCriterion, domain.RecordInput{Acceptance: []domain.AcceptanceOp{{Op: "check", Index: 1}}})
	p.op(ids["doc"], domain.OpComplete, domain.Input{Docs: []string{"/decisions/format.md"}})
	return p, ids
}

// withCheck gives a model the check the CLI runs, without drift.
func withCheck(m *Model, p *project) *Model {
	m.opts.Check = func() ([]domain.Diagnostic, error) {
		t, err := p.load()
		if err != nil {
			return nil, err
		}
		return domain.Validate(t), nil
	}
	return m
}

func TestKnowledgeScreen(t *testing.T) {
	p, ids := knowledgeSample(t)
	m := withCheck(openModel(t, p, 120, 30), p)
	run(m, "b")
	if m.screen != screenKnowledge {
		t.Fatal("b did not open the knowledge screen")
	}
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"component  stable     ! CLI", "decision   draft        Default export format"} {
		if !strings.Contains(v, want) {
			t.Fatalf("knowledge list lacks %q:\n%s", want, v)
		}
	}

	// a lists only what needs an agent's attention: the broken link.
	run(m, "!")
	if paths, _ := m.knowledgePaths(); len(paths) != 1 || paths[0] != "/components/cli.md" {
		t.Fatalf("attention-only = %v", paths)
	}
	v = ansi.Strip(m.View().Content)
	if !strings.Contains(v, "Needs an agent's attention") || !strings.Contains(v, "K003 broken link to /components/gone.md") {
		t.Fatalf("entry lacks its finding:\n%s", v)
	}
	run(m, "!")

	// f filters by type; the entry shows its body and who changed it.
	run(m, "f")
	typeIn(m, "--type decision")
	run(m, "enter")
	if paths, _ := m.knowledgePaths(); len(paths) != 1 || m.know.path != "/decisions/format.md" {
		t.Fatalf("filtered = %v, selected %s", paths, m.know.path)
	}
	v = ansi.Strip(m.View().Content)
	if !strings.Contains(v, "CSV, because spreadsheets open it.") || !strings.Contains(v, "Document the format") {
		t.Fatalf("entry lacks body or backlink:\n%s", v)
	}
	run(m, "o")
	run(m, "1")
	if m.screen != screenIssues || m.selected() != ids["doc"] {
		t.Fatalf("o 1 went to %s on screen %d", m.selected(), m.screen)
	}
}

func TestKnowledgeFromAnIssue(t *testing.T) {
	p, ids := knowledgeSample(t)
	m := openModel(t, p, 120, 30)
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "≡ knows component CLI") {
		t.Fatalf("issue detail lacks its knowledge:\n%s", v)
	}
	run(m, "o")
	var key string
	rels := m.relations(m.selected())
	for n, k := range linkTargets(rels) {
		if rels[k].id == "/components/cli.md" {
			key = linkKeys[n : n+1]
		}
	}
	if v := ansi.Strip(m.View().Content); !m.linkMode || !strings.Contains(v, key+" ≡ knows component CLI") {
		t.Fatalf("link mode does not show key %q on the knowledge link:\n%s", key, v)
	}
	run(m, key)
	if m.screen != screenKnowledge || m.know.path != "/components/cli.md" {
		t.Fatalf("o %s: screen %d entry %s", key, m.screen, m.know.path)
	}
	run(m, "backspace")
	if m.screen != screenIssues || m.selected() != ids["csv"] {
		t.Fatalf("backspace: screen %d issue %s", m.screen, m.selected())
	}
}

func keys(m *Model, ks ...string) {
	for _, k := range ks {
		m.Update(keyMsg(k))
	}
}

// keyMsg builds the key press a terminal sends for a key name.
func keyMsg(k string) tea.KeyPressMsg {
	named := map[string]tea.KeyPressMsg{
		"enter": {Code: tea.KeyEnter}, "esc": {Code: tea.KeyEscape}, "tab": {Code: tea.KeyTab},
		"down": {Code: tea.KeyDown}, "up": {Code: tea.KeyUp}, "right": {Code: tea.KeyRight}, "left": {Code: tea.KeyLeft},
		"pgdown": {Code: tea.KeyPgDown}, "end": {Code: tea.KeyEnd}, "backspace": {Code: tea.KeyBackspace},
		"space": {Code: tea.KeySpace, Text: " "}, "shift+up": {Code: tea.KeyUp, Mod: tea.ModShift},
		"ctrl+e": {Code: 'e', Mod: tea.ModCtrl}, "ctrl+s": {Code: 's', Mod: tea.ModCtrl},
	}
	if msg, ok := named[k]; ok {
		return msg
	}
	r := []rune(k)[0]
	return tea.KeyPressMsg{Code: r, Text: k}
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
	checkSize(t, wide.View().Content, 110, 28)
	golden(t, "screen-110x28", wide.View().Content)

	narrow := openModel(t, p, 80, 24)
	keys(narrow, "6", "down", "enter")
	checkSize(t, narrow.View().Content, 80, 24)
	golden(t, "screen-80x24-detail", narrow.View().Content)

	// Dialogs float over the dimmed screen; the inline editor and the
	// editable settings at both sizes.
	for _, size := range [][2]int{{110, 28}, {80, 24}} {
		w, h := size[0], size[1]
		var edits []string
		m := editable(t, p, &edits)
		m.Update(tea.WindowSizeMsg{Width: w, Height: h})
		keys(m, "6", "down", "n")
		typeIn(m, "Stream rows")
		run(m, "enter")
		golden(t, fmt.Sprintf("dialog-create-kind-%dx%d", w, h), m.View().Content)
		run(m, "c")
		typeIn(m, "Rows go to stdout.")
		checkSize(t, m.View().Content, w, h)
		golden(t, fmt.Sprintf("dialog-create-%dx%d", w, h), m.View().Content)
		run(m, "esc")
		run(m, "esc")
		run(m, "esc")
		run(m, "?")
		checkSize(t, m.View().Content, w, h)
		golden(t, fmt.Sprintf("dialog-keys-%dx%d", w, h), m.View().Content)
		run(m, "esc")
		run(m, "e")
		run(m, "r")
		checkSize(t, m.View().Content, w, h)
		golden(t, fmt.Sprintf("dialog-requirement-%dx%d", w, h), m.View().Content)
		run(m, "esc")
		run(m, "f")
		typeText(m, "--state ")
		checkSize(t, m.View().Content, w, h)
		golden(t, fmt.Sprintf("filter-picker-%dx%d", w, h), m.View().Content)
		run(m, "esc")
		run(m, "s")
		run(m, "3")
		checkSize(t, m.View().Content, w, h)
		golden(t, fmt.Sprintf("settings-%dx%d", w, h), m.View().Content)
		kp, _ := knowledgeSample(t)
		k := withCheck(openModel(t, kp, w, h), kp)
		run(k, "b")
		checkSize(t, k.View().Content, w, h)
		golden(t, fmt.Sprintf("knowledge-%dx%d", w, h), k.View().Content)
	}
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
	if v := ansi.Strip(m.View().Content); m.focus != focusDetail || !strings.Contains(v, "Export data in several") {
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
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "All 7") {
		t.Fatalf("tab count not updated:\n%s", v)
	}
}

// A project as prep init leaves it has the Unresolved and All tabs: all
// unresolved work first, every issue second.
func TestDefaultViews(t *testing.T) {
	p := newProject(t)
	open := p.issue("Open", domain.KindCode, "x", "")
	defined := p.issue("Defined", domain.KindCode, "x", "")
	p.op(defined, domain.OpDefine, domain.Input{})
	dropped := p.issue("Dropped", domain.KindCode, "x", "")
	p.op(dropped, domain.OpDrop, domain.Input{Reason: "fixture"})
	m := openModel(t, p, 110, 28)

	var names []string
	for _, tb := range m.tabs {
		names = append(names, tb.name)
	}
	if got := strings.Join(names, ","); got != "Unresolved,All" {
		t.Fatalf("default tabs = %s", got)
	}
	if got := strings.Join(rowIDs(m), ","); got != open+","+defined {
		t.Fatalf("Unresolved rows = %s", got)
	}
	keys(m, "2")
	if got := len(rowIDs(m)); got != 3 {
		t.Fatalf("All has %d rows, want 3", got)
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
	v := ansi.Strip(m.View().Content)
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
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "All › Export") {
		t.Fatalf("breadcrumb missing:\n%s", v)
	}
	keys(m, "down", "down", "right")
	if got := rowIDs(m); len(got) != 2 || got[0] != ids["json"] || got[1] != ids["schema"] {
		t.Fatalf("nested focus rows = %v", got)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "All › Export › JSON writer") {
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

	// Relations in the detail: tab moves, enter opens, backspace returns.
	keys(m, "6", "g", "down", "enter")
	var got []string
	for _, r := range m.relations(m.selected()) {
		got = append(got, r.label+":"+r.id)
	}
	wantRels := []string{"path:" + ids["export"], "blocks:" + ids["json"]}
	if strings.Join(got, ",") != strings.Join(wantRels, ",") {
		t.Fatalf("relations = %v, want %v", got, wantRels)
	}
	// Links show by shape: the parent as a breadcrumb, what waits on this
	// issue behind an arrow, a child parent with its progress.
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "↑ Export") || !strings.Contains(v, "→ unblocks ◐ "+shortID(ids["json"])+" ■■■■■ 0/1 JSON writer") {
		t.Fatalf("relations block missing:\n%s", v)
	}
	keys(m, "o", "2") // the second link: → unblocks JSON writer
	if m.selected() != ids["json"] || m.focus != focusDetail {
		t.Fatalf("o 2 selected %s", m.selected())
	}
	keys(m, "backspace")
	if m.selected() != ids["csv"] {
		t.Fatalf("backspace returned to %s", m.selected())
	}
}

func typeText(m *Model, s string) {
	for _, r := range s {
		m.Update(keyMsg(string(r)))
	}
}

func TestFilterBar(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6", "t", "/")
	if !m.filtering {
		t.Fatal("/ does not open the filter bar")
	}
	m.Update(tea.PasteMsg{Content: "--kind decision"})
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 1 || got[0] != ids["format"] {
		t.Fatalf("--kind decision rows = %v", got)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "/ --kind decision") {
		t.Fatalf("filter missing from the pane title:\n%s", v)
	}

	// Bare words match titles; the filter narrows the tab, never widens it.
	keys(m, "f")
	m.input.SetValue("")
	typeText(m, "writer")
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 2 || got[0] != ids["csv"] || got[1] != ids["json"] {
		t.Fatalf("text filter rows = %v", got)
	}
	keys(m, "3")
	keys(m, "f")
	typeText(m, "--state defined")
	keys(m, "enter")
	if got := rowIDs(m); len(got) != 0 {
		t.Fatalf("filter widened In progress: %v", got)
	}

	// A parse error keeps the bar open; esc restores the filter the tab had,
	// and esc in the list clears it.
	keys(m, "/")
	m.input.SetValue("--bogus")
	keys(m, "enter")
	if !m.filtering || m.filterErr == "" {
		t.Fatal("invalid filter accepted")
	}
	keys(m, "esc")
	if m.filtering || m.filters["In progress"] != "--state defined" {
		t.Fatalf("esc did not restore the filter: %q", m.filters["In progress"])
	}
	keys(m, "esc")
	if m.filters["In progress"] != "" || len(rowIDs(m)) != 1 {
		t.Fatalf("esc in the list did not clear the filter: %v", rowIDs(m))
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
	v := ansi.Strip(m.View().Content)
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

	cmd := m.key(keyMsg("c"))
	if m.screen != screenCheck || cmd == nil {
		t.Fatal("c does not open the check screen")
	}
	m.Update(cmd())
	v := ansi.Strip(m.View().Content)
	if calls != 1 || !strings.Contains(v, "Errors (1)") || !strings.Contains(v, "notes.txt") || !strings.Contains(v, "✕ 1") {
		t.Fatalf("check screen (calls %d):\n%s", calls, v)
	}
	keys(m, "esc")
	if m.screen != screenIssues {
		t.Fatal("esc does not leave the check screen")
	}

	keys(m, "s")
	v = ansi.Strip(m.View().Content)
	for _, want := range []string{"Settings", "Saved views", "Attention", "--stale", "Definition of Done"} {
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

// viaEditor opens a text dialog with key, takes the next text from the
// editor hook (ctrl+e) and saves it (ctrl+s).
func viaEditor(m *Model, key ...string) {
	for _, k := range key {
		run(m, k)
	}
	run(m, "ctrl+e")
	run(m, "ctrl+s")
}

// run sends a key and runs the commands it returns until they settle,
// skipping timers, the way the Bubble Tea runtime would.
func run(m *Model, k string) {
	_, cmd := m.Update(keyMsg(k))
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
		_, cmd := m.Update(keyMsg(string(r)))
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
	// Only what applies is listed: survey has open questions, so no Define.
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"Actions · " + shortID(ids["survey"]) + " Survey export tools", "Move to another parent", "Drop"} {
		if !strings.Contains(v, want) {
			t.Fatalf("menu lacks %q:\n%s", want, v)
		}
	}
	for _, gone := range []string{"Define", "Mark ready", "Complete", "Edit requirement", "Set priority", "Edit tags", "New issue"} {
		if strings.Contains(v, gone) {
			t.Fatalf("menu lists %q, which does not apply:\n%s", gone, v)
		}
	}
	// The dialog floats over the screen: the list stays visible behind it.
	if !strings.Contains(v, "Choose the default format") {
		t.Fatalf("screen behind the dialog is gone:\n%s", v)
	}
	run(m, "esc")

	// ? lists every sequence, the ones that do not apply with their reason.
	run(m, "?")
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "a d") || !strings.Contains(v, "Define (not now") || !strings.Contains(v, "a p") {
		t.Fatalf("keymap lacks the action menu's sequences:\n%s", v)
	}
	run(m, "x") // any key closes
	if m.modal != nil {
		t.Fatal("? dialog did not close")
	}

	m.selectInCurrent(ids["csv"])
	run(m, "a")
	for _, it := range m.modal.items {
		if it.key == "f" {
			t.Fatalf("complete listed for an agent's code issue")
		}
	}
	run(m, "esc")
}

func TestEditMenuAndKeys(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["schema"])

	// e c edits the context inline; text typed in the dialog is saved with ctrl+s.
	run(m, "e")
	if m.modal == nil || m.modal.heading != "Edit" {
		t.Fatalf("e opened %+v", m.modal)
	}
	run(m, "c")
	if m.modal == nil || m.modal.kind != modalText || m.modal.field != "context" {
		t.Fatalf("e c opened %+v", m.modal)
	}
	typeIn(m, "Use JSON ")
	m.Update(tea.PasteMsg{Content: "Schema 2020-12."}) // a paste arrives as its own message
	run(m, "ctrl+s")
	if got := issue(t, p, ids["schema"]).Context; got != "Use JSON Schema 2020-12." {
		t.Fatalf("context = %q", got)
	}

	// p h sets the priority without the action menu.
	run(m, "p")
	run(m, "h")
	if got := issue(t, p, ids["schema"]).Priority; got != domain.PriorityHigh {
		t.Fatalf("i h: priority %q", got)
	}

	// Shifted letters and > do nothing; the footer keeps to the essentials.
	for _, k := range []string{"E", "K", ">"} {
		run(m, k)
		if m.modal != nil {
			t.Fatalf("%s opened %+v", k, m.modal)
		}
	}
	m.notice = ""
	if f := ansi.Strip(m.footer()); !strings.Contains(f, "? all keys") || strings.Contains(f, "tree") {
		t.Fatalf("footer: %q", f)
	}

	// o on an issue without links says so; tab switches views in the
	// detail as it does in the list.
	m.selectInCurrent(ids["survey"])
	run(m, "enter")
	run(m, "o")
	if m.modal != nil || !strings.Contains(m.notice, "no linked issues") {
		t.Fatalf("o without links: modal %v notice %q", m.modal, m.notice)
	}
	before := m.active
	run(m, "tab")
	if m.active == before {
		t.Fatal("tab in the detail did not switch the view")
	}
}

func TestContextRowsShowProgress(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 110, 28)
	run(m, "5") // To define: JSON writer is defined, a context row above its open child
	var line string
	for _, l := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if strings.Contains(l, "JSON writer") && line == "" {
			line = l
		}
	}
	if !strings.Contains(line, "0/1") {
		t.Fatalf("context row without progress: %q", line)
	}
	_ = ids
}

func TestRelationsByShape(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	run(m, "6")
	m.selectInCurrent(ids["export"])
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"Children ■■■■■ 0/2", "├─ ▶ " + shortID(ids["csv"]) + " CSV writer", "└─ ◐ " + shortID(ids["json"]) + " ■■■■■ 0/1 JSON writer"} {
		if !strings.Contains(v, want) {
			t.Fatalf("children tree lacks %q:\n%s", want, v)
		}
	}
	m.selectInCurrent(ids["json"])
	v = ansi.Strip(m.View().Content)
	for _, want := range []string{"↑ Export", "← needs ▶ " + shortID(ids["csv"]) + " CSV writer", "└─ ○ " + shortID(ids["schema"]) + " JSON schema"} {
		if !strings.Contains(v, want) {
			t.Fatalf("relations lack %q:\n%s", want, v)
		}
	}
	if strings.Contains(v, "child ") || strings.Contains(v, "depends on") {
		t.Fatalf("label column still drawn:\n%s", v)
	}
}

func TestMenusMoveWithJK(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["export"])
	run(m, "o") // link mode over the children
	run(m, "j")
	run(m, "j")
	run(m, "k")
	if !m.linkMode || m.linkSel != 1 {
		t.Fatalf("j j k in link mode: mode %v selection %d, targets %v", m.linkMode, m.linkSel, m.relations(m.selected()))
	}
	want := m.relations(ids["export"])[linkTargets(m.relations(ids["export"]))[1]].id
	run(m, "enter")
	if m.selected() != want {
		t.Fatalf("enter after j j k went to %s, want the second link %s", m.selected(), want)
	}
	run(m, "esc")
	run(m, "a")
	run(m, "j")
	if m.modal == nil || m.modal.cursor != 1 {
		t.Fatal("j does not move in the action menu")
	}
}

func TestCreateWizard(t *testing.T) {
	p, _ := sample(t)
	edits := []string{"Rows go to stdout.\n\nOne per line."}
	m := editable(t, p, &edits)
	run(m, "6")
	run(m, "n")
	run(m, "enter")
	if m.modal.step != 0 || m.modal.err == "" {
		t.Fatalf("enter without a title moved on: %+v", m.modal.step)
	}
	typeIn(m, "Stream rows")
	run(m, "enter")
	if m.modal.step != 1 || !strings.Contains(ansi.Strip(m.View().Content), "New issue · 2/3") {
		t.Fatalf("after the title: step %d", m.modal.step)
	}
	run(m, "esc") // back to the title, not out
	if m.modal == nil || m.modal.step != 0 {
		t.Fatal("esc did not step back")
	}
	run(m, "enter")
	run(m, "r") // research, and on to the requirement
	if m.modal.step != 2 || kinds[m.modal.kindIdx] != domain.KindResearch {
		t.Fatalf("r: step %d kind %s", m.modal.step, kinds[m.modal.kindIdx])
	}
	typeIn(m, "Rows go")
	run(m, "enter") // a new line, not a submit
	if m.modal == nil {
		t.Fatal("enter in the requirement submitted the wizard")
	}
	run(m, "ctrl+e") // replaced by the editor's text
	run(m, "ctrl+s")
	if m.modal != nil {
		t.Fatalf("create failed: %s", m.modal.err)
	}
	i := issue(t, p, m.selected())
	if i.Title != "Stream rows" || i.Kind != domain.KindResearch || i.Prose != "Rows go to stdout.\n\nOne per line." {
		t.Fatalf("created %q (%s) with requirement %q", i.Title, i.Kind, i.Prose)
	}
}

func TestEditSettings(t *testing.T) {
	p, _ := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "s")
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "Saved views") || !strings.Contains(v, "6  All") {
		t.Fatalf("settings screen:\n%s", v)
	}
	cfg := func() domain.Config {
		tr, err := p.load()
		if err != nil {
			t.Fatal(err)
		}
		return tr.Project.Config
	}

	// t previews the next theme on the issue screen without saving; enter
	// keeps it; esc after another step restores it.
	run(m, "t")
	if m.preview == nil || m.screen != screenIssues || m.th.Palette.Name != "high-contrast" || cfg().Theme != "" {
		t.Fatalf("t: preview %v, screen %d, shown %q, saved %q", m.preview != nil, m.screen, m.th.Palette.Name, cfg().Theme)
	}
	if f := ansi.Strip(m.footer()); !strings.Contains(f, "Theme high-contrast 2/8") {
		t.Fatalf("theme bar: %q", f)
	}
	run(m, "enter")
	if m.preview != nil || m.screen != screenSettings || cfg().Theme != "high-contrast" || m.th.Palette.Name != "high-contrast" {
		t.Fatalf("enter: preview %v, screen %d, saved %q", m.preview != nil, m.screen, cfg().Theme)
	}
	run(m, "enter") // the Theme row: preview again
	run(m, "right")
	run(m, "right")
	if m.th.Palette.Name != "pastel" {
		t.Fatalf("right right shows %s", m.th.Palette.Name)
	}
	run(m, "esc")
	if m.preview != nil || m.screen != screenSettings || m.th.Palette.Name != "high-contrast" || cfg().Theme != "high-contrast" {
		t.Fatalf("esc: shown %q, saved %q", m.th.Palette.Name, cfg().Theme)
	}
	run(m, "down")

	// Add a view; an invalid query is rejected and the dialog stays open.
	run(m, "n")
	typeIn(m, "Hot")
	run(m, "tab")
	typeIn(m, "--priority urgent")
	run(m, "enter")
	if m.modal == nil || !strings.Contains(m.modal.err, "--priority must be one of") {
		t.Fatalf("invalid query: modal %+v", m.modal)
	}
	for range "urgent" {
		m.modal.inputs[1], _ = m.modal.inputs[1].Update(keyMsg("backspace"))
	}
	typeIn(m, "high")
	run(m, "enter")
	if c := cfg(); c.Views["Hot"] != "--priority high" || c.ViewOrder[len(c.ViewOrder)-1] != "Hot" {
		t.Fatalf("added view: %+v", c)
	}
	if m.tabs[len(m.tabs)-1].name != "Hot" {
		t.Fatal("tabs did not follow the new view")
	}

	// Digits select the view row; m moves it, enter renames, d d deletes.
	run(m, "6") // All
	run(m, "m")
	run(m, "down")
	run(m, "enter")
	if c := cfg(); c.ViewOrder[5] != "Hot" || c.ViewOrder[6] != "All" || m.moving {
		t.Fatalf("after m down enter: %v (moving %v)", c.ViewOrder, m.moving)
	}
	run(m, "6") // Hot now
	run(m, "enter")
	typeIn(m, "ter")
	run(m, "enter")
	if c := cfg(); c.Views["Hotter"] != "--priority high" || c.ViewOrder[5] != "Hotter" {
		t.Fatalf("after rename: %+v", c)
	}
	run(m, "d")
	if !strings.Contains(m.notice, `delete view "Hotter"? press d again`) {
		t.Fatalf("first d: %q", m.notice)
	}
	run(m, "d")
	if c := cfg(); len(c.ViewOrder) != 6 || c.Views["Hotter"] != "" {
		t.Fatalf("after d d: %+v", c)
	}
}

func TestSetPriority(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	run(m, "p")
	if m.modal == nil || m.modal.heading != "Priority" {
		t.Fatal("p does not open the priority menu")
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "Priority · ") || !strings.Contains(v, "Medium (current)") {
		t.Fatalf("priority menu:\n%s", v)
	}
	run(m, "c")
	if got := issue(t, p, ids["csv"]).Priority; got != domain.PriorityCritical {
		t.Fatalf("priority = %q", got)
	}
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "!crit") || !strings.Contains(v, "priority critical") {
		t.Fatalf("row mark or detail missing:\n%s", v)
	}
	// Medium unsets the priority again.
	run(m, "p")
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
	run(m, "enter")
	run(m, "m") // kind: manual
	run(m, "ctrl+s")
	if m.modal != nil {
		t.Fatalf("create failed: %s", m.modal.err)
	}
	id := m.selected()
	i := issue(t, p, id)
	if i == nil || i.Title != "XML writer" || i.Kind != domain.KindManual || i.Parent != ids["export"] {
		t.Fatalf("created issue = %+v (selected %s)", i, id)
	}

	viaEditor(m, "e", "r")
	if got := issue(t, p, id).Prose; got != "Export rows as XML." {
		t.Fatalf("requirement after editor = %q", got)
	}
	edits = append(edits, issue(t, p, id).Body)
	viaEditor(m, "e", "r")
	if m.notice != "no changes" {
		t.Fatalf("unchanged editor text: notice %q", m.notice)
	}

	run(m, "e")
	run(m, "t")
	for range "XML writer" {
		m.modal.inputs[0], _ = m.modal.inputs[0].Update(keyMsg("backspace"))
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
	viaEditor(m, "e", "c")
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
	viaEditor(m, "e", "r")
	if m.modal == nil || !strings.Contains(m.modal.err, "disk full") || m.pending[ids["survey"]+"|requirement"] != "New text" {
		t.Fatalf("rejected edit: modal %+v pending %q", m.modal, m.pending)
	}
	run(m, "esc")
	var opened string
	m.runEditor = func(path string, done func(error) tea.Msg) tea.Cmd {
		return func() tea.Msg {
			b, _ := os.ReadFile(path)
			opened = string(b)
			return done(nil)
		}
	}
	viaEditor(m, "e", "r")
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
	run(m, "v")
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
	run(m, "enter")
	run(m, "m")
	run(m, "ctrl+s")
	id := m.selected()
	edits = append(edits, "Post a note in the changelog.")
	viaEditor(m, "e", "r")
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
	run(m, "v")
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
	if v := ansi.Strip(c.View().Content); !strings.Contains(v, "[x] Claude Code  detected") || !strings.Contains(v, "[ ] Pi") {
		t.Fatalf("view:\n%s", v)
	}
	for _, k := range []tea.KeyPressMsg{keyMsg("down"), keyMsg("space"), keyMsg("up"), keyMsg("space")} {
		c.Update(k)
	}
	_, cmd := c.Update(keyMsg("enter"))
	if cmd == nil || c.cancelled || c.items[0].Checked || !c.items[1].Checked {
		t.Fatalf("after toggling: %+v cancelled=%v", c.items, c.cancelled)
	}
	c2 := NewChecklist(testTheme(), "x", []ChecklistItem{{Label: "a"}})
	c2.Update(keyMsg("esc"))
	if !c2.cancelled {
		t.Fatal("esc does not cancel")
	}
}

func TestDetailDoesNotScrollSideways(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6")
	m.selectInCurrent(ids["export"])
	keys(m, "enter")
	m.View()
	// A line Glamour cannot wrap, such as a long code line, is wider than the pane.
	m.vp.SetContent(strings.Repeat("x", 300))
	m.detailKey("right", keyMsg("right"))
	m.Update(tea.MouseWheelMsg{X: 100, Y: 10, Button: tea.MouseWheelRight})
	if m.vp.XOffset() != 0 {
		t.Fatalf("detail scrolled sideways: x offset %d", m.vp.XOffset())
	}
}

func TestGallerySwitchesThemes(t *testing.T) {
	cfg := domain.Config{Theme: "mine", Themes: map[string]domain.ThemeDef{"mine": {Base: "nord"}}}
	g := newGallery(cfg, testTheme())
	g.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if g.th.Palette.Name != "mine" || !strings.Contains(ansi.Strip(g.render()), "theme mine") {
		t.Fatalf("gallery starts with %s", g.th.Palette.Name)
	}
	g.Update(keyMsg("t"))
	if g.th.Palette.Name != "default" {
		t.Fatalf("t after the last theme: %s", g.th.Palette.Name)
	}
	g.Update(keyMsg("t"))
	if g.th.Palette.Name != "high-contrast" || !strings.Contains(ansi.Strip(g.render()), "theme high-contrast") {
		t.Fatalf("t: %s", g.th.Palette.Name)
	}
	g.Update(keyMsg("T"))
	g.Update(keyMsg("T"))
	if g.th.Palette.Name != "mine" {
		t.Fatalf("T T: %s", g.th.Palette.Name)
	}
}

func TestUnsetThemeFollowsTheTerminal(t *testing.T) {
	p, _ := sample(t)
	m := openModel(t, p, 110, 28)
	m.retheme(false, m.th.Profile)
	if m.th.Palette.Name != "light" || m.th.Dark {
		t.Fatalf("light terminal shows %s", m.th.Palette.Name)
	}
	m.retheme(true, m.th.Profile)
	if m.th.Palette.Name != "default" || !m.th.Dark {
		t.Fatalf("dark terminal shows %s", m.th.Palette.Name)
	}

	// A configured theme ignores the background.
	m.tree.Project.Config.Theme = "nord"
	m.retheme(false, m.th.Profile)
	if m.th.Palette.Name != "nord" {
		t.Fatalf("configured nord shows %s", m.th.Palette.Name)
	}
}

func TestFilterBarSuggestsValues(t *testing.T) {
	p, _ := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6", "f")
	typeText(m, "--pri")
	if len(m.comp.items) != 1 || m.comp.items[0].value != "--priority" {
		t.Fatalf("--pri suggests %+v", m.comp.items)
	}
	keys(m, "tab")
	if m.input.Value() != "--priority " {
		t.Fatalf("tab inserted %q", m.input.Value())
	}
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"critical", "high", "medium", "low"} {
		if !strings.Contains(v, want) {
			t.Fatalf("priorities missing from the picker:\n%s", v)
		}
	}
	keys(m, "down", "tab") // high
	if m.input.Value() != "--priority high" {
		t.Fatalf("down tab inserted %q", m.input.Value())
	}

	// A click on a candidate inserts it.
	typeText(m, " --state ")
	clickText(t, m, "in-progress", 0, 120)
	if m.input.Value() != "--priority high --state in-progress" || !m.filtering {
		t.Fatalf("click inserted %q (filtering %v)", m.input.Value(), m.filtering)
	}

	// A typed ! keeps negating the value it completes.
	if f, err := parseFilterText("--state !open"); err != nil || len(f.NotStates) != 1 || len(f.Text) != 0 {
		t.Fatalf("--state !open: %+v, %v", f, err)
	}

	// --priority takes its value (it used to become a text search).
	keys(m, "enter")
	f, err := parseFilterText(m.filters[m.current().name])
	if err != nil || len(f.Priorities) != 1 || f.Priorities[0] != domain.PriorityHigh || len(f.Text) != 0 {
		t.Fatalf("applied filter %+v, %v", f, err)
	}
}

func TestFilterBarEnterPicksThenApplies(t *testing.T) {
	p, _ := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6", "f")
	typeText(m, "--sta")
	keys(m, "enter")
	if m.input.Value() != "--state " || !m.filtering || m.filterErr != "" {
		t.Fatalf("enter on --sta: %q filtering %v err %q", m.input.Value(), m.filtering, m.filterErr)
	}
	typeText(m, "d")
	keys(m, "down", "enter")
	if m.input.Value() != "--state done" || len(m.comp.items) != 0 {
		t.Fatalf("down enter: %q, suggestions %v", m.input.Value(), m.comp.items)
	}
	keys(m, "enter")
	if m.filtering || m.filters[m.current().name] != "--state done" {
		t.Fatalf("enter without suggestions did not apply: filtering %v, filter %q", m.filtering, m.filters[m.current().name])
	}

	// An error clears as soon as the input changes.
	keys(m, "f")
	m.input.SetValue("--nope x")
	m.refreshCompletion()
	keys(m, "enter")
	if m.filterErr == "" {
		t.Fatal("no error for --nope")
	}
	typeText(m, "y")
	if m.filterErr != "" {
		t.Fatalf("error kept after typing: %q", m.filterErr)
	}
}

func TestFilterBarFiltersLive(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 120, 30)
	keys(m, "6")
	all := len(rowIDs(m))
	keys(m, "f")
	typeText(m, "-")
	if len(rowIDs(m)) != all {
		t.Fatalf("a half-typed flag filtered: %v", rowIDs(m))
	}
	typeText(m, "-kind ")
	if len(rowIDs(m)) != all || m.filterErr != "" {
		t.Fatalf("--kind without a value changed rows or errored: %v %q", rowIDs(m), m.filterErr)
	}
	typeText(m, "decision")
	if got := rowIDs(m); len(got) != 1 || got[0] != ids["format"] || !m.filtering {
		t.Fatalf("--kind decision before enter: %v", got)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "/ --kind decision") {
		t.Fatalf("live query missing from the title:\n%s", v)
	}
	// esc goes back to the filter the tab had before (none).
	keys(m, "esc")
	if len(rowIDs(m)) != all || m.filters[m.current().name] != "" {
		t.Fatalf("esc kept the live filter: %v", rowIDs(m))
	}
	// enter keeps it.
	keys(m, "f")
	typeText(m, "--kind decision")
	keys(m, "enter")
	if m.filtering || m.filters[m.current().name] != "--kind decision" {
		t.Fatalf("enter: filtering %v, filter %q", m.filtering, m.filters[m.current().name])
	}
}

func TestLinkModeSelectsInPlace(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 110, 28)
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	run(m, "o")
	v := ansi.Strip(m.View().Content)
	if !m.linkMode || m.focus != focusDetail || !strings.Contains(v, "1 ↑ Export") || !strings.Contains(v, "2 → unblocks") {
		t.Fatalf("link mode does not show keys in place:\n%s", v)
	}
	golden(t, "links-110x28", m.View().Content)
	if f := ansi.Strip(m.footer()); !strings.Contains(f, "enter go") {
		t.Fatalf("footer in link mode: %q", f)
	}
	// The breadcrumb is the parent: the first link, enter goes there.
	run(m, "enter")
	if m.selected() != ids["export"] || m.linkMode {
		t.Fatalf("enter on the breadcrumb went to %s (link mode %v)", m.selected(), m.linkMode)
	}
	run(m, "backspace")
	// A key without a link leaves link mode and acts as usual: s opens
	// the settings.
	run(m, "o")
	run(m, "s")
	if m.linkMode || m.screen != screenSettings {
		t.Fatalf("s in link mode with two links: mode %v, screen %d", m.linkMode, m.screen)
	}
	run(m, "esc")
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	// esc leaves link mode and the detail looks as before.
	run(m, "o")
	run(m, "esc")
	if m.linkMode || strings.Contains(ansi.Strip(m.View().Content), "1 ↑ Export") {
		t.Fatal("esc did not leave link mode")
	}
	// A click on the breadcrumb goes to the parent too.
	clickText(t, m, "↑ Export", 0, 110)
	if m.selected() != ids["export"] {
		t.Fatalf("click on the breadcrumb went to %s", m.selected())
	}
}

// However many children a parent has, the detail shows one heading with
// the progress and the first unresolved ones; the heading opens them all in
// the list, focused on the parent, and backspace comes back.
func TestManyChildren(t *testing.T) {
	p, ids := sample(t)
	for n := range 30 {
		p.issue(fmt.Sprintf("Part %d", n+1), domain.KindCode, "", ids["export"])
	}
	m := openModel(t, p, 110, 40)
	run(m, "6")
	m.selectInCurrent(ids["export"])
	kids := len(m.tree.Children(ids["export"]))
	v := ansi.Strip(m.View().Content)
	pr := m.tree.ChildProgress(ids["export"])
	head := fmt.Sprintf("%d/%d · %d open  list all ›", pr.Done+pr.Dropped, kids, kids-pr.Done-pr.Dropped)
	if !strings.Contains(v, head) || strings.Contains(v, "more · o selects") {
		t.Fatalf("children heading missing or the block overflows:\n%s", v)
	}
	shown := 0
	for _, r := range m.relations(ids["export"]) {
		if r.label == "child" {
			shown++
			if m.tree.State(r.id).Terminal() {
				t.Fatalf("a resolved child is listed: %s", r.id)
			}
		}
	}
	if shown != maxChildren {
		t.Fatalf("%d children listed, want %d", shown, maxChildren)
	}
	// The heading is the first link; following it opens the children view:
	// unfiltered, every child, open ones first, resolved last.
	p.op(ids["survey"], domain.OpDrop, domain.Input{Reason: "not needed"})
	m.Update(loadedMsg(func() loadedMsg { tr, err := p.load(); return loadedMsg{tr, err} }()))
	run(m, "1") // a filtered view: Attention
	m.jump(ids["export"], false)
	m.focus = focusDetail
	from := m.active
	run(m, "o")
	run(m, "enter")
	tb := m.current()
	if !tb.transient || m.focus != focusList || m.scopeTop() != ids["export"] || len(tb.rows) < kids+1 {
		t.Fatalf("heading: tab %q transient %v, focus %d, scope %q, %d rows for %d children", tb.name, tb.transient, m.focus, m.scopeTop(), len(tb.rows), kids)
	}
	if m.tree.Issues[ids["survey"]].Parent == ids["export"] && tb.index(ids["survey"]) < 0 {
		t.Fatal("the dropped child is missing: the children view must not filter")
	}
	last := -1
	for _, r := range tb.rows[1:] {
		if m.tree.Issues[r.id].Parent != ids["export"] {
			continue
		}
		rank := stateRank(m.tree.State(r.id))
		if rank < last {
			t.Fatalf("children not ordered by state: %v", rowIDs(m))
		}
		last = rank
	}
	// ← at its top closes it and returns to the view it came from.
	run(m, "left")
	if m.current().transient || m.active != from || m.childView != "" {
		t.Fatalf("left: tab %q, active %d (from %d)", m.current().name, m.active, from)
	}
	// A click on the heading opens it too; switching views closes it.
	m.jump(ids["export"], false)
	m.focus = focusDetail
	clickText(t, m, "list all", 50, 110)
	if !m.current().transient {
		t.Fatal("a click on the heading did not open the children view")
	}
	run(m, "1")
	if m.childView != "" || len(m.tabs) != len(domain.ViewNames(m.tree.Project.Config)) {
		t.Fatal("switching views did not close the children view")
	}
}

func TestThemePreviewOnTheIssueScreen(t *testing.T) {
	p, _ := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 28})
	run(m, "6")
	run(m, "s")
	run(m, "enter") // the Theme row
	run(m, "left")  // the last theme: light
	if m.th.Palette.Name != "light" || m.screen != screenIssues {
		t.Fatalf("left from default shows %s on screen %d", m.th.Palette.Name, m.screen)
	}
	checkSize(t, m.View().Content, 110, 28)
	golden(t, "theme-preview-110x28", m.View().Content)

	// A reload during the preview keeps the previewed theme.
	tr, err := p.load()
	if err != nil {
		t.Fatal(err)
	}
	m.apply(tr, nil)
	if m.th.Palette.Name != "light" {
		t.Fatalf("reload reset the preview to %s", m.th.Palette.Name)
	}
	run(m, "esc")
	if m.th.Palette.Name != "default" || m.screen != screenSettings {
		t.Fatalf("esc: %s on screen %d", m.th.Palette.Name, m.screen)
	}
}

func TestEditTags(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	// A resolved issue may change its tags.
	m.selectInCurrent(ids["format"])
	run(m, "e")
	run(m, "g")
	if m.modal == nil || m.modal.kind != modalTags {
		t.Fatal("e g did not open the tags dialog")
	}
	typeIn(m, "export cli")
	run(m, "enter")
	if got := issue(t, p, ids["format"]).Tags; strings.Join(got, " ") != "export cli" {
		t.Fatalf("saved tags %v", got)
	}

	// Another issue: typing suggests tags in use; tab inserts, enter saves.
	m.selectInCurrent(ids["csv"])
	run(m, "e")
	run(m, "g")
	typeIn(m, "ex")
	if len(m.modal.picks) != 1 || m.modal.picks[0] != "export" {
		t.Fatalf("suggestions for ex: %v", m.modal.picks)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "Tags · ") || !strings.Contains(v, "export") {
		t.Fatalf("tags dialog:\n%s", v)
	}
	run(m, "enter") // picks export
	if m.modal == nil || m.modal.inputs[0].Value() != "export " {
		t.Fatalf("enter with a suggestion: %q", m.modal.inputs[0].Value())
	}
	run(m, "enter") // nothing left to pick: saves
	if got := issue(t, p, ids["csv"]).Tags; strings.Join(got, " ") != "export" {
		t.Fatalf("csv tags %v", got)
	}

	// Invalid tags keep the dialog with the error; empty removes all.
	run(m, "e")
	run(m, "g")
	m.modal.inputs[0].SetValue("not-ready")
	run(m, "enter")
	if m.modal == nil || m.modal.err == "" {
		t.Fatal("not-ready accepted")
	}
	m.modal.inputs[0].SetValue("")
	run(m, "enter")
	if got := issue(t, p, ids["csv"]).Tags; len(got) != 0 {
		t.Fatalf("empty field kept tags %v", got)
	}
}
