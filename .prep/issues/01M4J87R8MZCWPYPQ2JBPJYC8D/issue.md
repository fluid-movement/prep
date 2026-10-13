---
title: 'Codex CLI: hooks make sure prep calls are made'
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - integration
---

Build and document a Codex plugin integration for prep, installed and refreshed through `prep setup`. The plugin packages the prep skill and Codex lifecycle hooks. In a repository initialized with prep, hooks inject a current `prep prime` briefing at session start and provide task-relevant guidance on user prompts so agents consult `prep next` or `prep guide <id>` before acting on tracked work. Hooks should identify likely prep-relevant work and redirect unsafe direct edits to `.prep` records toward the CLI; failures in the integration must not prevent ordinary tools from running. Where Codex exposes useful tool events, report them to `prep activity add` as the Claude Code bridge does. Document plugin installation, the one-time hook trust/review requirement, and behavior when hooks are unavailable or not trusted. Do not require projects using prep to add `AGENTS.md`.

## Open questions
