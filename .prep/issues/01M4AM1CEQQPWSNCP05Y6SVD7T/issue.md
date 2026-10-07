---
title: Reduce the TUI palette to eight tokens
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - theming
---

A theme is a small palette: a gray ramp, one accent, a selection background and the status colors. The tokens are `text`, `muted`, `subtle`, `accent`, `selection`, `success`, `warning` and `error`; `border` merges into `subtle`, and the per-state and per-kind tokens go away, since state badges have their own glyphs and kinds are written out.

- Kinds render in `muted`.
- States map onto the palette: open `muted`, defined `text`, ready `success`, in progress `accent`, done `subtle`, dropped `subtle` (struck through as today).
- Every built-in theme defines the eight tokens for dark and light; custom themes and `prep theme new` use the same eight. `state.*`, `kind.*` and `border` are unknown tokens; no migration, prep has no outside users.
- The default theme intentionally looks calmer than before: screen and gallery snapshots change once.

## Open questions
