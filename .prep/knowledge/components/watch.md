---
type: component
title: Watcher
description: internal/watch — reports changes under .prep, debounced and merged, split into project and local changes; used by prep tui's live reload and prep watch.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T17:45:26Z
scope:
  - internal/watch
  - internal/cli/watch.go
---

# Watcher

Applies when changing how prep notices changes to `.prep`: [TUI](/components/tui.md) live reload and `prep watch` for harness integrations ([CLI](/components/cli.md)).

- `watch.Dir(dir)` returns a channel of `Change` and a stop function. fsnotify is not recursive, so every directory under `dir` is watched and directories created later are added as they appear.
- A `Change` says what changed: `Tree` (issues, knowledge, config) or `Local` (anything under `.prep/local`, the per-machine [activity stream](/features/agent-activity.md)). The TUI reloads the issue tree only for `Tree` and the activity only for `Local`, so a busy agent never reloads the tree; `prep watch` prints a line only for `Tree`.
- Events wait 150 ms (`debounce`) for writes to settle, since a transition writes several files. One goroutine owns the pending change, the timer and all sends: at most one report waits in the channel, and a new one merges with it, so a slow reader never blocks the watcher.
- A watcher error (an overflow, a failed watch) is reported as a full change (`Tree` and `Local`), so the reader reloads everything rather than missing changes.
- Tests (`watch_test.go`) cover nested and new directories and the split between project and local changes.
