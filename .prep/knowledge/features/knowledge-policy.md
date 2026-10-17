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

## What to retain

Keep decisions and their rationale, constraints that are easy to miss, and precise navigation hints. Inspect current code for implementation facts. Start with a small project briefing, then use `prep knowledge find` and section-level `show` for the current question. When an entry is touched, remove duplicate inventories incrementally while preserving useful reasoning; do not prune the whole bundle indiscriminately. When shortening entries, retain mandatory validation commands and environment prerequisites. Track removed constraints as information loss in comparisons. The project Definition of Done is shown as guidance in the briefing and guide; it does not impose a completion gate. See [Knowledge conventions](/conventions/knowledge-base.md).
