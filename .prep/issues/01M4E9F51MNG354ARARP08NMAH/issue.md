---
title: 'TUI: tiny terminals, wheel follow, empty knowledge list, screen-aware keys'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

Found in the 0.2.0 codebase audit.

- The Agent screen panicked ("negative Repeat count") in a terminal of four rows or fewer.
- After the wheel scrolled a list, a tab click, a relation click or a jump moved the selection while the view stayed put, so the selection could be off screen.
- A knowledge filter or `a` that left the list empty still showed the old entry, and `o` listed its issues.
- `t` toggled the issue tree from any screen and `y` copied the issue ID on the knowledge and Agent screens; 1–9 switched the issue tab without leaving another screen, unlike a tab click.
- The filter footer listed the issue flags on the knowledge screen and left out --under, --parent, --leaf and --top; the keymaps of the knowledge, Agent and check screens missed keys that work there.

## Open questions
