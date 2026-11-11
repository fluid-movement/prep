## Summary
Checked the uncommitted implementation in `/Users/andre/coding/tooling/prep-wt-t-fold`. The folding implementation is in `internal/tui/app.go`; no changes were made.

## Criteria
1. met — `internal/tui/app.go:518-530` removes rows with a folded ancestor; `listKey` rebuilds and reselects the issue ID after folding (`:821-856`).
2. met — `issueRow` adds `▸` and `len(t.Descendants(id))` when folded, `▾` when open, and no marker for leaves (`internal/tui/app.go:1559-1571`).
3. met — `listKey`'s `x` case selects parent-bearing issues sharing the selected issue's parent and folds all if any are open, otherwise unfolds all (`internal/tui/app.go:838-854`).
4. met — fold state is keyed by ID and retained across `rebuild`; IDs absent from the current tree are deleted (`internal/tui/app.go:335-340`).
5. met — `listBindings` lists `space` and `x` (`internal/tui/app.go:1341-1342`); the handler is inert for leaves and flat mode only sets a notice (`:823-831`).
6. met — `go test ./...` passed; `just check` passed (`prep check`: 0 errors, 0 warnings; `prep fmt --check` succeeded).

## Problems
No dedicated folding interaction/reload tests were found in `internal/tui/*test.go`; the folding behavior and reload preservation are evidenced by code inspection rather than tests. Mouse marker activation is not implemented; the brief allows this as a follow-up if the hit-region is not cheap.
