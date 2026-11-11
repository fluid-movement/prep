## Summary
Implemented TUI-only hiding of registered linked worktrees in `/Users/andre/coding/tooling/prep-wt-t-worktrees`. Worktree checkouts remain registered and resolvable; the current checkout remains visible.

## Decisions
### Detect linked worktrees with git metadata
Compared `git rev-parse --git-dir` and `--git-common-dir`, resolving relative paths against the checkout. Errors return false, preserving visibility when git detection is unavailable.

### Derive worktree state without storing it
Added `Project.Worktree` as a JSON-ignored derived field. This keeps registry contents unchanged, and the existing registry rewrite comparison remains limited to identity fields.

### Filter only when opening the TUI switcher
Filtered the project list used by both switcher candidates and rows, retaining the current root even when it is a linked worktree. Registry listing and project resolution are untouched.

## Findings
### Git reports worktree metadata relative to the checkout
The `rev-parse` paths are resolved relative to the project root before comparison, allowing correct detection regardless of whether the paths are absolute or relative.

## Problems
The brief's Definition of Done requests updating knowledge entries, but its rule forbids reading or changing `.prep` except for this report. I left knowledge entries unchanged to follow that restriction.

## Checks
- `go test ./...` — passed.
- `just check` — passed (`prep check` and `prep fmt --check`).
