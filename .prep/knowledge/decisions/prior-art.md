---
type: decision
title: Prior art
description: What prep took from and avoided in Backlog.md and Beads, and why spec-driven tools are a different category.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T08:37:30Z
---

# Prior art

Applies when a design question has an established answer in a similar tool. Decided in the original design.

Task tracking is largely solved by existing tools; what prep adds is the define/enrich lifecycle with drift detection, state derived from records with one record per transition, and the agent-maintained knowledge base.

| Tool | What it does | Taken | Avoided |
| --- | --- | --- | --- |
| Backlog.md (github.com/MrLesk/Backlog.md) | markdown tasks in the repository, CLI, MCP, terminal Kanban, acceptance criteria, Definition of Done | instructions served by the binary; project-level Definition of Done; dogfooding | titles in file names; status as a field; scanning branches to show cross-branch state |
| Beads (github.com/steveyegge/beads) | dependency graph, ready work queue, claims, hash IDs | session-start briefing (`prime`); JSON output everywhere; contract tests | syncing two representations (SQLite and JSONL); many records in one file |

Spec-driven tools such as GitHub's Spec Kit, Kiro and OpenSpec focus on per-feature specification documents rather than a persistent tracker with a lifecycle.

Related: [Design principles](/decisions/design-principles.md), [Out of scope](/decisions/out-of-scope.md).
