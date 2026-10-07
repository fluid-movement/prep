---
type: component
title: TUI mouse
description: 'How the TUI takes clicks and the wheel: elements marked as Lip Gloss layers while rendering, hit-testing, click semantics, and the per-user mouse capture setting.'
generated:
  by: claude-code/2.1.291
  at: 2026-10-06T13:36:45Z
scope:
  - internal/tui/mouse.go
  - internal/tui/mouse_test.go
  - internal/tui/program_test.go
confirmed_commit: 480ad2abb0bcbcefd900b1bbf60f03b5c3096d0d
---

# TUI mouse

Applies when adding a clickable element to a TUI screen or changing what clicks do. The keys stay the primary way in; every click does what an element's key does.

- **Marking** (`mouse.go`): `render` clears `m.hits` and `m.panes`, then each screen records what it draws as a Lip Gloss layer with an ID at its cells: `m.mark(id, x, y, w, h)` relative to `m.at` (the origin of the pane being rendered; `paneInner` is where a pane's body starts), `m.markAt` in screen cells, `m.pane(name, w, h)` for whole panes. IDs: `tab:n`, `pane:list|detail|knowledge|entry|page`, `row:n` (index into the tab's rows), `rel:n` (index into `relations`), `know:n`, `set:n`, `sugg:n` (filter bar candidates), and in dialogs `dialog`, `menu:n`, `pick:n`, `crit:n`, `kind:n`, `field:n`. Positions come from the drawing code: `ui.TabSpans` for tabs, `ui.OverlayAt` for where a dialog lands, `relationBlock`'s per-line link indices, and `modalView`'s `entryMark`s (body line, column, size). A new clickable element needs a mark where it is drawn and a case in `click`.
- **Hit-testing**: `hitAt` asks `lipgloss.NewCompositor(layers...).Hit(x, y)` for the topmost layer; heights are `zPane` < `zRow` < `zDialog` < `zEntry`, so a dialog's entries win over the screen behind it.
- **Clicks** (`click`, `clickDialog`; left button presses only): a click selects (row, knowledge entry, settings row, menu entry, move pick, kind option) and a click on the selected one does what enter does (opens the detail or the entry, runs the entry, edits or toggles the setting, moves, continues the wizard). Tabs switch views from any screen; relation lines follow their link like `o n` and push the back stack; a click in a pane focuses it; criteria toggle; form fields take focus. With a dialog open, clicks outside it do nothing (typed text survives); the keymap closes on any click. While the filter bar is open, only clicks on its candidates count (they insert the candidate). There is no timed double-click.
- **Wheel** (`wheel`): scrolls the pane under the pointer, focused or not. Every list scrolls its view, never its selection (the selection is made by clicking or with keys); this holds for any list, including new ones. The issue and knowledge lists move their offset `wheelRows` (3) per notch and set `m.wheeled`, so `listOffset` stops following the cursor and the selection may scroll out of view; any key clears it and the view follows the selection again. The move picker does the same with `modal.scroll`/`scrolled`. Viewports scroll their text by their own delta.

- **Setting**: mouse capture is the user's choice, not the project's: `tui.mouse` in the [user configuration](/components/setup.md) (unset means on). `prep tui` passes it as `Options.NoMouse` and saves changes through `Options.SaveMouse`; `View` sets `MouseMode` from `m.mouse`, so a toggle applies on the next frame. The settings screen shows it as the last row, under "You"; space or enter (or a second click) toggles it. With capture off the terminal selects text with a plain drag; the TUI does not list mouse interactions in its keymap or footer.
- **Tests** (`mouse_test.go`): `find` locates text on the rendered screen (inside an open dialog only), `click` renders then sends a `tea.MouseClickMsg`, `wheel` a `tea.MouseWheelMsg`; tabs, rows and relations run at 110×28 and 80×24. `program_test.go` runs the real program on SGR escape sequences to cover parsing end to end.
