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
var paneInner = point{1 + theme.SpaceS, 1}

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
		switch msg.Button {
		case tea.MouseLeft:
			return m.click(msg.X, msg.Y)
		case tea.MouseBackward:
			// The mouse's back button is backspace: back where the person
			// was. An open dialog or the filter bar keeps it, as with clicks.
			if m.modal == nil && !m.filtering {
				return m.goBack()
			}
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
		// The filter bar has the keyboard until enter or esc; only its
		// candidates take clicks.
		if kind == "sugg" {
			m.insertSuggestion(n)
		}
		return nil
	}
	switch kind {
	case "screen":
		return m.goScreen(navScreens[n])
	case "check":
		return m.goScreen(screenCheck)
	case "settings":
		return m.toggleScreen(screenSettings)
	case "tab":
		return m.setView(n)
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
			m.know.onEntry = m.shownEntry() != ""
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
		m.linkMode = false
		if n >= len(rels) {
			return nil
		}
		return m.followRel(id, rels[n])
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

// wheelRows is how far a wheel notch scrolls a list, as far as it scrolls
// a viewport's text.
const wheelRows = 3

// wheel scrolls the pane under the pointer, focused or not. Lists scroll
// their view and keep their selection (a click selects); the other panes
// scroll their text. Rendering clamps the offsets.
func (m *Model) wheel(msg tea.MouseWheelMsg) tea.Cmd {
	step := map[tea.MouseButton]int{tea.MouseWheelDown: wheelRows, tea.MouseWheelUp: -wheelRows}[msg.Button]
	if d := m.modal; d != nil {
		if d.kind == modalReparent && step != 0 {
			d.scroll, d.scrolled = max(0, d.scroll+step), true
		}
		return nil
	}
	_, name := hitAt(m.panes, msg.X, msg.Y)
	var cmd tea.Cmd
	switch name {
	case "list":
		if tb := m.current(); tb != nil && step != 0 {
			m.offset[tb.name], m.wheeled = max(0, m.offset[tb.name]+step), true
		}
	case "detail":
		m.vp, cmd = m.vp.Update(msg)
	case "knowledge":
		if step != 0 {
			m.know.offset, m.wheeled = max(0, m.know.offset+step), true
		}
	case "entry":
		m.know.vp, cmd = m.know.vp.Update(msg)
	case "feed", "issue", "usage":
		// The Agent views scroll their lines; rendering clamps the offset.
		if step != 0 {
			off := &m.agentS.offsets[m.agentS.view]
			*off = max(0, *off+step)
		}
	case "page":
		m.page, cmd = m.page.Update(msg)
	}
	return cmd
}

// listOffset is the first visible row of a list of n rows that shows rows
// at a time: the stored offset kept in range and, when follow is set,
// moved just enough to show the cursor. After a wheel scroll the view
// stays where the wheel put it.
func listOffset(off, cursor, n, rows int, follow bool) int {
	off = clamp(off, 0, max(0, n-rows))
	if follow {
		if cursor < off {
			off = cursor
		}
		if cursor >= off+rows {
			off = cursor - rows + 1
		}
	}
	return off
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
