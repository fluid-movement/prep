---
type: feature
title: Agent context
description: 'How agents get context without loading everything: prime at session start, guide for the chosen issue, then the files it points to; knowledge retrieval order.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T08:37:30Z
scope:
  - internal/domain/guide.go
  - internal/cli/read.go
confirmed_commit: 54e8a28a1b0c5a0588d2c4face4c41b3e1f76986
---

# Agent context

Applies when changing how agent sessions receive prep context or discover the next workflow step. Context is layered so each step gives the agent only what it needs.

- **Session briefing**: the Claude Code plugin and Codex plugin run `prep prime --hook` at session start in a prep project. It provides capped pointers: bootstrap alert, top-level parents with progress, actionable issues, stale issues, claims, check failures, and recently worked-on issues.
- **Task guidance**: Codex `UserPromptSubmit` hooks add a task-time reminder to search for a matching issue, check `prep next`, and run `prep guide <id>` before changing code. A complete issue ID in the prompt triggers a live `prep guide`. Claude Code receives its workflow through the prep skill and agent context.
- **Guide**: `prep guide <id>` tells the agent the current step, unmet gates, files to read and exact commands for outputs and transitions. Follow the guide before opening records or changing code.
- **Limits**: plugin hooks must be trusted in Codex before they run. They provide automatic context and supported-tool guardrails, not perfect semantic enforcement; the skill remains the detailed workflow. No prep project needs an `AGENTS.md` file for the integration to work.
## Planning and tracking facts

Prime includes actionable issue tags and a capped list of unresolved dependencies blocking leaf work, with blocked_count in JSON and a footer for omitted blocked issues. Counts reflect all blocked work even when the list is capped. Shared prompt guidance asks the agent to read requirements and metadata before recommendations and forbids guessing effort from titles. See [Tracking implementation work](/decisions/tracking.md) for search/create behavior, explicit opt-outs and scope through Stop reviews.
