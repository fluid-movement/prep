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
confirmed_commit: b5239826093409fdd1b095d4b70a807d84db8687
---

- **Guide**: `prep guide <id>` tells the agent the current step, unmet conditions, files to read and exact commands for outputs and transitions. Follow the guide before opening records or changing code. It offers up to five relevant knowledge candidates with targeted reads. Completion requires a current, non-stale issue; checked criteria when any exist; the kind's result; and a documentation decision. Non-parent decision and research issues may complete from defined or in-progress; code/manual and parent issues complete only from in-progress. Defined decision and research guides direct the next step to record the outcome or findings before completing. Readiness and claiming remain available for handoff. The project Definition of Done is guidance: `prep prime` always shows it, and `prep guide` shows it at the check step once all acceptance criteria are checked. Agents execute project checks and update affected knowledge entries. See [Completion requirements](/features/completion-policy.md) and [Knowledge lifecycle](/features/knowledge-lifecycle.md).
