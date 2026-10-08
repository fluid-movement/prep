---
type: component
title: OKF store
description: internal/okf — reads the knowledge bundle (links, drift) and writes entries for prep knowledge new, update and confirm.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/okf
confirmed_commit: 361f9e651425cf3b1bca9dd0fa626e954b6e9a3b
---

# OKF store

Applies when changing how knowledge entries are read, checked or written. Conventions: [Knowledge base conventions](/conventions/knowledge-base.md).

- Walks `.prep/knowledge/**.md` except `index.md` and `log.md`; frontmatter is parsed loosely because OKF allows custom keys. `splitEntry` closes the frontmatter only on a line that is exactly `---` (a leading BOM and CRLF are tolerated); an entry whose frontmatter does not parse is a diagnostic and stays out of the tree and the index.
- `parseEntry` keeps the text after the frontmatter as `Entry.Body` (trimmed), for display; validation and retrieval do not read it.
- Links: markdown link targets ending in `.md`; absolute ones are bundle-relative, relative ones resolve against the entry's directory; external links are ignored.
- Drift (only in `prep check` and `prep knowledge confirm --drifted`): scoped paths changed since the later of `confirmed_commit` and the last commit that changed the entry file (`gitx.LastCommit`, if it descends from `confirmed_commit`): writing or confirming an entry is its review, so code and entry can share one commit (01M4AVPH9A56WJ4JFMYZT476H0). Against the working tree (`gitx.ChangedSince`), except while the entry file itself has uncommitted changes (`gitx.Dirty`): then only up to HEAD (`gitx.ChangedBetween`), since the entry being written covers the uncommitted changes. An entry without `confirmed_commit` counts from its last commit alone, so a scoped entry is checked from its first commit on; one never committed has nothing to drift from. An unknown commit (for example after a squash merge) is a warning. `gitx` reads path lists NUL-separated (`-z`), so non-ASCII names match, and runs git with `GIT_OPTIONAL_LOCKS=0`, so a read never takes `index.lock` from a commit running alongside.
- Index files (`index.go`): `IndexFiles` computes an OKF `index.md` for every directory holding entries (Entries section `* [Title](/path) - description`, Directories section with entry counts; the root adds `okf_version: "0.2"` frontmatter). `WriteIndexes` writes changed ones atomically and removes unneeded ones; it runs after every knowledge write and from `prep fmt` and `prep fix`. `Load` reports missing, outdated or unneeded index files as K007 and skips the reserved `log.md`.
- Writes (`write.go`): `Render` builds the file a `domain.KnowledgeEdit` would produce and parses it back into an `Entry`, so `CheckWrite` validates fields and links before anything is written; `Apply` writes exactly that content atomically and refuses a file changed since `Render` read it (`E_CONFLICT`). `Render` also refuses a new entry over an existing file (one that failed to load), and `PlanKnowledge` refuses the reserved names `index.md` and `log.md`.
- New entries get frontmatter in the order type, title, description, status, generated (by, at), scope, confirmed_commit. Updates edit the YAML node in place: only the given keys change, unknown OKF keys (verified, sources, resource) and their order stay. Clearing the scope also drops `confirmed_commit`.
