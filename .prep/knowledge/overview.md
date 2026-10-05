---
type: overview
title: Project overview
description: Entry point to prep's knowledge base; what prep is and links to every other entry.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# prep

prep is a workflow engine and memory for coding agents, with a human supervising. It enforces the steps an issue goes through and holds the context agents need to implement it. It is change management for agent-built projects: issues record what should change and what changed; this knowledge base records what is true now. Completing an issue requires a documentation decision, so every completed issue moves this knowledge base forward.

Applies always; start here and follow links.

## Entries

- [Design principles](/decisions/design-principles.md): the rules every change must respect.
- [Architecture](/components/architecture.md): domain layer, ports, adapters, interfaces.
- [Stack](/decisions/stack.md): Go, Charm, JSON-RPC for third-party adapters.
- [Issue lifecycle](/features/lifecycle.md): states, transitions, gates, drift.
- [Domain package](/components/domain.md), [markdown store](/components/markdown-store.md), [OKF store](/components/okf-store.md), [CLI](/components/cli.md), [Claude Code integration](/components/claude-code.md).
- [Storage format](/conventions/storage-format.md): files, frontmatter, canonical form.
- [Knowledge base conventions](/conventions/knowledge-base.md): how entries are written.
- [Out of scope](/decisions/out-of-scope.md): what prep deliberately does not do.

The original design is frozen in `DESIGN.md` as a historical snapshot; open work lives in `.prep/issues`.
