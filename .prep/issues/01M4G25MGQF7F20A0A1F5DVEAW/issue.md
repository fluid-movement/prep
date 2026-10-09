---
title: 'Children view: unfiltered, by state, from the detail'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

Following "list all" from the detail focused the current view on the parent, so the view's filter still applied: from Unresolved, a release showed almost none of its children. A list reached from the detail takes no filtering.

- "list all" opens a children view after the configured ones (`↳ <parent title>`): no view flags, the parent's whole subtree, its children by state with open ones first and resolved ones last, each with its subtree.
- Switching views closes it; ← or esc at its top returns to the view it was opened from; backspace goes back as anywhere.

## Open questions
