---
title: Pi setup
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
depends_on:
  - 01M48721G8Y21HT0ZGAXB5QZF2
tags:
  - integration
  - pi
---

prep works in Pi as well as it works in Claude Code: an agent in a Pi session gets the same briefing, skill, commands and side pane, and `prep setup` installs and updates the integration like the Claude Code plugin. The integration is a Pi package served from this repository, a thin TypeScript adapter over the CLI's JSON output with no FSM logic ([Stack](/decisions/stack.md)).

Feature parity with [Claude Code integration](/components/claude-code.md) and [Claude Code side panel](/components/claude-code-panel.md) comes first; its children are the parity work:

- the prep skill, identical to `internal/cli/skill.md` and worded so it holds in either harness
- the session-start briefing from `prep prime`, silent without the binary or outside a prep project, with the version warning
- the status, next and guide commands
- the pane: Live view following the agent, focus and pane commands, Project and Usage views, the prep theme
- installation, refresh and removal through `prep setup`, pinned to the binary's version

The integration follows Pi's extension practices so agents stay efficient and reliable: long-lived resources start at `session_start` and stop idempotently at `session_shutdown`, terminal UI is guarded by the mode so print, JSON and RPC runs keep working, the briefing reaches the model without breaking the prompt cache, and a failure in the extension never changes a tool call.

What Pi can do beyond Claude Code (own tools, prompt shaping, and so on) is discussed once parity is reached and comes as separate issues.

## Open questions
