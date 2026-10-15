---
type: feature
title: Lean knowledge policy
description: Decisions, constraints and navigation first; incremental maintenance and bounded comparative knowledge experiments.
generated:
  by: codex
  at: 2026-10-15T08:04:07Z
scope:
  - eval/experiments/knowledge-policy
  - internal/cli/knowread.go
---

# Lean knowledge policy

Applies when writing knowledge, deciding what to retrieve, or evaluating whether it helps an agent. Keep OK Format as the storage contract; content policy does not need a new format or lifecycle gate.

## What to retain

Keep decisions and their rationale, constraints that are easy to miss, and precise navigation hints. Inspect the current code for implementation facts. Start with a small project briefing, then use `prep knowledge find` and section-level `show` for the current question. When an entry is touched, remove duplicate inventories incrementally while preserving useful reasoning; do not prune the whole bundle indiscriminately. When shortening entries, retain mandatory validation commands and environment prerequisites. Track any removed constraint as information loss in comparisons. Put project-wide required verification in the completion policy and small briefing; targeted component retrieval may miss a global convention. See [Knowledge conventions](/conventions/knowledge-base.md).

Documentation is implementation work: update affected entries before completion or explicitly review them unchanged, or justify no impact. Candidate matches indicate possible relevance. A stored `confirmed_commit` and scoped change information are review provenance; neither proves semantic correctness. See [Knowledge lifecycle](/features/knowledge-lifecycle.md) and [CLI contracts](/features/cli-contracts.md).

## Comparing usefulness

`eval/experiments/knowledge-policy` is a bounded, task-specific Codex pilot over the existing webapp fixture and natural matching-issue task, reusing the fixture builder. It does not replace `prep eval`, its plugin A/B work or its Claude-specific live runner. Live Claude runs remain outside the authorized scope of this pilot.

Compare code-only advisory context, current entries, lean entries and deliberately stale entries at the same pinned code/issue snapshot with identical task, workflow skill and completion policy. Keep a minimal common overview so bootstrap requirements do not change between cells. Record exact prompts, model request versus reported model, CLI version, seed, repetitions and retrieval conditions. Separate primary contract checks from exploratory correctness, missed constraints, token/time use and knowledge-maintenance counts. Preserve failed attempts, unavailable counters and sample sizes. The research issue 01M4TSWY40Z4473D9SPRTT711C holds the actual observations and limitations; a small single-task pilot cannot establish superiority or justify hard CI thresholds.
