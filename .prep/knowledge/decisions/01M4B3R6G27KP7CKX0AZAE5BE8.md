---
type: decision
title: "Palette lives in its own package"
description: "Palette lives in its own package"
date: "2026-10-07"
issue: 01M46QWSKGQZ5S0N3RHTEE86Y1
---

Token names, built-in themes and theme resolution go into `internal/palette`, which imports nothing from the TUI. The theme package builds Lip Gloss colors and styles from a palette; mdstore validates config with it; the CLI prints and writes themes with it.

Rationale: `prep check` and `prep theme new` must know the tokens without pulling the TUI into the store or the CLI's non-TUI paths, and domain must stay free of UI packages.

Alternatives: keep everything in `internal/tui/theme` (mdstore would import a TUI package); put tokens in domain (UI vocabulary in the domain model).
