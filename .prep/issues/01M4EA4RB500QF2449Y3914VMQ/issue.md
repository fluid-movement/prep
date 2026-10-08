---
title: 'TUI two-tier navigation: screens on top, views below'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

The TUI gets one navigation tree, visible and clickable on every screen, so a person always sees where they are and can get anywhere with the mouse as well as with keys.

- Tier one, the screens: Issues, Agent, Knowledge. Always shown at the top, the current one marked, each a click target and a key.
- Tier two, the views of the current screen: the configured issue lists on Issues; Issue, Activity and Usage on Agent; tabs on Knowledge, if it gets any. Shown under tier one, clickable, switched with tab/shift+tab and 1–9 as the issue lists are now.

Why: today the issue lists are the only tabs, Knowledge, Agent, Check and Settings are reachable only by keys, and clicking a list from the Agent screen leaves no click path back. A clear hierarchy also lets the TUI work in a short terminal, so a person can stack the agent and the TUI on top of each other instead of side by side.

The children carry the parts: the screen bar, the Agent tabs, short terminals and the knowledge tabs.

## Open questions
- Keys for tier one. Proposed: q w e r by position. Today q quits, w opens Agent, e opens the edit menu and r reloads, and TUI keys are mnemonic letters by convention, not positions. Options: (a) q w e r for the screens, quit moves to Q or ctrl+c, the edit menu and reload get new letters; (b) mnemonic letters that are free (i issues, w agent, b knowledge as now); (c) a leader key, such as g then i/w/k.
- Where Check and Settings live: a fourth and fifth screen in tier one, small items at the right end of the bar (next to the ✕/▲ counts, which would open Check), or views of another screen.
- Does Knowledge get tier-two tabs, and which: by type (components, features, decisions, conventions), attention only, or none for now?
- Should backspace and a clickable back crumb go back across screens (Agent → issue detail → back to Agent), or only within one?
