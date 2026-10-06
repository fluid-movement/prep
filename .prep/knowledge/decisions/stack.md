---
type: decision
title: Stack
description: Go single binary; Charm v2 (Bubble Tea, Bubbles, Lip Gloss, Glamour under charm.land) for the TUI; JSON-RPC over stdio for third-party storage adapters.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Stack

Applies when adding dependencies or new interfaces.

- **Go**, one static binary holding core, CLI and TUI: near-instant startup for frequent agent calls, simple install. Dependencies: `gopkg.in/yaml.v3`, the Charm libraries below and `fsnotify`. The TUI libraries grew the binary from about 3 MB to about 17 MB stripped (23 MB with symbols, on Charm v2); CLI startup stays around 10–20 ms.

- **TUI**: Charm v2 under `charm.land` (Bubble Tea, Bubbles, Lip Gloss and Glamour, all `/v2`, kept at their latest releases; 20261006-130742 moved from v1): Bubble Tea app structure with declarative views (alt screen, mouse mode), Bubbles components, Lip Gloss layout and layers, Glamour markdown rendering. Pin major versions here when adding a Charm library; v1 import paths (`github.com/charmbracelet/...`) are not used. Looks come from the [
TUI design system](/components/tui-design-system.md).
- **Third-party storage adapters**: JSON-RPC over stdio, like LSP and MCP, so the core stays a closed, tested binary.
- **Harness adapters**: thin TypeScript only where a harness requires it (Claude Code mod, Pi extension).

Alternatives considered: TypeScript (shares code with harness mods, but needs a runtime and starts slower; OpenTUI not production-ready); Rust (no decisive advantage, slower iteration).
