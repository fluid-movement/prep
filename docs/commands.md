# Cloud handoff commands

## `prep handoff start <id>... [--yes]`

Start one Claude Code cloud session per issue. Every issue must be a ready,
unclaimed code or research issue, not stale or a parent, with dependencies done.
HEAD must already be pushed to its upstream. All issues are checked before any
session starts. The command prints the session count and asks for confirmation;
`--yes` skips the question.

Run from an interactive terminal. prep runs `claude --cloud` on a pseudo-terminal,
records the session ID and link in each issue's work log, and prints one line per
issue. Without a terminal it prints the exact commands to run and exits non-zero.
The environment is the default chosen with Claude Code's `/remote-env`.

The standard prompt asks the session to follow `prep guide`, implement and run
the checks, log “Implemented, ready for check”, and not complete the issue. It
also asks it to store the full handback-format report in a versioned finding
and commit and push the code and records on its branch.

## `prep handoff fetch [--merge]`

Fetch and list unmerged `claude/*` branches from origin, with each issue touched
and whether its cloud handback finding is present on the branch.

With `--merge`, merge each branch separately using `git merge --no-edit`, then
run `prep fmt` and `prep check` in process after each merge. Stop at the first
merge, format or validation failure and report it; resolve that failure before
continuing. No octopus merge is used.

Neither command pushes the local branch. The checker pass, session watching,
model selection and usage limits remain outside these commands. Both commands
support `--json`.
