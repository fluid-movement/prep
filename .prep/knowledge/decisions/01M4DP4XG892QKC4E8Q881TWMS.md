---
type: decision
title: "No pane: prep tui's Agent view follows the agent"
description: "No pane: prep tui's Agent view follows the agent"
date: "2026-10-08"
issue: 01M46ARM9G3492V298TM82KD45
supersedes:
  - 01M4DP4XG1Q64YH150NETX5Z61
---

Rebuilding the TUI in every harness's toolkit (the Claude Code panel, the Pi pane, a theme bridge between them) is the wrong layer. The person runs `prep tui` in a split next to any agent; its Agent view shows the issue the agent works on, a live feed of what it does, and its usage. The binary records an activity stream under `.prep/local/` (gitignored) from every prep command, and the agent's current issue is derived from it, so issue inference moves from TypeScript into prep (also superseding D5). Harnesses add token and tool events through `prep activity add`. The Pi integration is the skill plus one extension: the briefing section (D2) and activity events. No `/prep` command (supersedes D7). Decided by the user while planning 0.2.0.
