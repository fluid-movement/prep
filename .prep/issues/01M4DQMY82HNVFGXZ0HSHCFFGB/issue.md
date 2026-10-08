---
title: 'Pi pane: Usage view'
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMG9ZVZT172V0SC967R2C
tags:
  - integration
  - pi
---

The Pi pane gets the Usage view of the Claude Code panel: the session's context fill and cost, request tokens (input, cache read and write, output), tool output read back per area (knowledge base including `prep knowledge list|find|show`, issue files, other prep commands, code) and the largest knowledge reads, attributed to the issue being worked on.

In Claude Code this needs the separate token-ledger plugin; in Pi the extension sees tool results and the session's entries itself, so it measures without another package, from the same area rules as the ledger (shared or kept in step with `plugins/token-ledger`). Measurement only: prep itself never reads it.

## Open questions
