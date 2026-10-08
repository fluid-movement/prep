---
title: 'Screen bar: tier-one navigation between Issues, Agent and Knowledge'
kind: code
parent: 01M4EA4RB500QF2449Y3914VMQ
tags:
  - tui
---

A bar at the top of every screen names the screens (Issues, Agent, Knowledge), marks the current one and switches on a click or the screen's key (`i`, `w`, `b`). The issue lists move to a second line below it as tier two, so the screen bar is never mistaken for a list. Clicking a list from another screen goes to Issues on that list.

- The Agent entry shows whether the agent is active: a live dot while active, muted when idle.
- At the bar's right: the ✕/▲ counts, which open Check on a click, and a gear for Settings.
- One back stack records screen changes and jumps; backspace and a clickable `‹ back` crumb in the bar return to the previous screen and selection.

## Open questions
