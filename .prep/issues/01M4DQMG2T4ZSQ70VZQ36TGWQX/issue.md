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

A Pi session in a prep project starts with the `prep prime` briefing in the model's context, as the Claude Code SessionStart hook provides it.

- Silent without the binary and, through `prime --hook`, outside a prep project; an extension failure never blocks the session.
- The version warning works as in Claude Code: the package's version is compared with the binary's, and the briefing leads with `prep setup --refresh` or `prep update`; development versions are not compared.
- The briefing is added once per session and again after compaction, in a way that keeps the prompt prefix stable for caching.
- Works in interactive, print, JSON and RPC modes.

## Open questions
