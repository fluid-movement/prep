## D1: Infer the current issue from the agent's prep commands
date: 2026-10-06

Decided with the user: the panel follows prep writes and prep guide naming an issue, ignores reads, falls back to a claimed issue, and /prep:focus pins it. Alternatives: an explicit prep focus command (ceremony agents forget, and needs per-session storage the CLI cannot see); claim only (blind during define and enrich).

## D2: prep watch as the update channel
date: 2026-10-06

A streaming prep watch command reuses the TUI watcher, updates the panel within a moment of any change from any process, and serves future harness integrations. Alternative: the mod polls prep on a timer (lag, constant work).
