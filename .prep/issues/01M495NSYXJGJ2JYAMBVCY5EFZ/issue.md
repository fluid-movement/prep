---
title: 'TUI tabs: open work first'
kind: code
parent: 01M48FAKAG2X998F2ZHE1MHVTC
tags:
  - tui
---

The TUI's tabs are the project's saved views (`views:` in `.prep/config.yaml`); the defaults that `prep init` writes are Attention (`--stale`), Actionable (`--actionable`), In progress, To enrich (defined), To define (open) and All. The user spends most of the time in the list of all unresolved work and has to switch between three state tabs to see it, while the state glyphs on each row already tell those states apart. Ready issues have no tab of their own: they appear only under Actionable (when unblocked and not a parent) or under All.

Rework the default views so the first tab is all unresolved work (open, defined, ready and in progress together; the state glyphs show each row's state), followed only by tabs that earn their place. Two tabs, unresolved work and All, is an acceptable outcome. This repository's `config.yaml` is converted to the new defaults once; no migration for other projects.

## Open questions

- Is Attention (`--stale`: a defined, ready or in-progress issue whose requirement or kind changed since its baseline, so it needs `prep define` again or `prep ack`) worth a tab, or is the stale note on the row in the first tab enough?
- Is Actionable (`--actionable`: ready, not stale, no unresolved dependencies, not a parent; what `prep next` lists for an agent) worth a tab for the human, or is it agent vocabulary?
- What is the first tab called? "Open" collides with the state open (not yet defined); alternatives are "Unresolved", "Active", "Work".
- Does All keep done and dropped issues, or does it become "Resolved" (done and dropped only)?
- Should the query language gain a shorthand for unresolved work (for example `--resolved=false`), instead of the view spelling out `--state open,defined,ready,in_progress`?
