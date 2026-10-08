---
title: 'Pi integration design: package, pane, briefing and tests'
kind: research
parent: 01M46ARM9G3492V298TM82KD45
tags:
  - integration
  - pi
---

Before building the Pi integration, settle how each Claude Code feature maps onto Pi 1.x and how the package is built, shipped and tested, following Pi's documented extension and package practices. The findings decide the children of the parent and become the start of a Pi integration knowledge entry.

- **Pane**: Pi has no split pane. Compare a right-anchored, non-capturing overlay (`ctx.ui.custom` with `overlay: true`, `nonCapturing`, `visible` by terminal width), a widget above or below the editor (`ctx.ui.setWidget`), and a combination (compact widget, full pane when wide); include how each behaves in fullscreen and regular TUI mode, how the person scrolls it and how it gets keyboard focus.
- **Briefing**: how the `prep prime` briefing reaches the model at session start and after compaction without changing the system prompt on every turn (prompt cache), and how the version check works without a `plugin.json`.
- **Following**: what a Bash call and an edit or write look like in `tool_result` events, so the Claude Code panel's issue inference (`plugins/claude-code/hooks/infer.ts`) can be shared or reused rather than copied.
- **Commands**: names (Pi commands carry no plugin namespace), and how the status, next and guide commands hand prep's output to the model.
- **Package**: layout under `plugins/pi`, a `pi` manifest reachable from a git source of this repository pinned to a tag, peer dependencies on the host packages, development loading from the checkout, and whether a project skill copy (as `.claude/skills` is for cloud sessions) would collide with the package skill.
- **Tests**: how the extension is tested here (unit tests of pure parts, an SDK session with a scripted model, a live check in Pi).

## Open questions
