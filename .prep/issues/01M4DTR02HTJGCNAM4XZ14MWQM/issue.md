---
title: Slim Claude Code plugin with an activity bridge
kind: code
parent: 01M4DTQZPMFCKN0DHDWA95VKTN
depends_on:
  - 01M4DTQZTJ2ZXGR04GWNMPB6WZ
tags:
  - integration
---

The Claude Code plugin keeps what only a harness can do and drops what the TUI's Agent view now does.

- **Removed**: the side panel module (pane, Live, Project and Usage views, issue inference), `/prep:status`, `/prep:next`, `/prep:guide`, `/prep:focus`, `/prep:pane`, and the plugin's `pane` option.
- **Kept**: the skill and the SessionStart briefing hook.
- **Activity bridge**: hooks report each tool call (tool, area, output size) and each request's tokens, the context fill and the cost to `prep activity add`, in prep's event format, so the Agent view shows Claude Code's figures. It replaces the token-ledger plugin, which leaves the marketplace; `prep setup` stops reinstalling it.
- A failing bridge never changes a tool call or a turn, and it stays silent outside a prep project or without the binary.
- The Claude Code knowledge entries describe the slim plugin; the side panel entry is removed.

## Open questions
