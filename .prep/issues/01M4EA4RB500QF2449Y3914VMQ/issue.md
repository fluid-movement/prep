---
title: 'TUI two-tier navigation: screens on top, views below'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

The TUI gets one navigation tree, visible and clickable on every screen, so a person always sees where they are and can get anywhere with the mouse as well as with keys.

- Tier one, the screens: Issues, Agent, Knowledge. Always shown at the top, the current one marked, each a click target and a key: `i` Issues, `w` Agent, `b` Knowledge. Check and Settings are small items at the right end of the bar: the ✕/▲ counts open Check, a gear opens Settings.
- Tier two, the views of the current screen: the configured issue lists on Issues; Issue, Activity and Usage on Agent. Shown under tier one, clickable, switched with tab/shift+tab and 1–9. Knowledge has no views yet, but uses the same mechanism so views can be added later (01M4EA4RZ6VDWCR6XR6DCGAA8T).
- Back across screens: backspace and a clickable back crumb return to where the person came from, screen and selection, whether they moved by key or by click (Agent → issue detail → back to Agent).

Why: today the issue lists are the only tabs, Knowledge, Agent, Check and Settings are reachable only by keys, and clicking a list from the Agent screen leaves no click path back. A clear hierarchy also lets the TUI work in a short terminal, so a person can stack the agent and the TUI on top of each other instead of side by side.

The children carry the parts: the screen bar with the back stack, the Agent views, and short terminals.

## Open questions
