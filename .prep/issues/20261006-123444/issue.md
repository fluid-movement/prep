---
title: Select linked issues directly in the TUI detail
kind: code
parent: 20261006-123443
tags:
  - tui
---

The detail pane shows an issue's links by shape (breadcrumb, children tree, needs and unblocks arrows), but they are display only: following one takes o and its number from a Go to menu, and nothing in the detail itself can be selected. Tab used to move a selection over the links, but it meant something else in the list, so it now always switches views (20261006-062512, D9). This issue brings back selecting a link where it is drawn, without giving a key two meanings depending on focus.

## Open questions

- How is a link selected in place: arrows or j/k while the detail has focus (then how does the document scroll), a link mode entered with a letter, or the mouse (20261006-084337)?
- Should the selection stay visible when the detail is not focused?
- Does o <n> stay as the keyboard path next to it?
