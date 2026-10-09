package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// find returns the screen cell where text first appears inside the columns
// [x0, x1), scanning rows from the top; with a dialog open, only inside it.
func find(t *testing.T, m *Model, text string, x0, x1 int) (x, y int) {
	t.Helper()
	y0 := 0
	if m.modal != nil {
		m.View()
		for _, l := range m.hits {
			if l.GetID() == "dialog" {
				x0, x1, y0 = l.GetX(), l.GetX()+l.Width(), l.GetY()
			}
		}
	}
	for y, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if y < y0 {
			continue
		}
		cells := ansi.Cut(line, x0, x1)
		if before, _, ok := strings.Cut(cells, text); ok {
			return x0 + ansi.StringWidth(before), y
		}
	}
	t.Fatalf("%q not on screen in columns %d-%d:\n%s", text, x0, x1, ansi.Strip(m.View().Content))
	return 0, 0
}

// click renders the screen, as the runtime does before input arrives, and
// clicks a cell.
func click(m *Model, x, y int) {
	m.View()
	_, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	settle(m, cmd)
}

// clickText clicks the first cell of text inside the columns [x0, x1).
func clickText(t *testing.T, m *Model, text string, x0, x1 int) {
	t.Helper()
	x, y := find(t, m, text, x0, x1)
	click(m, x, y)
}

func wheel(m *Model, x, y int, b tea.MouseButton) {
	m.View()
	m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: b})
}

func TestClickTabsRowsAndRelations(t *testing.T) {
	for _, size := range [][2]int{{110, 28}, {80, 24}} {
		w, h := size[0], size[1]
		p, ids := sample(t)
		m := openModel(t, p, w, h)
		listW, detailW := ui.Split(w, listRatio, minListW, minDetailW)
		if detailW == 0 {
			listW = w
		}

		// The screens bar switches screens and backspace goes back; a tab
		// switches the view of the screen shown.
		m.after = func(time.Duration, func(time.Time) tea.Msg) tea.Cmd { return nil }
		clickText(t, m, "Agent", 0, w)
		if m.screen != screenAgent {
			t.Fatalf("%dx%d: Agent click left screen %d", w, h, m.screen)
		}
		keys(m, "backspace")
		if m.screen != screenIssues {
			t.Fatalf("%dx%d: backspace left screen %d", w, h, m.screen)
		}
		clickText(t, m, "⚙", 0, w)
		if m.screen != screenSettings {
			t.Fatalf("%dx%d: gear click left screen %d", w, h, m.screen)
		}
		clickText(t, m, "Issues", 0, w)
		clickText(t, m, "To define", 0, w)
		if m.screen != screenIssues || m.current().name != "To define" {
			t.Fatalf("%dx%d: tab click: screen %d tab %q", w, h, m.screen, m.current().name)
		}

		keys(m, "6")

		// A row selects; the selected row opens its detail.
		clickText(t, m, "Survey export tools", 0, listW)
		if m.selected() != ids["survey"] {
			t.Fatalf("%dx%d: row click selected %s", w, h, m.selected())
		}
		clickText(t, m, "Export", 0, listW)
		if m.selected() != ids["export"] || m.focus != focusList {
			t.Fatalf("%dx%d: row click selected %s, focus %d", w, h, m.selected(), m.focus)
		}
		clickText(t, m, "Export", 0, listW)
		if m.focus != focusDetail {
			t.Fatalf("%dx%d: second click did not open the detail", w, h)
		}

		// A relation line follows the link; backspace comes back.
		x0 := listW
		if detailW == 0 {
			x0 = 0 // only the detail shows
		}
		clickText(t, m, "CSV writer", x0, w)
		if m.selected() != ids["csv"] {
			t.Fatalf("%dx%d: relation click selected %s", w, h, m.selected())
		}
		keys(m, "backspace")
		if m.selected() != ids["export"] {
			t.Fatalf("%dx%d: backspace after a relation click: %s", w, h, m.selected())
		}

		if detailW == 0 {
			continue
		}
		// A click in a pane focuses it.
		click(m, 5, h-3) // an empty cell of the list pane
		if m.focus != focusList {
			t.Fatalf("%dx%d: click in the list did not focus it", w, h)
		}
		click(m, listW+detailW/2, h-3)
		if m.focus != focusDetail {
			t.Fatalf("%dx%d: click in the detail did not focus it", w, h)
		}
	}
}

// crowd adds n open issues and n knowledge entries, so lists can scroll.
func crowd(t *testing.T, p *project, n int) {
	t.Helper()
	for k := range n {
		p.issue(fmt.Sprintf("Extra %02d", k), domain.KindCode, "More work.", "")
		path := filepath.Join(p.dir, ".prep", "knowledge", "components", fmt.Sprintf("extra%02d.md", k))
		os.MkdirAll(filepath.Dir(path), 0o755)
		body := fmt.Sprintf("---\ntype: component\ntitle: Extra %02d\ndescription: More.\nstatus: stable\n---\n\n# Extra %02d\n\nMore.\n", k, k)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWheelScrollsListsNotSelections(t *testing.T) {
	p, ids := sample(t)
	crowd(t, p, 30)
	m := openModel(t, p, 110, 28)
	keys(m, "6")
	listW, _ := ui.Split(110, listRatio, minListW, minDetailW)

	// The list scrolls three rows a notch; the selection stays, also out of view.
	m.focus = focusDetail
	wheel(m, 5, 10, tea.MouseWheelDown)
	if m.offset["All"] != wheelRows || m.selected() != ids["export"] || m.focus != focusDetail {
		t.Fatalf("wheel over the list: offset %d, selected %s, focus %d", m.offset["All"], m.selected(), m.focus)
	}
	for range 3 {
		wheel(m, 5, 10, tea.MouseWheelDown)
	}
	off := m.offset["All"]
	m.View()
	if off != 4*wheelRows || m.offset["All"] != off || m.selected() != ids["export"] {
		t.Fatalf("scrolled list: offset %d after render %d, selected %s", off, m.offset["All"], m.selected())
	}

	// A click selects a visible row without the view jumping.
	target := m.current().rows[off+1].id
	clickText(t, m, m.tree.Issues[target].Title, 0, listW)
	if m.selected() != target || m.offset["All"] != off {
		t.Fatalf("click in a scrolled list: selected %s, offset %d", m.selected(), m.offset["All"])
	}

	// A key moves the selection and brings it back into view.
	for range 4 {
		wheel(m, 5, 10, tea.MouseWheelUp)
	}
	if m.offset["All"] != 0 {
		t.Fatalf("wheel up: offset %d", m.offset["All"])
	}
	keys(m, "j")
	m.View()
	c := m.cursor["All"]
	if c != off+2 || c < m.offset["All"] || c >= m.offset["All"]+m.listHeight() {
		t.Fatalf("key after scrolling: cursor %d offset %d", c, m.offset["All"])
	}

	// The detail scrolls its text while the list keeps the focus.
	m.focus = focusList
	m.View()
	m.vp.SetContent(strings.Repeat("line\n", 200))
	wheel(m, listW+10, 10, tea.MouseWheelDown)
	if m.vp.YOffset() == 0 || m.focus != focusList {
		t.Fatalf("wheel over the detail: offset %d, focus %d", m.vp.YOffset(), m.focus)
	}

	// The knowledge list scrolls the same way.
	keys(m, "b")
	path := m.know.path
	wheel(m, 5, 10, tea.MouseWheelDown)
	m.View()
	if m.know.offset != wheelRows || m.know.path != path {
		t.Fatalf("wheel over the knowledge list: offset %d, path %s → %s", m.know.offset, path, m.know.path)
	}
}

func TestWheelScrollsTheMovePicker(t *testing.T) {
	p, ids := sample(t)
	crowd(t, p, 30)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["csv"])
	run(m, "a")
	run(m, "m")
	m.View()
	cursor, start := m.modal.cursor, m.modal.scroll
	wheel(m, 60, 15, tea.MouseWheelDown)
	m.View()
	if m.modal.cursor != cursor || m.modal.scroll != start+wheelRows {
		t.Fatalf("wheel over the picker: cursor %d → %d, first pick %d → %d", cursor, m.modal.cursor, start, m.modal.scroll)
	}
	// Clicking a pick scrolled into view selects it.
	pick := m.modal.picks[m.modal.scroll+1]
	clickText(t, m, m.tree.Issues[pick].Title, 0, 120)
	if m.modal.picks[m.modal.cursor] != pick {
		t.Fatalf("click after scrolling selected %s, want %s", m.modal.picks[m.modal.cursor], pick)
	}
}

func TestClickKnowledgeAndSettingsRows(t *testing.T) {
	p, _ := knowledgeSample(t)
	var saved []bool
	m := openModel(t, p, 110, 28)
	m.opts.SaveMouse = func(on bool) error { saved = append(saved, on); return nil }
	listW, _ := ui.Split(110, listRatio, minListW, minDetailW)

	keys(m, "b")
	clickText(t, m, "Default export format", 0, listW)
	clickText(t, m, "CLI", 0, listW)
	if m.know.path != "/components/cli.md" || m.know.onEntry {
		t.Fatalf("knowledge row click: path %s, on entry %v", m.know.path, m.know.onEntry)
	}
	clickText(t, m, "CLI", 0, listW)
	if !m.know.onEntry {
		t.Fatal("second click on the knowledge row did not open the entry")
	}
	click(m, 5, 10) // back in the list pane
	if m.know.onEntry {
		t.Fatal("click in the knowledge list did not focus it")
	}

	keys(m, "esc", "s")
	clickText(t, m, "--actionable", 0, 110)
	if m.setIdx != 2 {
		t.Fatalf("settings row click: row %d", m.setIdx)
	}
	clickText(t, m, "Mouse", 0, 110)
	clickText(t, m, "Mouse", 0, 110)
	if m.mouse || len(saved) != 1 || saved[0] {
		t.Fatalf("mouse row toggle: on %v, saved %v", m.mouse, saved)
	}
	if m.View().MouseMode != tea.MouseModeNone {
		t.Fatal("mouse off still captures the mouse")
	}
	// Off means off: clicks no longer act; the keyboard turns it back on.
	clickText(t, m, "--actionable", 0, 110)
	if m.setIdx == 2 {
		t.Fatal("a click acted with the mouse off")
	}
	keys(m, "space")
	if !m.mouse || m.View().MouseMode != tea.MouseModeCellMotion || len(saved) != 2 || !saved[1] {
		t.Fatalf("space on the mouse row: on %v, saved %v", m.mouse, saved)
	}

	// A failed save keeps the choice for the session and says so.
	m.opts.SaveMouse = func(bool) error { return errors.New("read-only") }
	keys(m, "space")
	if m.mouse || !strings.Contains(m.notice, "not saved: read-only") {
		t.Fatalf("failed save: on %v, notice %q", m.mouse, m.notice)
	}
}

func TestMouseOffFromTheUserConfiguration(t *testing.T) {
	p, _ := sample(t)
	m := NewModel(testTheme(), Options{Load: p.load, NoMouse: true})
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 28})
	if m.View().MouseMode != tea.MouseModeNone {
		t.Fatal("NoMouse still captures the mouse")
	}
}

func TestClickDialogs(t *testing.T) {
	p, ids := sample(t)
	var edits []string
	m := editable(t, p, &edits)
	run(m, "6")
	m.selectInCurrent(ids["csv"])

	// Menu entries: a click selects, a second click runs; outside does nothing.
	run(m, "a")
	clickText(t, m, "Set priority", 0, 120)
	if m.modal == nil || m.modal.items[m.modal.cursor].key != "p" {
		t.Fatalf("menu click: %+v", m.modal)
	}
	click(m, 0, 0)
	if m.modal == nil {
		t.Fatal("a click outside the dialog closed it")
	}
	clickText(t, m, "Set priority", 0, 120)
	if m.modal == nil || m.modal.heading != "Priority" {
		t.Fatalf("second menu click did not run the entry: %+v", m.modal)
	}
	clickText(t, m, "High", 0, 120)
	clickText(t, m, "High", 0, 120)
	if got := issue(t, p, ids["csv"]).Priority; got != "high" {
		t.Fatalf("priority by clicks: %q", got)
	}

	// An unavailable entry says why instead of running (the menus built
	// today list only available entries; runAction guards them anyway).
	m.modal = &modal{kind: modalMenu, items: []action{
		{key: "n", label: "New issue", run: m.openCreate},
		{key: "d", label: "Define", reason: "open questions remain", run: func() tea.Cmd { panic("ran") }},
	}}
	clickText(t, m, "Define", 0, 120)
	clickText(t, m, "Define", 0, 120)
	if !strings.Contains(m.notice, "define: open questions remain") {
		t.Fatalf("unavailable entry: notice %q", m.notice)
	}
	run(m, "esc")

	// Criteria toggle on click.
	run(m, "a")
	run(m, "v")
	before := m.modal.checks[1]
	clickText(t, m, "quotes fields", 0, 120)
	if m.modal.checks[1] == before || m.modal.cursor != 1 {
		t.Fatalf("criterion click: checks %v cursor %d", m.modal.checks, m.modal.cursor)
	}
	run(m, "esc")

	// Move picks select, and a second click moves.
	run(m, "a")
	run(m, "m")
	clickText(t, m, "Survey export tools", 0, 120)
	clickText(t, m, "Survey export tools", 0, 120)
	if got := issue(t, p, ids["csv"]).Parent; got != ids["survey"] {
		t.Fatalf("move by clicks: parent %q", got)
	}

	// Form fields take focus.
	run(m, "a")
	run(m, "x")
	if m.modal == nil || m.modal.kind != modalDrop {
		t.Fatalf("a x opened %+v", m.modal)
	}
	run(m, "esc")
	run(m, "s")
	run(m, "n")
	clickText(t, m, "Query", 0, 120)
	if m.modal.focus != 1 || !m.modal.inputs[1].Focused() || m.modal.inputs[0].Focused() {
		t.Fatalf("field click: focus %d", m.modal.focus)
	}
	run(m, "esc")
	run(m, "esc")

	// The wizard's kind options: a click picks, a second click continues.
	run(m, "n")
	typeIn(m, "Clicked")
	run(m, "enter")
	clickText(t, m, "research", 0, 120)
	if kinds[m.modal.kindIdx] != "research" || m.modal.step != 1 {
		t.Fatalf("kind click: %s step %d", kinds[m.modal.kindIdx], m.modal.step)
	}
	clickText(t, m, "research", 0, 120)
	if m.modal.step != 2 {
		t.Fatalf("second kind click: step %d", m.modal.step)
	}
	run(m, "esc")
	run(m, "esc")
	run(m, "esc")

	// The keymap closes on any click.
	run(m, "?")
	click(m, 0, 0)
	if m.modal != nil {
		t.Fatal("a click did not close the keymap")
	}
}
