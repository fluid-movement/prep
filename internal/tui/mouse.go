package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Mouse support: render marks every clickable element as a Lip Gloss layer
// with an ID ("tab:2", "row:5", "menu:1", ...) at the cells it was drawn
// on; a click asks the compositor which layer is on top at the pointer and
// does what that element's key does. Dialog layers sit above the screen.

// Layer heights: panes, then their rows, then an open dialog and its entries.
const (
	zPane = iota
	zRow
	zDialog
	zEntry
)

// point is a cell position on the screen.
type point struct{ x, y int }

// mark records a clickable region relative to the pane being rendered.
func (m *Model) mark(id string, x, y, w, h int) {
	m.markAt(id, m.at.x+x, m.at.y+y, w, h, zRow)
}

// markAt records a clickable region in screen cells.
func (m *Model) markAt(id string, x, y, w, h, z int) {
	if w > 0 && h > 0 {
		m.hits = append(m.hits, lipgloss.NewLayer(blank(w, h)).ID(id).X(x).Y(y).Z(z))
	}
}

// pane records the pane being rendered, w×h cells at m.at; a click inside
// focuses it and the wheel scrolls it.
func (m *Model) pane(name string, w, h int) {
	l := lipgloss.NewLayer(blank(w, h)).ID("pane:" + name).X(m.at.x).Y(m.at.y).Z(zPane)
	m.hits = append(m.hits, l)
	m.panes = append(m.panes, l)
}

// paneInner is where a pane's body starts, relative to the pane.
var paneInner = point{1 + theme.Pad, 1}

func blank(w, h int) string {
	return strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", w)+"\n", h), "\n")
}

// hitAt names the topmost clickable element at a cell, split into its kind
// and argument: "row", "5"; "" when nothing is there.
func hitAt(layers []*lipgloss.Layer, x, y int) (kind, arg string) {
	kind, arg, _ = strings.Cut(lipgloss.NewCompositor(layers...).Hit(x, y).ID(), ":")
	return kind, arg
}

func (m *Model) mouseMsg(msg tea.MouseMsg) tea.Cmd {
	if !m.mouse || m.tree == nil {
		return nil
	}
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return m.click(msg.X, msg.Y)
		}
	case tea.MouseWheelMsg:
		return m.wheel(msg)
	}
	return nil
}

// click does what the key of the element under the pointer does: a click
// selects, a click on the selected element activates it like enter.
func (m *Model) click(x, y int) tea.Cmd {
	kind, arg := hitAt(m.hits, x, y)
	n, _ := strconv.Atoi(arg)
	if m.modal != nil {
		return m.clickDialog(kind, n)
	}
	if m.filtering {
		return nil // the filter bar has the keyboard until enter or esc
	}
	switch kind {
	case "tab":
		m.screen, m.moving = screenIssues, false
		m.switchTab(n)
	case "pane":
		switch arg {
		case "list":
			m.focus = focusList
		case "detail":
			if m.selected() != "" {
				m.focus = focusDetail
			}
		case "knowledge":
			m.know.onEntry = false
		case "entry":
			m.know.onEntry = m.know.path != ""
		}
	case "row":
		tb := m.current()
		m.focus = focusList
		if m.cursor[tb.name] == n {
			return m.listKey("enter")
		}
		m.cursor[tb.name] = n
	case "rel":
		id := m.selected()
		rels := m.relations(id)
		if n >= len(rels) {
			return nil
		}
		if r := rels[n]; r.label == "knowledge" {
			m.back = append(m.back, id)
			return m.openKnowledge(r.id)
		}
		return m.jump(rels[n].id, true)
	case "know":
		paths, _ := m.knowledgePaths()
		if n >= len(paths) {
			return nil
		}
		if m.know.cursor == n {
			m.know.onEntry = true
			return nil
		}
		m.know.cursor, m.know.path, m.know.onEntry = n, paths[n], false
	case "set":
		if m.setIdx == n {
			return m.settingsKey("enter")
		}
		m.setIdx = n
	}
	return nil
}

// clickDialog handles clicks while a dialog is open; clicks outside it do
// nothing, so typed text is never thrown away.
func (m *Model) clickDialog(kind string, n int) tea.Cmd {
	d := m.modal
	if d.kind == modalHelp {
		m.modal = nil // any click closes it, as any key does
		return nil
	}
	switch kind {
	case "menu":
		if d.cursor == n {
			return m.runAction(d.items[n])
		}
		d.cursor = n
	case "pick":
		if d.cursor == n {
			return m.modalKey(tea.KeyPressMsg{Code: tea.KeyEnter})
		}
		d.cursor = n
	case "crit":
		d.cursor = n
		d.checks[n] = !d.checks[n]
	case "kind":
		if d.kindIdx == n {
			return m.createStep(2)
		}
		d.kindIdx = n
	case "field":
		d.focus = n
		for k := range d.inputs {
			d.inputs[k].Blur()
		}
		return d.inputs[n].Focus()
	}
	return nil
}

// wheel scrolls the pane under the pointer, focused or not: the list moves
// its selection a row per notch, the other panes scroll their text.
func (m *Model) wheel(msg tea.MouseWheelMsg) tea.Cmd {
	if m.modal != nil {
		return nil
	}
	_, name := hitAt(m.panes, msg.X, msg.Y)
	step := map[tea.MouseButton]int{tea.MouseWheelDown: 1, tea.MouseWheelUp: -1}[msg.Button]
	var cmd tea.Cmd
	switch name {
	case "list":
		if tb := m.current(); tb != nil && step != 0 {
			m.cursor[tb.name] = clamp(m.cursor[tb.name]+step, 0, len(tb.rows)-1)
		}
	case "detail":
		m.vp, cmd = m.vp.Update(msg)
	case "knowledge":
		if paths, _ := m.knowledgePaths(); len(paths) > 0 && step != 0 {
			m.know.cursor = clamp(m.know.cursor+step, 0, len(paths)-1)
			m.know.path = paths[m.know.cursor]
		}
	case "entry":
		m.know.vp, cmd = m.know.vp.Update(msg)
	case "page":
		m.page, cmd = m.page.Update(msg)
	}
	return cmd
}

// toggleMouse turns mouse capture on or off from the settings screen and
// records the choice in the user configuration.
func (m *Model) toggleMouse() tea.Cmd {
	m.mouse = !m.mouse
	state := map[bool]string{true: "on", false: "off"}[m.mouse]
	if m.opts.SaveMouse != nil {
		if err := m.opts.SaveMouse(m.mouse); err != nil {
			return m.flashErr(fmt.Errorf("mouse %s for this session, not saved: %v", state, err))
		}
	}
	return m.flash("mouse " + state)
}
