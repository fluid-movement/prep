---
title: 'Pi pane: Live view following the agent'
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMY0R6D02TSYS40YATN29
tags:
  - integration
  - pi
---

A prep pane in Pi shows the issue the agent works on, with the content and behavior of the Claude Code panel's Live view ([Claude Code side panel](/components/claude-code-panel.md)), in the form the design issue chose.

- **Following**: successful Bash calls with a prep write or `prep guide` on an issue, `prep new`'s output, and edits or writes under `.prep/issues/<id>/` set the focus; reads do not. Each session follows its own agent. The shown issue is the pin, else the focus, else the first claim in `prep prime`, else the project overview.
- **Data and refresh**: `prep show`, `guide`, `list` and `prime` with `--json`; only the newest of overlapping refreshes applies; `prep watch` runs from `session_start` and stops at `session_shutdown`; nothing starts in the extension factory.
- **Drawing**: header, next transition with unmet gates, requirement, open questions, acceptance with progress, DoD, decisions, surroundings, last history lines and priority marks, colored from Pi theme tokens mapped like the Claude Code panel's states.
- **Commands**: focus (pin an issue, or follow the agent again without an id) and pane (open or close, or open on a view), answered without reaching the model.
- **Opening**: open, closed or remember per project, as `userConfig.pane` does in Claude Code.
- Only in the interactive TUI; other modes are unaffected, and a pane failure never changes a tool call.
- Tried live: opening, following guide and writes, ignoring reads, live updates through `prep watch`.

## Open questions
