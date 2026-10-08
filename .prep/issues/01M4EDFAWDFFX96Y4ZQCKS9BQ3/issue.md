---
title: 'Spacing scale in the TUI design system: tokens, an outer margin, a roomier header'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - design
---

The TUI feels cramped, the header navigation most. Spacing is ad hoc: `theme` has three spacing tokens (`Gap`, `Pad`, `Section`, all 1, `Gap` unused), components use literal paddings, about 40 places write runs of spaces by hand, and ten places compute a pane's inner width as `w - 2 - 2*theme.Pad`.

Terminals have no pixels: Bubble Tea and Lip Gloss lay out in character cells, and the font size is the terminal's. A cell is about twice as tall as it is wide, so one blank row reads like two columns of space. The scale is therefore in cells, with horizontal and vertical steps paired.

Proposed:

- A small scale in `internal/tui/theme`, used everywhere instead of literal numbers and runs of spaces, for example `XS` = 1 column / 0 rows, `S` = 2 columns / 1 row, `M` = 4 columns / 1 row, `L` = 6 columns / 2 rows; components take a size, not a number.
- An outer margin around the whole TUI (horizontal `S`; vertical only when the terminal has rows to spare, never in the short layout).
- A roomier header: more space between `prep`, the screens and the right-end items, a blank row between the screens bar and the views when there is room, and a row between the views and the body.
- A helper for a pane's inner width and offsets, replacing the repeated arithmetic and `paneInner`.
- The gallery shows the scale; goldens at 80×24 and 110×28 still fit (80×24 must stay usable after the margins).

## Open questions
- How many sizes: three (S, M, L) or four (XS, S, M, L)?
- Outer margin: horizontal only, or also a top and bottom row on tall terminals?
