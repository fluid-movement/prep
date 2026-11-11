## Summary
Checked the implementation in `/Users/andre/coding/tooling/prep-wt-t-fold` without changing it. The change is limited to `internal/tui/app.go`; the requested TUI behavior lacks dedicated tests and the full checks do not pass.

## Criteria
1. met — `internal/tui/app.go:821-855` handles space by toggling the selected parent and calls `selectInCurrent(id)` after rebuilding; descendants are filtered at lines 518-530.
2. met — `internal/tui/app.go:1561-1570` shows folded/open markers for parents and the folded count uses `len(t.Descendants(id))`; leaves get no marker. No dedicated rendering test was found.
3. met — `internal/tui/app.go:831-855` targets same-parent issues with children and folds all if any is open, otherwise unfolds all. No dedicated behavior test was found.
4. met — `internal/tui/app.go:334-339` prunes absent IDs during rebuild and folding state is retained by issue ID in the model; no dedicated reload test was found.
5. met — `internal/tui/app.go:821-829` limits actions to tree view and parent rows; `internal/tui/app.go:1338-1343` registers both keys in help. No dedicated key-help test was found.
6. not met — `go test ./...` failed in `internal/tui` due to snapshot mismatches (`TestOnePaneLayout`, `TestScreenSnapshots`, `TestLinkModeSelectsInPlace`, `TestKnowledgeRecords`). `just check` reported zero errors/warnings from `prep check` but exited unsuccessfully at `go test ./...`.

## Problems
- No TUI fold/unfold tests or golden updates are present in the change, so the specific behavior above is supported by code inspection rather than executable tests.
- `just check` failed because `internal/tui` snapshot goldens differ, including list/screen snapshots; run output specifically recommended updating intended snapshots.
- `go test ./...` also failed on snapshot mismatches; all other listed packages passed.
- The definition-of-done requirement to update knowledge entries could not be verified: the brief prohibited reading `.prep` except to write this report.
