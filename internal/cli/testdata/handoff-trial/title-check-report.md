## Summary
Checked the change in `/Users/andre/coding/tooling/prep-wt-t-title` without modifying implementation files. The title and stack handling are in `internal/tui/app.go`; tests are in `internal/tui/app_test.go`.

## Criteria
1. met — `internal/tui/app.go` sets `v.WindowTitle` from `m.projectName()`; `TestProjectWindowTitle` checks the named project title.
2. met — `projectName()` falls back to `prep` in `internal/tui/projects.go`; `TestProjectWindowTitle` checks `prep: prep`.
3. met — `Run` pushes before running Bubble Tea and defers the pop after a successful push, covering returned errors and project-switch exits (`internal/tui/app.go`). Project switching returns from this `Run` before the CLI restarts it. `TestTitleStackSequences` verifies the CSI sequences.
4. met — `pushTerminalTitle` only writes when `out` is an `*os.File` for which `term.IsTerminal` is true (`internal/tui/app.go`). Buffer writers therefore skip both escapes.
5. met — `go test ./...` passed; `just check` passed (`prep check`: 0 errors, 0 warnings; `prep fmt --check` succeeded).

## Problems
The tests verify the view title and sequence writer separately, but do not exercise `Run` with a non-terminal writer or directly verify deferred pop on an error/switch. Those paths are supported by the implementation structure; terminal integration was not independently exercised.
