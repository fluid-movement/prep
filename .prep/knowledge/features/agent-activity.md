---
type: feature
title: Agent activity
description: 'The local activity stream under .prep/local: what prep and harnesses record, the event format, how an agent''s current issue is derived, prep focus and prep activity.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T13:28:19Z
scope:
  - internal/activity
  - internal/cli/activity.go
confirmed_commit: 9c61698b74938475088e43b2ab260f5d0777b24c
---

# Agent activity

Applies when changing what prep records about agents, the event format, or anything that reads the stream (the TUI's Agent view, `prime`, harness bridges). Harnesses report to the local stream; the person runs `prep tui` next to the agent.

- **Where**: `.prep/local/activity.jsonl`, per machine, ignored through `.prep/.gitignore` (`local/`; P007 warns when `.prep/local` exists unignored and `prep fix` adds the line). It is UI state, never a record: reads still never change records, and nothing in it is staged or committed. Each git worktree has its own.
- **Writing** (`internal/activity`): `Append` writes one JSON line per event with `O_APPEND` in a single write below 4 KiB (every string field clipped to 300 characters, the target dropped if the line is still too long, and an event that cannot fit refused), so concurrent agents never interleave lines; past 1 MiB the file is rewritten (temp file, rename) with its newest lines, at most 2,000 and half the limit. `Read`/`Parse` skip lines that do not parse.
- **Event**: `at`, `actor`, `kind`, optional `session`, `op` (the prep command or harness tool), `issue`, `verb`, `target`, `area` (issue, knowledge, code, prep, other), `chars`, `failed`, `tokens` (input, output, cache_read, cache_write), `context` (tokens, window), `cost_usd`.
- **Who is recorded**: writes always; reads only for a named actor (`--by` or `PREP_ACTOR`), as agents run them, because panels, hooks and people reading without a name buried the agent's own events. A harness may set session and actor for its commands; explicit `--by` takes precedence.
- **Sessions and focus**: `Foci` gives each agent's latest focus-moving event: `focus` events and prep events naming an issue except `show` (`MovesFocus`), never an event with `project` (the copy of a command run in another project, [Cross-project work](/features/cross-project.md)); harness tool events never move it. Actors are grouped by `session:<id>`, else `actor:<actor>`.
- **Kinds and readers**: `prep` is the CLI's own events; `focus` is `prep focus <id>|--clear`; `tool` and `request` are harness events passed to `prep activity add` as JSON lines. `prep activity [--max N] [--actor a] [--json]` reads events and foci; `prep tui` shows them in the Agent screen.
- **Claude Code bridge**: reports tool calls, model token usage, context fill and turn cost; see [Claude Code integration](/components/claude-code.md).
- **Codex bridge**: the plugin's `PostToolUse` hook reports supported local tool calls through `prep activity add`, omitting prep Bash calls that prep already records. Its `PreToolUse` hook prefixes Bash calls with `PREP_SESSION` and `PREP_ACTOR=codex` so CLI events group with the Codex session. Codex hooks currently do not expose the token usage and cost fields collected by the Claude bridge. Hook failures fail open, while the direct `.prep` edit guard is an intentional denial for supported Bash and `apply_patch` edits; coverage is not complete. See [Codex integration](/components/codex.md).
