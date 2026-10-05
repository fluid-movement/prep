---
type: component
title: OKF store
description: internal/okf — reads the knowledge bundle, extracts links, computes drift of scoped paths since confirmed_commit.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/okf
confirmed_commit: d902990b28eaba324184b9331bc0b3b243104443
---

# OKF store

Applies when changing how knowledge entries are read or checked. Conventions: [Knowledge base conventions](/conventions/knowledge-base.md).

- Walks `.project/knowledge/**.md` except `index.md`; frontmatter is parsed loosely because OKF allows custom keys.
- Links: markdown link targets ending in `.md`; absolute ones are bundle-relative, relative ones resolve against the entry's directory; external links are ignored.
- Drift (only in `prep check`): `git diff --name-only <confirmed_commit> -- <scope>` against the working tree; an unknown commit (for example after a squash merge) is a warning.
