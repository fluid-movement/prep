## Summary
Updated the TUI snapshot goldens to match the current list/fold implementation in this working directory.

## Decisions
### Regenerated affected snapshots
The failing tests explicitly identified snapshot mismatches and recommended `go test ./internal/tui/... -update`; regenerated the intended goldens rather than changing application behavior or tests.

## Findings
### Snapshot tests now match
`go test ./...` passes after regenerating the goldens. `just check` also passes, including `prep check` (zero errors/warnings) and `prep fmt --check`.

## Problems
No problems encountered.
