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
confirmed_commit: c3050ddbab8ba6fe21b8ccefdd1caa4d26be3cd8
---

# Architecture

Applies when adding a command, a storage adapter or an interface.

```
cmd/prep            entry point
internal/cli        agent interface: commands, JSON output, git staging
internal/domain     model, derived state, FSM and gates, validation, query engine, guide
internal/mdstore    issue store adapter: .prep/issues as markdown
internal/okf        knowledge store adapter: .prep/knowledge as an OKF v0.2 bundle
internal/gitx       the few git calls (staging, drift, commit evidence)
internal/activity   the local activity stream under .prep/local: append, prune, read, derive each actor's current issue (no store imports)
internal/watch      fsnotify watcher over .prep (recursive, debounced), shared by prep watch and the TUI
internal/palette    TUI color tokens, built-in themes, custom theme resolution, derived state colors (no TUI imports)
internal/tui        human interface: screens (Bubble Tea) loading through an injected loader; theme and ui hold the design system
internal/setup      harness integrations installed, refreshed and removed by prep setup
internal/userconfig the user's choices in ~/.config/prep/config.yaml (harnesses, TUI mouse)
internal/update     prep update: download, checksum check and replacement of the binary
```

- Ports are defined in `internal/domain/ports.go`: `IssueStore` (transactional records: `Load`, `Apply`) and `KnowledgeStore` (retrieval: `Load`; writes: `Render` then `Apply`). They are separate because access patterns differ.
- Adapters contain no FSM logic. The domain plans a write as a `domain.Change`; `Tree.CheckWrite` validates the tree with the change applied in memory and rejects writes that introduce errors; the adapter renders and writes it.
- The CLI and the TUI share `domain.Filter` and `Tree.Query`, so both agree on derived states.
- Git is versioning, not the system of record; every git call tolerates git being absent. A database adapter would have to provide what git gives today (history, branch-scoped review) itself; the domain assumes neither.

Details: [domain](/components/domain.md), [markdown store](/components/markdown-store.md), [OKF store](/components/okf-store.md), [CLI](/components/cli.md), [TUI design system](/components/tui-design-system.md), [setup and user configuration](/components/setup.md), [release, install and update](/components/release.md).
