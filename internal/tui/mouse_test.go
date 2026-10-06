package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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

		// A tab switches the view, also from another screen.
		keys(m, "s")
		clickText(t, m, "To define", 0, w) // the header comes first
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

func TestWheelScrollsThePaneUnderThePointer(t *testing.T) {
	p, _ := sample(t)
	m := openModel(t, p, 110, 28)
	keys(m, "6")
	listW, _ := ui.Split(110, listRatio, minListW, minDetailW)

	// The list moves its selection, even when the detail has the focus.
	m.focus = focusDetail
	before := m.cursor["All"]
	wheel(m, 5, 10, tea.MouseWheelDown)
	if m.cursor["All"] != before+1 || m.focus != focusDetail {
		t.Fatalf("wheel over the list: cursor %d → %d, focus %d", before, m.cursor["All"], m.focus)
	}
	wheel(m, 5, 10, tea.MouseWheelUp)
	if m.cursor["All"] != before {
		t.Fatalf("wheel up over the list: cursor %d", m.cursor["All"])
	}

	// The detail scrolls its text while the list keeps the focus.
	m.focus = focusList
	m.View()
	m.vp.SetContent(strings.Repeat("line\n", 200))
	wheel(m, listW+10, 10, tea.MouseWheelDown)
	if m.vp.YOffset() == 0 || m.focus != focusList {
		t.Fatalf("wheel over the detail: offset %d, focus %d", m.vp.YOffset(), m.focus)
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
	if m.modal == nil || m.modal.items[m.modal.cursor].key != "i" {
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
