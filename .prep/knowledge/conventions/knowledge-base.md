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
- **Bootstrapping**: a knowledge base counts as bootstrapped once `/overview.md` exists. `prep init` creates a bootstrap parent and a survey research issue (tagged `bootstrap`; `--no-bootstrap` skips them, `prep knowledge bootstrap` creates them later). Until the overview exists, `prep prime` and `prep guide` lead with an alert and `prep knowledge new` refuses other entries. The survey writes a map of proposed entries as findings for the user to review, then the overview as a draft; each area of the map becomes a child issue that writes its entries as drafts, which turn stable after review.
- Write entries with `prep knowledge new` and `prep knowledge update`, not by editing files. After re-checking an entry against its scoped code, run `prep knowledge confirm <entry>` (or `--drifted`) to set `confirmed_commit`; `prep check` warns when scoped paths changed since.
