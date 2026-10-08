---
title: Pi session briefing
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMY0R6D02TSYS40YATN29
tags:
  - integration
  - pi
---

A Pi session in a prep project has the `prep prime --hook` briefing in the model's context, as the Claude Code SessionStart hook provides it, as the system prompt section `prep` (D2).

- **When**: computed at `session_start` (startup, resume, new, fork, reload) and after `session_compact`; between those the text stays the same, so Pi appends no delta and the prompt prefix stays cached. A `before_agent_start` handler sets the section from the last computed text and never runs prep itself.
- **Silence**: no section without the binary or, through `prime --hook`, outside a prep project; a failing or slow prime never blocks the session or a turn.
- **Version warning**: the package's version is compared with the binary's and the briefing leads with `prep setup --refresh` (older package) or `prep update` (older binary), as with the Claude Code plugin; development versions are not compared. prime learns the package version without reading a `plugin.json`.
- Works in interactive, print, JSON and RPC modes; covered by the SDK session test (section present, unchanged across turns, refreshed after compaction) and seen in a live session.

## Open questions
