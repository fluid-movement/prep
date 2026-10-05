---
type: convention
title: Knowledge base conventions
description: How knowledge entries are written, typed, linked, scoped and kept current; precedence when sources disagree.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Knowledge base conventions

Applies when creating or updating an entry in `.prep/knowledge`, typically in the documentation step of `prep complete`.

- Follows OKF v0.2. Required frontmatter: `type` (feature, component, decision, convention, pitfall, overview), `title`, `description` (the triage line `prep guide` lists). Optional: `status` (draft | stable; stale entries are deleted, not deprecated), `generated`, `verified`, `resource`, `sources`, and custom `scope` (paths or globs) and `confirmed_commit`.
- One concept per entry, small (lint warns above 8 KiB). Concept IDs are readable paths; entries are never renamed; links are bundle-relative like `/components/cli.md` and must resolve.
- Current state only; history lives in the issue that changed it. Precedence when sources disagree: code over knowledge base over old issue decisions; fix the stale entry.
- Bodies say when the entry applies, use precise references (paths, symbols, commands) and link related entries densely. Start from the [overview](/overview.md).
- Update `confirmed_commit` whenever an entry is re-checked against its scoped code; `prep check` warns when scoped paths changed since.
