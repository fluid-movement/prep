## D1: Hit-test with geometry recorded while rendering, not bubblezone
date: 2026-10-06

The model records where tabs, rows, relation lines and dialog lines are drawn while `View` renders, and `mouse.go` maps click positions to those records.

Alternatives considered:
- bubblezone (lrstanley/bubblezone) marks regions with zero-width escape sequences and scans the final output. `ui.Overlay` strips and truncates background lines with `ansi.Strip`/`ansi.Truncate`, so markers behind a dialog get cut or dropped. It also adds a dependency and a `zone.Scan` wrapper around every `View`, which the goldens would have to go through.
- Recomputing the geometry in `Update` from widths and heights would duplicate the layout math from `View`, and the two copies would drift apart as screens change.

Recording during rendering is exact by construction and keeps the TUI dependency-free for this feature. It is also testable because `View` runs in the existing tests.

## D2: Click semantics: select, click again to activate
date: 2026-10-06

- A left press acts. Releases and motion are ignored, so `WithMouseCellMotion` stays enough.
- A click on an unselected row (issue, knowledge entry, settings row, menu entry, reparent pick) selects it. A click on the selected row does what enter does: opens the detail, runs the entry, or scrolls to the knowledge entry. There is no timed double-click: Bubble Tea v1 reports no click count, a timer is fragile over SSH, and a second click already covers it.
- A click on an unavailable menu entry shows its reason, as its key does.
- A click on a tab switches to that view. From the check, settings or knowledge screen it also returns to the issue views.
- A click on a relation line follows it as `o <n>` does and pushes the back stack, so backspace returns.
- A click inside a pane focuses that pane. The wheel scrolls the pane under the pointer, not the focused one: the list moves its selection by one row per tick, and viewports scroll by their own wheel delta.
- In dialogs, clicks hit menu entries, reparent picks, criteria (a click toggles the checkbox), the wizard's kind options and form fields (a click focuses the field). A click outside an open dialog does nothing, so it cannot throw away typed text. The help dialog closes on any click, as on any key.
- Footer key hints are not clickable. They document keys rather than being controls.
- In the narrow layout, clicks go to the pane that is showing.

## D3: Mouse capture is a per-user setting in the user configuration
date: 2026-10-06

Whether the TUI captures the mouse depends on the person and their terminal, not on the project. So it lives in `~/.config/prep/config.yaml` (`tui.mouse`, default on), which the setup knowledge entry already names as the place for per-user TUI choices. It does not go in `.prep/project.md`, which is committed and shared. The settings screen shows it in its own group, labeled as a setting for this user, so it isn't mistaken for a project setting. The CLI loads and saves it through `Options` callbacks, so `internal/tui` does not import `internal/userconfig`. Toggling it applies at once (`tea.EnableMouseCellMotion` / `tea.DisableMouse`).

The selection modifier differs by terminal (shift in most of them, option in iTerm2 and Terminal.app). The keymap names both instead of detecting the terminal.

## D4: Hit-test with Lip Gloss v2 layers and Compositor.Hit
date: 2026-10-06
supersedes: D1

After the Charm v2 migration (20261006-130742), the screen is composed with Lip Gloss v2 layers. Clickable elements get a layer with an ID that names them: `tab:<n>`, `row:<issue id>`, `rel:<n>`, `pane:list`, `pane:detail`, `know:<path>`, `set:<n>`, `menu:<n>`, `pick:<n>`, `crit:<n>`, `kind:<n>`, `field:<n>`, and `dialog` for the dialog as a whole. Each layer sits at the position where it is drawn. The open dialog is a layer with a higher Z, so `Compositor.Hit(x, y)` returns its elements first, and a hit on `dialog` without a more specific element does nothing. The model keeps the compositor from the last `View` and resolves `tea.MouseClickMsg` and `tea.MouseWheelMsg` with `Hit`, then calls the same functions the keys call.

This is Charm's own mechanism. It replaces the plan to record geometry by hand (D1) and the string splicing in `ui.Overlay`. Alternatives considered: bubblezone, a third-party library whose markers break under the overlay, and hand-recorded geometry, which would duplicate what layers already provide.

## D5: Mouse capture is a per-user setting, applied through View.MouseMode
date: 2026-10-06
supersedes: D3

Whether the TUI captures the mouse depends on the person and their terminal, not on the project. So it lives in `~/.config/prep/config.yaml` (`tui.mouse`, default on), not in the committed `.prep/project.md`. The settings screen shows it in its own group, labeled as a setting for this user. The CLI loads and saves it through `Options` callbacks (`Mouse bool`, `SaveMouse func(bool) error`), so `internal/tui` does not import `internal/userconfig`. `View` sets `MouseMode` to `tea.MouseModeCellMotion` or `tea.MouseModeNone` from the model, so a toggle applies on the next render without enable or disable commands.

The selection modifier differs by terminal (shift in most of them, option in iTerm2 and Terminal.app). The keymap names both instead of detecting the terminal.
