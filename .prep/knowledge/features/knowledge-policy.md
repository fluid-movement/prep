---
type: feature
title: Lean knowledge policy
description: Decisions, constraints and navigation first; incremental knowledge maintenance.
generated:
  by: codex
  at: 2026-10-15T08:04:07Z
scope:
  - internal/cli/knowread.go
---

# Lean knowledge policy

Applies when writing knowledge, deciding what to retrieve, or evaluating whether it helps an agent. Keep OK Format as the storage contract; content policy does not need a new format or lifecycle gate.

## What to retain

Keep decisions and their rationale, constraints that are easy to miss, and precise navigation hints. Inspect the current code for implementation facts. Start with a small project briefing, then use `prep knowledge find` and section-level `show` for the current question. When an entry is touched, remove duplicate inventories incrementally while preserving useful reasoning; do not prune the whole bundle indiscriminately. When shortening entries, retain mandatory validation commands and environment prerequisites. Track any removed constraint as information loss in comparisons. Put project-wide required verification in the completion policy and small briefing; targeted component retrieval may miss a global convention. See [Knowledge conventions](/conventions/knowledge-base.md).

Documentation is implementation work: update affected entries before completion or explicitly review them unchanged, or justify no impact. Candidate matches indicate possible relevance. A stored `confirmed_commit` and scoped change information are review provenance; neither proves semantic correctness. See [Knowledge lifecycle](/features/knowledge-lifecycle.md) and [CLI contracts](/features/cli-contracts.md).
