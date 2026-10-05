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
confirmed_commit: db27d6b1996f239fe8e18dc951b950ad5cab77ec
---

# Architecture

Applies when adding a command, a storage adapter or an interface.

```
cmd/prep            entry point
internal/cli        agent interface: commands, JSON output, git staging/commits
internal/domain     model, derived state, FSM and gates, validation, query engine, guide
internal/mdstore    issue store adapter: .prep/issues as markdown
internal/okf        knowledge store adapter: .prep/knowledge as an OKF v0.2 bundle
internal/gitx       the few git calls (staging, commits, drift, commit evidence)
```

- Ports are defined in `internal/domain/ports.go`: `IssueStore` (transactional records: `Load`, `Apply`) and `KnowledgeStore` (retrieval: `Load`). They are separate because access patterns differ.
- Adapters contain no FSM logic. The domain plans a write as a `domain.Change`; `Tree.CheckWrite` validates the tree with the change applied in memory and rejects writes that introduce errors; the adapter renders and writes it.
- The CLI and the future TUI share `domain.Filter` and `Tree.Query`, so both agree on derived states.
- Git is versioning, not the system of record; every git call tolerates git being absent.

Details: [domain](/components/domain.md), [markdown store](/components/markdown-store.md), [OKF store](/components/okf-store.md), [CLI](/components/cli.md).
