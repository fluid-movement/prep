---
title: 'Claude Code integration: SessionStart hook and slash commands'
kind: code
---

Claude Code runs prep prime at session start through a SessionStart hook and offers slash commands for common reads. In cloud sessions without the binary, the hook can install it. The integration stays thin: it calls the CLI with --json and renders results, with no FSM logic of its own.

## Open questions

- Should the hook install the binary in cloud sessions, or only run it when present?

## Open questions
