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
confirmed_commit: b5239826093409fdd1b095d4b70a807d84db8687
---

# Agent activity

Applies when changing what prep records about agents, the event format, or anything that reads the stream (the TUI's Agent view, `prime`, harness bridges). Harnesses report to the local stream; the person runs `prep tui` next to the agent.

- **Where**: `.prep/local/activity.jsonl`, per machine, ignored through `.prep/.gitignore` (`local/`; P007 warns when `.prep/local` exists unignored and `prep fix` adds the line). It is UI state, never a record: reads still never change records, and nothing in it is staged or committed. Each git worktree has its own.
- **Writing** (`internal/activity`): `Append` writes one JSON line per event with `O_APPEND` in a single write below 4 KiB (every string field clipped to 300 characters, the target dropped if the line is still too long, and an event that cannot fit refused), so concurrent agents never interleave lines; past 1 MiB the file is rewritten (temp file, rename) with its newest lines, at most 2,000 and half the limit. `Read`/`Parse` skip lines that do not parse.
- **Event**: `at`, `actor`, `kind`, optional `session`, `previous` (the predecessor session on a session start), `op` (the prep command, harness tool, or session lifecycle operation), `issue`, `verb`, `target`, `area` (issue, knowledge, code, prep, other), `chars`, `failed`, `tokens` (input, output, cache_read, cache_write), `context` (tokens, window), `cost_usd`.
- **Who is recorded**: writes always; reads only for a named actor (`--by` or `PREP_ACTOR`), as agents run them, because panels, hooks and people reading without a name buried the agent's own events. A harness may set session and actor for its commands; explicit `--by` takes precedence.
- **Sessions and focus**: `Foci` gives each agent's latest focus-moving event: `focus` events and prep events naming an issue except `show` (`MovesFocus`), never an event with `project` (the copy of a command run in another project, [Cross-project work](/features/cross-project.md)); harness tool events never move it. Actors are grouped by `session:<id>`, else `actor:<actor>`. A `session` event requires a session ID and `op: started|ended`; only `started` may carry `previous`, which must name a different session. `AgentKeys` resolves linked IDs to the newest successor, so `Foci` and prime's Recently worked on retain the predecessor's issue until a later focus change or clear. `Agents(events, now)` counts each linked chain once, omits ended successor sessions even if late tool events arrive, and hides any agent with no events for 30 minutes (fixed). These filters affect the Agent view, never the stored events or historical foci.
- **Kinds and readers**: `prep` is the CLI's own events; `focus` is `prep focus <id>|--clear`; `tool`, `request` and `session` are harness events passed to `prep activity add` as JSON lines. `prep activity [--max N] [--actor a] [--json]` reads events and foci; `prep tui` shows them in the Agent screen.
- **Claude Code bridge**: reports session lifecycle, tool calls, model token usage, context fill and turn cost; see [Claude Code integration](/components/claude-code.md).
- **Codex bridge**: the plugin's `PostToolUse` hook reports supported local tool calls through `prep hook after-tool`, taking actor/session from the harness event and omitting prep-only Bash calls that prep records itself. Session context asks agents to set `PREP_SESSION` and `PREP_ACTOR=codex` only on prep CLI invocations; explicit --by overrides the actor. Other shell commands keep their usual form, preserving approved command matching. `PreToolUse` guards direct record access without prefixing, rewriting or auto-approving ordinary calls. Codex hooks do not expose the token usage and cost fields collected by the Claude bridge. Hook failures fail open, while the direct `.prep` edit guard is an intentional denial for supported Bash and `apply_patch` edits; coverage is not complete. See [Codex integration](/components/codex.md).
