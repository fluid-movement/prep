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
---

# Agent activity

Applies when changing what prep records about agents, the event format, or anything that reads the stream (the TUI's Agent view, prime, harness bridges). Decided in 01M4DTQZPMFCKN0DHDWA95VKTN: harnesses stop drawing their own panels; the person runs `prep tui` next to the agent and watches this stream.

- **Where**: `.prep/local/activity.jsonl`, per machine, ignored through `.prep/.gitignore` (`local/`; P007 warns when `.prep/local` exists unignored and `prep fix` adds the line). It is UI state, never a record: reads still never change records, and nothing in it is staged or committed. Each git worktree has its own.
- **Writing** (`internal/activity`): `Append` writes one JSON line per event with `O_APPEND` in a single write below 4 KiB (every string field clipped to 300 characters, the target dropped if the line is still too long, and an event that cannot fit refused), so concurrent agents never interleave lines; past 1 MiB the file is rewritten (temp file, rename) with its newest lines, at most 2,000 and half the limit. `Read`/`Parse` skip lines that do not parse.
- **Event**: `at`, `actor`, `kind`, optional `session`, `op` (the prep command or the harness tool), `issue`, `verb` (plain words: "checked criterion 3", "recorded decision D2: …", "read the guide"), `target` (issue title, knowledge query or entry, file), `area` (issue, knowledge, code, prep, other), `chars` (output read back), `failed`, `tokens` (input, output, cache_read, cache_write), `context` (tokens, window), `cost_usd`.
- **Who is recorded**: writes always; reads only for a named actor (`--by` or `PREP_ACTOR`, `app.named`), as agents run them, because panels, hooks and people reading without a name buried the agent's own events (seen live: the Claude Code panel's refreshes wrote hundreds of reads).
- **Sessions**: the binary stamps `PREP_SESSION` on its events and `prep activity add` fills it in when missing; a harness exports it for the agent's shell and sends it from its bridge, so an agent is one `Agent()` key (`session:<id>`, else `actor:<actor>`) whatever actor names its prep calls (`claude-code/opus-5.5`) and its harness (`claude-code/2.1.294`) use. `Foci` and `Agents` key by it.
- **Kinds**: `prep` — the binary, after a command succeeds (`app.note`/`noteRead`, appended by `Main` through `recordEvent`; reads carry the size they printed through a counting writer): every issue write (`new`, `edit`, records, transitions), `guide`, `show`, `list`, `next`, `prime`, knowledge `list|find|show|new|update|confirm`. `focus` — `prep focus <id>` or `--clear`. `tool` and `request` — harnesses through `prep activity add` (JSON lines on stdin, validated by `Event.Validate`; `at` and `actor` filled in when missing). prep never reads harness files. A failed append never fails a command.
- **Current issue**: `Foci` gives each agent's latest focus-moving event: `focus` events and `prep` events naming an issue except `show` (`MovesFocus`); a cleared focus drops the actor. Harness tool events never move it. The actor is `--by`/`PREP_ACTOR`, else `cli/prep-<version>`.
- **Readers**: `prep activity [--max N] [--actor a] [--json]` (events and the foci, newest focus first); `prep prime` lists up to three foci from the last day ("Recently worked on", JSON `focus`) so a resumed or compacted session sees where it was, whatever actor its hook runs as.
