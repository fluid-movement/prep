## Summary
Checked the implementation in `/Users/andre/coding/tooling/prep-wt-t-worktrees` against the brief. The implementation changes are in `internal/gitx`, `internal/projects`, and `internal/tui`.

## Criteria
1. met — `internal/tui/projects.go` filters worktrees in `openProjects` before both candidate and row generation; `TestHeaderShowsTheProjectAndPOpensTheSwitcher` verifies a non-current worktree is absent.
2. met — `TestIsLinkedWorktree` checks a main checkout and plain directory return false; linked-worktree detection compares git metadata, so ordinary repositories/clones remain visible. The filter only excludes `Worktree` entries.
3. met — `internal/tui/projects.go` retains entries whose root equals `m.opts.Root`, even if `Worktree` is true; the switcher test exercises a current worktree.
4. met — `internal/cli` listing/resolution code was not changed; `internal/projects/projects.go` adds only a JSON-ignored derived field. The TUI-only filtering is in `openProjects`.
5. met — `internal/gitx/git_test.go:55` tests main checkout, linked worktree, and plain directory detection. `go test ./...` passed; `just check` passed (`prep check`: 0 errors, 0 warnings; `prep fmt --check` succeeded).

## Problems
- I did not inspect `.prep` knowledge entries because the brief prohibited reading `.prep` except for writing this report; therefore the Definition of Done requirement to verify knowledge-entry updates could not be confirmed.
