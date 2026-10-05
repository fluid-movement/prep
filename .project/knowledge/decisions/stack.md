---
type: decision
title: Stack
description: Go single binary; Charm (Bubble Tea, Bubbles, Lip Gloss, Glamour) for the TUI; JSON-RPC over stdio for third-party storage adapters.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Stack

Applies when adding dependencies or new interfaces.

- **Go**, one static binary holding core, CLI and (later) TUI: near-instant startup for frequent agent calls, simple install. The only dependency so far is `gopkg.in/yaml.v3`.
- **TUI** (not built yet): Bubble Tea app structure, Bubbles components, Lip Gloss layout, Glamour markdown rendering.
- **Third-party storage adapters**: JSON-RPC over stdio, like LSP and MCP, so the core stays a closed, tested binary.
- **Harness adapters**: thin TypeScript only where a harness requires it (Claude Code mod, Pi extension).

Alternatives considered: TypeScript (shares code with harness mods, but needs a runtime and starts slower; OpenTUI not production-ready); Rust (no decisive advantage, slower iteration).
