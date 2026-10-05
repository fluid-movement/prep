---
type: decision
title: Design principles
description: The nine principles prep is built on; check changes against them.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Design principles

Applies to every change. Decided in the original design (`DESIGN.md`).

1. **The system enforces steps, not how work happens.** It defines what each step must produce and the gates between steps.
2. **Write permissions belong to the harness.** No approval mechanism of prep's own; reads and writes are separate commands so harness rules can tell them apart.
3. **It is a small database.** Store every fact that cannot be derived; never store anything that can. State, progress, blocked and actionable are computed.
4. **Every transition leaves a record** in the file that represents it; no fact depends on git history.
5. **Files never move.** An issue lives at `.project/issues/<id>/` forever.
6. **Storage and presentation are decoupled.** Files are storage, git is versioning, CLI and TUI are query interfaces.
7. **Domain and storage are separate layers.** Markdown is the first storage adapter, see [Architecture](/components/architecture.md).
8. **Keep it simple.** Scan all files on every run; no index or cache until proven necessary.
9. **Issues record change, the knowledge base records state.** See [Knowledge base conventions](/conventions/knowledge-base.md).
