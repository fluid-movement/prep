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

Completion checks current issue state and staleness, checked acceptance criteria, the result required by the issue kind, and a documentation decision (`--docs` or `--no-impact`). Code issues may include an optional Git commit hash; prep checks its format only. Prep does not execute project checks.

The project Definition of Done is guidance. `prep prime` always shows it, and `prep guide` shows it at the check step once all acceptance criteria are checked. It does not gate readiness or completion. `prep dod --add <item>` and `--remove <item>` edit the project list.

Knowledge entries describing changed behavior should be updated with `prep knowledge update` and reviewed as part of implementation. See [Issue lifecycle](/features/lifecycle.md) and [Storage format](/conventions/storage-format.md).
