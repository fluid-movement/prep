---
type: feature
title: Completion requirements
description: Completion transition requirements and project Definition of Done guidance
generated:
  by: codex
  at: 2026-10-13T16:47:43Z
status: stable
---

# Completion requirements

Completion checks current state and staleness, checked acceptance criteria when any exist, and the result required by the issue kind. Non-parent decision and research issues can complete from defined or in-progress: decisions require a linked knowledge decision record with `outcome: true` (`G_OUTCOME`), and research requires a linked finding record (`G_FINDINGS`). Standalone records do not satisfy issue gates. Code/manual issues require in-progress; parents require ready or in-progress and resolved children. Any actor may complete code. Optional code commit hashes are checked for format only. Prep does not execute project checks.

Use `prep guide` for the next step. Record results with `prep decide <issue> --title T --description D --outcome` or `prep finding <issue> --title T --description D`, optionally with body and scopes, then `prep complete <issue>`. Ready and claim remain available for handoff. Completion accepts no documentation flags and has no knowledge-documentation gate.

Definition of Done remains project guidance shown by prime and guide at the check step. See [Issue lifecycle](/features/lifecycle.md) and [Storage format](/conventions/storage-format.md).
