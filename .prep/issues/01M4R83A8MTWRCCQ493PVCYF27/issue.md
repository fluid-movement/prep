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

## Open questions

- At which widths do the columns disappear, and all at once or one by one (ID first, then state text, then kind)? In the screenshot the list is about 65 columns wide and loses most of the title; the TUI already has a "short layout" and a single-pane mode below 92 columns (`ui.Split`), so the thresholds should be chosen with those in mind.
- Should `z` (the existing toggle between the adaptive split and one pane) or a new key override the automatic hiding, for people who want the ID even in a narrow pane?
