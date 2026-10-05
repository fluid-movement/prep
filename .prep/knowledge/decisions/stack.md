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

- **Go**, one static binary holding core, CLI and TUI: near-instant startup for frequent agent calls, simple install. Dependencies: `gopkg.in/yaml.v3`, the Charm libraries below and `fsnotify`. The TUI libraries grew the binary from about 3 MB to about 15 MB; CLI startup stays around 10–20 ms.
- **TUI**: Bubble Tea app structure, Bubbles components, Lip Gloss layout, Glamour markdown rendering. Glamour v1.0.0 pins Lip Gloss to a v1.1.1 pseudo-version. Looks come from the [TUI design system](/components/tui-design-system.md).
- **Third-party storage adapters**: JSON-RPC over stdio, like LSP and MCP, so the core stays a closed, tested binary.
- **Harness adapters**: thin TypeScript only where a harness requires it (Claude Code mod, Pi extension).

Alternatives considered: TypeScript (shares code with harness mods, but needs a runtime and starts slower; OpenTUI not production-ready); Rust (no decisive advantage, slower iteration).
