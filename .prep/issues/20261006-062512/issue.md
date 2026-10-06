---
title: Improve the TUI editing experience
kind: code
parent: 20261005-152616
tags:
  - tui
---

Editing issues in the TUI works but is not yet pleasant to use: dialogs replace the screen instead of appearing over it, the action menu is long, and text fields are single-line inputs or a round trip through $EDITOR. Editing by hand is not a priority while agents do most writing; this issue collects the improvements for when it becomes one.

## Open questions

- Which editing flows matter most in practice (creating issues, writing requirements, checking criteria, transitions)?
- Inline multi-line editing in the TUI, or keep $EDITOR as the main path?
- Should dialogs overlay the current screen instead of replacing it?
- Is the action menu the right entry point, or do frequent actions need direct keys?
