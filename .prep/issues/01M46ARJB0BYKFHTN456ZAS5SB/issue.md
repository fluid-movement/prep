---
title: 'Claude Code integration: SessionStart hook and slash commands'
kind: code
tags:
  - integration
---

Claude Code runs prep prime at session start through a SessionStart hook, so every session opens with the briefing. The hook runs only when the prep binary is on PATH and stays silent otherwise; installing the binary in cloud sessions waits for the install and update flow.

Slash commands cover the common reads, so the user can see the state of the project from inside the harness while there is no TUI: /prep-status lists every issue with its state, /prep-next shows actionable issues, and /prep-guide <id> shows the guide for one issue and works from it. Each command says so when the binary is missing.

The integration stays thin: it calls the CLI and renders the results, with no FSM logic of its own. It lives in this repository's .claude directory; shipping it to other projects as a plugin waits for the install and update flow.

## Open questions
