---
title: Select linked issues directly in the TUI detail
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

The detail pane shows an issue's links by shape (breadcrumb, children tree, needs and unblocks arrows), but they are display only: following one takes o and its number from a Go to menu, and nothing in the detail itself can be selected. Tab used to move a selection over the links, but it meant something else in the list, so it now always switches views (01M47Y6EA0SMCZAKG03785YF76, D9). This issue brings back selecting a link where it is drawn, without giving a key two meanings depending on focus.

## Open questions

- How is a link selected in place: arrows or j/k while the detail has focus (then how does the document scroll), a link mode entered with a letter, or the mouse (01M4863WN8D1HHTP302Q8G2PQX)?
- Should the selection stay visible when the detail is not focused?
- Does o <n> stay as the keyboard path next to it?
