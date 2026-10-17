---
type: component
title: Architecture
description: One Go binary; the domain layer sits between agent/human interfaces and storage ports with markdown and OKF adapters.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - cmd/**
  - internal/**
confirmed_commit: d73c7a0eab6c048393bb0c2f9462fe34e3c655c7
---

# Architecture

The CLI and TUI use the same domain model and transition planners. Storage adapters load and persist records; they do not implement lifecycle rules.

```
cmd/prep            entry point
internal/cli        agent interface: commands, JSON output, git staging
internal/domain     model, derived state, transitions, validation, query engine, guide
internal/mdstore    issue store adapter: .prep/issues as markdown
internal/okf        knowledge store adapter: .prep/knowledge as an OKF v0.2 bundle
internal/gitx       the few git calls (staging, drift, commit evidence)
internal/activity   the local activity stream under .prep/local: append, prune, read, derive each actor's current issue (no store imports)
internal/watch      fsnotify watcher over .prep (recursive, debounced), used by the TUI
internal/palette    TUI color tokens, the dark and light themes, derived state colors (no TUI imports)
internal/tui        human interface (Bubble Tea); screens load through an injected loader, theme and ui hold the design system
internal/setup      harness integrations installed, refreshed and removed by prep setup
internal/userconfig the user's choices in ~/.config/prep/config.yaml (harnesses, TUI mouse/layout, prompt hook context)
internal/projects   the machine-local registry of project roots ($XDG_STATE_HOME/prep/projects.json), project identity and resolution (--project)
internal/update     prep update: download, checksum check and replacement of the binary
internal/eval       agent behavior tests: fixtures, isolated runs of Claude Code, transcripts, assertions, judge, matrix, results, variants (no imports of cli)
```

Ports are defined in `internal/domain/ports.go`: `IssueStore` (checked record writes: `Load`, `Apply`) and `KnowledgeStore` (retrieval: `Load`; writes: `Render` then `Apply`). They are separate because access patterns differ.

The domain plans a write as a `domain.Change`; `Tree.CheckWrite` validates the tree with the change applied in memory and rejects writes that introduce errors; the adapter renders and writes it. Git provides versioning, not the system of record; every git call tolerates git being absent. A database adapter would have to provide history and branch-scoped review itself; the domain assumes neither.

The CLI and the TUI share `domain.Filter` and `Tree.Query`, so both agree on derived states. Lifecycle hook behavior is shared in `internal/cli/hook.go` and `hook_finish.go`; harnesses provide event adapters, not workflow logic. See [Shared lifecycle hooks](/components/hooks.md) and [Issue lifecycle](/features/lifecycle.md).
