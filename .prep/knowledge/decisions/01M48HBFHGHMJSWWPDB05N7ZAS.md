---
type: decision
title: "Hit-test with geometry recorded while rendering, not bubblezone"
description: "Hit-test with geometry recorded while rendering, not bubblezone"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
---

The model records where tabs, rows, relation lines and dialog lines are drawn while `View` renders, and `mouse.go` maps click positions to those records.

Alternatives considered:
- bubblezone (lrstanley/bubblezone) marks regions with zero-width escape sequences and scans the final output. `ui.Overlay` strips and truncates background lines with `ansi.Strip`/`ansi.Truncate`, so markers behind a dialog get cut or dropped. It also adds a dependency and a `zone.Scan` wrapper around every `View`, which the goldens would have to go through.
- Recomputing the geometry in `Update` from widths and heights would duplicate the layout math from `View`, and the two copies would drift apart as screens change.

Recording during rendering is exact by construction and keeps the TUI dependency-free for this feature. It is also testable because `View` runs in the existing tests.
