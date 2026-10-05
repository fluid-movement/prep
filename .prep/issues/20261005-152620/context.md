The integration is configuration and prompts only; no Go code changes. It calls read commands of the CLI described in [CLI](/components/cli.md) (`prime`, `list`, `next`, `guide`) and never a write command, so it holds no FSM logic and cannot change state behind the user's back. Writes stay with the agent following `prep guide`, as the [prep skill](/overview.md) already instructs.

Files:

- `.claude/settings.json`: project settings with a `SessionStart` hook. Hook stdout is added to the session context. The command must exit 0 even without the binary so sessions never show a hook error.
- `.claude/commands/prep-status.md`, `.claude/commands/prep-next.md`, `.claude/commands/prep-guide.md`: project slash commands. Each runs the CLI through inline bash (`!` lines) with `allowed-tools: Bash(prep:*)` and asks Claude to present the result to the user. `prep-guide` takes the issue ID as `$ARGUMENTS`.
- `.claude/skills/prep/SKILL.md` stays unchanged; it must match `internal/cli/skill.md` (tested in `internal/cli/cli_test.go`).

Text output is used rather than `--json`: it is already compact and readable, and the commands only render it. `prep list` shows ID, state, kind, progress or blocked status and title per issue.
