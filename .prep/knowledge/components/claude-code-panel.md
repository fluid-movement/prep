---
type: component
title: Claude Code side panel
description: 'Removed: the prep side panel in Claude Code, replaced by prep tui''s Agent screen; kept for the issues that link it.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T06:36:05Z
status: deprecated
---

# Claude Code side panel

Removed in 01M4DTR02HTJGCNAM4XZ14MWQM (decided in 01M4DTQZPMFCKN0DHDWA95VKTN). The plugin drew the current issue, the project and the session's usage in a Claude Code pane, with `/prep:focus` and `/prep:pane`, inferring the issue from the agent's Bash calls. Rebuilding the TUI in each harness was the wrong layer: the person now runs `prep tui --agent` next to any agent ([TUI Agent screen](/components/tui-agent.md)), prep records the agent's actions itself ([Agent activity](/features/agent-activity.md)), and the plugin only bridges tool calls and tokens ([Claude Code integration](/components/claude-code.md)).
