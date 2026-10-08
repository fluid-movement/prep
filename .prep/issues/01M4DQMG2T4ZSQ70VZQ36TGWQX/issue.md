---
title: 'Pi extension: briefing and activity events'
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMY0R6D02TSYS40YATN29
tags:
  - integration
  - pi
---

The Pi extension gives a Pi session in a prep project what the Claude Code plugin gives a Claude Code session: the briefing and the activity events behind `prep tui`'s Agent view.

- **Briefing** (D2): the `prep prime --hook` output as the system prompt section `prep`, computed at `session_start` (startup, resume, new, fork, reload) and after `session_compact`, unchanged between them so Pi appends no delta and the prompt prefix stays cached. A `before_agent_start` handler sets the section from the last computed text and never runs prep itself.
- **Silence**: no section and no events without the binary or, through `prime --hook`, outside a prep project; a failing or slow prep call never blocks the session or a turn.
- **Version warning**: the package's version is compared with the binary's and the briefing leads with `prep setup --refresh` or `prep update`, as with the Claude Code plugin; development versions are not compared.
- **Activity**: successful tool results (edits, writes, reads, bash with their area and output size) and each request's tokens, context fill and cost go to `prep activity add` in prep's event format, as the Claude Code bridge sends them, so the Agent view shows the same for both harnesses.
- Works in interactive, print, JSON and RPC modes; covered by the SDK session test and seen in a live session next to `prep tui --agent`.

## Open questions
