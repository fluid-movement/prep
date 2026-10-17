---
type: feature
title: Knowledge in the issue lifecycle
description: Documentation decisions and knowledge updates during issue completion
generated:
  by: codex
  at: 2026-10-14T12:48:05Z
---

# Knowledge in the issue lifecycle

Agents use `prep knowledge find` and `show` to inspect relevant entries while working. Completion records a documentation decision: name changed or relevant entries with `--docs`, or explain no impact with `--no-impact`. A reference to an entry documents the decision; prep does not claim that the content is correct. Update affected entries through `prep knowledge update` and review them against their scoped code.

The project Definition of Done appears as guidance in the briefing and at the check step of `prep guide`. Completion requirements are described in [Completion requirements](/features/completion-policy.md); record formats are in [Storage format](/conventions/storage-format.md).
