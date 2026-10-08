---
type: component
title: Git helpers
description: internal/gitx — the few git calls prep makes (drift, staging, evidence, HEAD), how their output is read, and why git is optional.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T17:45:26Z
scope:
  - internal/gitx
---

# Git helpers

Applies when prep calls git. Git is versioning, not the system of record ([Design principles](/decisions/design-principles.md)): every caller tolerates git being missing or the directory not being a repository.

- `Run` (and `run` with standard input) executes git in a directory and returns trimmed output. Every call sets `GIT_OPTIONAL_LOCKS=0`, so a read such as `git status` from a TUI reload or `prep check` never takes `index.lock` from a commit the person runs at the same time.
- Path lists are read NUL-separated (`-z`, split by `paths`): `ChangedSince` and `ChangedBetween` (`git diff --name-only`, scopes as pathspecs, globs as `:(glob)`), `FilesInCommit` and `Tracked` (`check-ignore -z --stdin`). The line form quotes non-ASCII and special characters, which would never match the real paths.
- Uses: knowledge drift ([OKF store](/components/okf-store.md): `CommitExists`, `LastCommit`, `IsAncestor`, `Dirty`, `ChangedSince`, `ChangedBetween`), staging every write (`Stage` of the `Tracked` paths, so the ignored own `config.yaml` never fails `git add`), evidence for completions (`AddedIn`, `FilesInCommit`), `prep knowledge confirm` (`Head`) and `prep fix` (`IsTracked`, `Untrack` for a committed own config).
- Tests (`git_test.go`) run against a temporary repository with non-ASCII and spaced file names.
