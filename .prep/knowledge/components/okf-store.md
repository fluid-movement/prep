---
type: component
title: OKF store
description: internal/okf — reads the knowledge bundle (links, drift) and writes entries for prep knowledge new, update and confirm.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/okf
confirmed_commit: 3f507f8eea6dba16070e7e432d4f9995d0bbed98
---

# OKF store

Applies when changing how knowledge entries are read, checked or written. Conventions: [Knowledge base conventions](/conventions/knowledge-base.md).

- Walks `.prep/knowledge/**.md` except `index.md`; frontmatter is parsed loosely because OKF allows custom keys.
- Links: markdown link targets ending in `.md`; absolute ones are bundle-relative, relative ones resolve against the entry's directory; external links are ignored.
- Drift (only in `prep check` and `prep knowledge confirm --drifted`): `git diff --name-only <confirmed_commit> -- <scope>` against the working tree; an unknown commit (for example after a squash merge) is a warning.
- Index files (`index.go`): `IndexFiles` computes an OKF `index.md` for every directory holding entries (Entries section `* [Title](/path) - description`, Directories section with entry counts; the root adds `okf_version: "0.2"` frontmatter). `WriteIndexes` writes changed ones atomically and removes unneeded ones; it runs after every knowledge write and from `prep fmt` and `prep fix`. `Load` reports missing, outdated or unneeded index files as K007 and skips the reserved `log.md`.
- Writes (`write.go`): `Render` builds the file a `domain.KnowledgeEdit` would produce and parses it back into an `Entry`, so `CheckWrite` validates fields and links before anything is written; `Apply` writes exactly that content atomically and refuses a file changed since `Render` read it (`E_CONFLICT`).
- New entries get frontmatter in the order type, title, description, status, generated (by, at), scope, confirmed_commit. Updates edit the YAML node in place: only the given keys change, unknown OKF keys (verified, sources, resource) and their order stay. Clearing the scope also drops `confirmed_commit`.
