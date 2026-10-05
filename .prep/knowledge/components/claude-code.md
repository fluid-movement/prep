---
type: component
title: Claude Code integration
description: Project-local SessionStart hook running prep prime and the /prep-status, /prep-next and /prep-guide slash commands; read-only, no FSM logic.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - .claude/settings.json
  - .claude/commands
---

# Claude Code integration

Applies when changing how Claude Code sessions in this repository start or which prep slash commands exist.

- `.claude/settings.json` registers a `SessionStart` hook: `command -v prep >/dev/null 2>&1 && prep prime || true`. Its stdout is the session's opening context; without the binary it is silent and exits 0. It never installs prep; that waits for the install and update flow.
- `.claude/commands/` holds `prep-status.md` (`prep prime` plus `prep list`, rendered grouped by state), `prep-next.md` (`prep next`) and `prep-guide.md` (`prep guide $ARGUMENTS`, then work on the current step). They run the CLI through inline `!` bash with `allowed-tools: Bash(prep:*), Bash(echo:*)` and fall back to a "not installed" message pointing at the skill.
- Only read commands run from the hook and commands; transitions stay with the agent following `prep guide` and the [CLI](/components/cli.md) contract. The skill in `.claude/skills/prep/SKILL.md` is separate and must match `internal/cli/skill.md`.
- The configuration is project-local; shipping it to other projects as a Claude Code plugin is deferred until binaries are released.
