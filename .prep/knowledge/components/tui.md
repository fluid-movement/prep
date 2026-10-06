---
type: component
title: TUI
description: 'prep tui screens in internal/tui — the issue views (tabs, list, detail), loader injection, live reload with fsnotify, keys, and how screens are tested.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T00:00:00Z
scope:
  - internal/tui/*.go
  - internal/cli/tui.go
---

# TUI

Applies when changing or adding TUI screens. Looks come from the [TUI design system](/components/tui-design-system.md); screens own state and compose its components.

- **Entry**: `prep tui` (`internal/cli/tui.go`) needs a terminal and calls `tui.Run` with `Options{Load, Watch}`: the CLI's own loader and the `.prep` directory. `internal/tui` never imports the storage adapters ([Architecture](/components/architecture.md)).
- **Issue views** (`app.go`, `Model`): tabs are the saved views in config order (`domain.ViewNames`, from `Config.ViewOrder`), each with its count, queried with `domain.ParseFilter` and `Tree.Query` like `prep list`. The list pane shows `ui.ListRow` per issue (short ID = time part, state, kind, progress, stale or blocked); the detail pane renders `detailMarkdown` (`detail.go`) with `ui.Markdown` into a Bubbles viewport, cached per issue, width and load generation. Markdown links show as text plus path.
- **Layout**: `ui.Split(width, 0.55, 52, 40)`; when both panes do not fit (below 92 columns) only the focused pane shows and enter/esc switch.
- **Keys**: tab/shift+tab and 1–9 switch tabs; up/down, j/k, g/G, pgup/pgdn move; enter/right focus the detail, esc/left return; y copies the ID with OSC 52; r reloads; q quits. The footer shows the keys of the focused pane, a notice, or the last load error.
- **Live reload** (`watch.go`): fsnotify on every directory under `.prep`, new directories added as they appear, debounced 150 ms, delivered as a message that triggers a reload. `apply` keeps each tab's selection by issue ID; a load error keeps the last good tree and shows the error.
- **Tests** (`app_test.go`): fixtures are built through the domain and markdown store with a fixed clock (stable IDs); screen snapshots at 110×28 and 80×24 in `internal/tui/testdata`, tests for tabs, selection across reloads, detail content, load errors and the watcher. `go test ./internal/tui/... -update` rewrites goldens.
- Startup asks the terminal for its background color (Lip Gloss); a terminal that never answers delays startup.
