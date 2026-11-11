Only the TUI project switcher hides worktrees; the registry and `--project` resolution stay untouched (`prep projects` still lists them).

Code:
- `internal/tui/projects.go`: `openProjects` calls `m.opts.Projects()` and stores the list in `m.modal.projects`; `projectCandidates` builds the visible picks and `projectRows` renders them. Filter there (or once in `openProjects`), so both use the same list. Keep the current project visible even if it is itself a worktree (so the cursor and "current" note still work).
- Detection belongs in `internal/gitx/git.go` (git access lives there), e.g. `IsLinkedWorktree(dir string) bool`: true when `git rev-parse --git-dir` differs from `--git-common-dir` (the main checkout has them equal; a linked worktree's git dir is `<common>/worktrees/<name>`). Use the existing `run` helper; return false on any error. A `.git` file instead of directory is a cheaper hint but also matches submodules, so prefer rev-parse.
- Add `Worktree bool` to `projects.Project` (`internal/projects/projects.go`, derived in `Describe` like Remote; tag `json:"-"` or `omitempty` so the registry file does not store it), or call gitx from the CLI wrapper. Simplest: set it in `Describe`, and `internal/cli/tui.go` (`Projects: projects.List`) can stay as is. Check `List` rewrites the registry only on ID/Name/Remote changes so the new field must not trigger writes.
- `prep projects` (internal/cli) and `resolveIn` must not filter.

Knowledge: [Keep --project and the registry](/decisions/01M529ANGSVEK0ESGFXHQEFZVR.md) (registry stays); [Prep's part in a handoff](/decisions/01M54VQCGG9212J72216RJ1J8V.md) (implementers work in git worktrees, which is why they crowd the list; prep adds no worktree abstraction, so use plain git).

Tests: `internal/gitx/git_test.go` with a temp repo plus `git worktree add`; `internal/projects/projects_test.go` for Describe; `internal/tui/projects_test.go` (see `TestHeaderShowsTheProjectAndPOpensTheSwitcher`) for the switcher not listing a worktree project. Checks: `go test ./...`, `just check`.

Size: small.
