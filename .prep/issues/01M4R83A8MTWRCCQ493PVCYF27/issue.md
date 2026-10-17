---
title: 'TUI list: hide the ID, the state text and the kind columns when the pane is narrow'
kind: code
tags:
  - tui
---

In a narrow pane the issue list spends about 30 of its columns on the ID, the state label and the kind, and the title is cut off ("Agent behavior test …", "Findings from the fir…"). Hide those three columns when the list is narrow, so the title gets the space:

- the ID (the last six characters, `PWJH2F`) goes first;
- the state text goes next, but the state glyph (`○ ● ◐`) stays, so the state is still readable at a glance;
- the kind (`code`, `manual`, `research`, `decision`) goes too.

The priority mark, the progress bar, `blocked`/`stale` notes and tags stay: they carry information the title does not. The selected row, mouse hit-testing and the detail pane must keep working with fewer columns (the ID is what a click or `enter` resolves, internally, not what is drawn).

Settled on 2026-10-17: no fixed widths. A row drops the columns one by one, ID first, then the state label, then the kind, until the title has at least 30 cells or all three are gone; the state glyph always stays. No key overrides it: the ID stays visible in the detail pane.

## Open questions
