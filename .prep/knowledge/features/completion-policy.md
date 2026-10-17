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

Completion checks current issue state and staleness, all acceptance criteria when any exist, the result required by the issue kind, and a documentation decision (`--docs` or `--no-impact`). Non-parent decision and research issues can complete from defined or in-progress: decision issues need an `outcome: true` decision, and research issues need findings. Code and manual issues complete only from in-progress; parents, which are never claimed, complete from ready or in-progress. Any human or agent can complete code; there is no actor gate. Parents also require all children resolved. Code issues may include an optional Git commit hash; prep checks its format only. Prep does not execute project checks.

Defined decision and research issues should use `prep guide` for the next step: record the decision outcome or research findings, then complete. If the work should be handed off, `prep ready` and `prep claim` remain available.

The project Definition of Done is guidance. `prep prime` always shows it, and `prep guide` shows it at the check step once all acceptance criteria are checked. It does not gate readiness or completion. `prep dod --add <item>` and `--remove <item>` edit the project list.

Knowledge entries describing changed behavior should be updated with `prep knowledge update` and reviewed as part of implementation. See [Issue lifecycle](/features/lifecycle.md) and [Storage format](/conventions/storage-format.md).
