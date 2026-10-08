---
title: 'Key actions: one meaning per key across screens and menus'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

After `i` (01M4EBKFBPPEWBM98CWSWM92MT), the keymap still gives some keys different meanings by screen, and the action menu repeats the edit menu. Keys are mnemonic letter sequences by convention; a key should mean the same thing wherever it works, or at least the same kind of thing.

Found in the current keymaps:

- `a`: the action menu on Issues, the next agent on Agent, the attention filter on Knowledge; inside the action menu `a` is Acknowledge.
- `t`: tree or flat on Issues, the next theme on Settings, Edit title in the action and edit menus.
- `o`: an issue's links on Issues, the issues that changed an entry on Knowledge (the same idea, links; probably fine).
- `1–9`: views on every screen, but rows on Settings.
- The action menu (`a`) repeats the edit menu (`e`: requirement, context, title, tags) and the priority picker (`p`), so the same action has two sequences (`a t` and `e t`, `a p` and `p`).
- esc means "leave the screen", "leave the detail", "clear the filter" or "cancel" depending on where; backspace goes back across screens.
- The footers differ in how they name the screen keys (`i b`, `i w`, `i w b`, `esc w`).

Outcome: a key table for every screen and dialog in which each key has one meaning, the action menu holding only what no other key does (the lifecycle transitions), and the keymap (`?`) and footers generated from it consistently.

## Open questions
- Which key for the next agent on Agent, and for the attention filter on Knowledge, once `a` means the action menu only? Or is `a` acceptable as "the screen's main action"?
- Should the action menu keep the edit entries for discoverability, listed with their direct keys, or drop them?
- Settings digits: keep rows (view n is row n) or make them follow the second-tier rule like everywhere else?
