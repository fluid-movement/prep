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
confirmed_commit: 94ee4aba23fb378ed2ec1805b1ecd478acd1a66b
---

# TUI

Applies when changing or adding TUI screens. Looks come from the [TUI design system](/components/tui-design-system.md); screens own state and compose its components.

- **Entry**: `prep tui` (`internal/cli/tui.go`) needs a terminal and calls `tui.Run` with `Options{Load, Watch}`: the CLI's own loader and the `.prep` directory. `internal/tui` never imports the storage adapters ([Architecture](/components/architecture.md)).
- **Issue views** (`app.go`, `Model`): tabs are the saved views in config order (`domain.ViewNames`, from `Config.ViewOrder`), each with its count, queried with `domain.ParseFilter` and `Tree.Query` like `prep list`. The list pane shows `ui.ListRow` per issue (short ID = time part, state, kind, progress, stale or blocked); the detail pane renders `detailMarkdown` (`detail.go`) with `ui.Markdown` into a Bubbles viewport, cached per issue, width and load generation. Markdown links show as text plus path.
- **Hierarchy**: tabs hold display rows, not bare IDs. Tree mode (default, `t` toggles flat) lays rows out with `Tree.WithAncestors` and `Tree.TreeOrder` like `prep list --tree`; ancestors that only give context are dimmed and not counted in the tab. A per-tab scope stack focuses a parent (`→` on a parent, `←`/esc up): the view narrows to its subtree, the parent first, with a breadcrumb in the pane title. `p` jumps to the parent. The detail lists parent, children, dependencies and blocked issues as `ui.LinkRow`s above the document; tab/shift+tab select, enter opens, backspace returns through a back stack. Jumps select the issue in the current tab when it is a row there, else in the first view without flags, clearing its focus.
- **Tags** show as `#tag` after titles in list rows and under the meta line in the detail; the filter bar accepts `--tag`.
- **Priority** (20261006-090120): `ui.PriorityMark` before titles in list rows (medium unmarked), `priority **<level>**` in the detail meta line; tabs and tree mode order by priority through `Query` and `TreeOrder`; the filter bar accepts `--priority`; action `i` (Set priority) opens the menu dialog with its own `heading` and the levels as entries (c, h, m, l; the current one marked) and writes a priority-only `PlanEdit`, also on resolved issues.
- **Filter bar**: `/` opens a text input; `parseFilterText` reads `prep list` flags and joins bare words into one `--text` phrase. The filter's query and the tab's query run separately and are intersected, so a filter only narrows. Filters are per tab, shown in the pane title; esc clears.
- **Stale diff**: for a stale issue the detail starts with `staleDiff`: the kind change and a `ui.DiffLines` line diff of the newest baseline's requirement against the current body, rendered with `ui.Diff`.
- **Screens**: `c` shows the check screen (`ui.DiagnosticRow` per diagnostic, errors then warnings), `s` the read-only settings (schema, commit mode, views in order, project DoD); esc or the same key returns. `Options.Check` loads with knowledge drift (git, about 170 ms) and runs only when the check screen opens or reloads while open; the header shows the last counts. The CLI builds fresh stores per load because reloads and checks can run concurrently.
- **Writes** (`actions.go`): `Options.Write` takes a plan function and runs the CLI's pipeline (fresh stores, plan, `CheckWrite`, Apply, stage or commit via `record` in `internal/cli/cli.go`), serialized by a mutex; `Options.Actor` is `human:<OS user>`, so the domain's `G_ACTOR` gate keeps code completion with agents. The TUI plans with the same domain functions as the CLI: `PlanNew`, `PlanEdit`, `PlanRecord`, `Plan`. After a write it reloads and selects the changed or created issue. Without `Write` the TUI is read-only.
- **Action menu** (`a`): new issue, title, requirement, context, criteria, move, priority, define, ready, acknowledge, drop, complete. Unavailable entries stay listed with the reason from `Tree.Gates` evaluated for the human actor with placeholder reason and no-impact arguments; code issues show that agents complete them. Menu letters run actions, arrows move, q or esc close. Shortcuts: `n` new issue (child of the focused parent), `e` edit requirement.
- **Dialogs** use `ui.Modal`, `ui.MenuRow` and `ui.Field` with Bubbles text inputs (static cursor, no blink timer); a rejected write keeps the dialog open with the error. Move offers top level and every unresolved issue outside the issue's own subtree, filtered as you type.
- **Editor**: `tea.ExecProcess` runs `$VISUAL`, `$EDITOR` or `vi` on a temp file; unchanged text writes nothing; text from a rejected write is kept per issue and field and reopened by the next edit. `runEditor` and `after` (timers) are hooks tests replace.
- **Layout**: `ui.Split(width, 0.55, 52, 40)`; when both panes do not fit (below 92 columns) only the focused pane shows and enter/esc switch.
- **Keys** depend on the focused pane. Both: 1–9 switch tabs, a actions, n new issue, e edit requirement, t tree/flat, p parent, c check, s settings, y copies the ID with OSC 52, r reloads, q quits. List: / filters. List: tab/shift+tab switch tabs; up/down, j/k, g/G, pgup/pgdn move; enter opens the detail; right/l focuses a parent or opens the detail; left/h/esc leave a focused parent. Detail: tab/shift+tab move between relations, enter opens one, backspace goes back, esc/left return to the list, other keys scroll. The footer shows the keys of the focused pane, a notice, or the last load error.
- **Live reload** (`internal/watch`, shared with `prep watch`): fsnotify on every directory under `.prep`, new directories added as they appear, debounced 150 ms, delivered as a message that triggers a reload. `apply` keeps each tab's selection by issue ID; a load error keeps the last good tree and shows the error.
- **Tests** (`app_test.go`): fixtures are built through the domain and markdown store with a fixed clock (stable IDs); screen snapshots at 110×28 and 80×24 in `internal/tui/testdata`, tests for tabs, selection across reloads, detail content and load errors; the watcher is tested in `internal/watch`. `go test ./internal/tui/... -update` rewrites goldens.
- Startup asks the terminal for its background color (Lip Gloss); a terminal that never answers delays startup.
