---
title: 'TUI: project name in the header and a project switcher'
kind: code
parent: 01M4N5ARVE6J5APZYJ7VA0N464
depends_on:
  - 01M4N5BKPZ1JAPSZC27P1272B4
tags:
  - projects
---

The TUI header shows the project's name where it now says "prep" (still plain text, not the accent). Clicking it, or its mnemonic letter sequence, opens a project switcher over the body: the registry's projects, newest first, with name, remote and last use, a filter as you type, and the current project marked. Choosing one reloads the TUI on that project in place: the watcher, activity stream, screens and the back stack restart for the new root; the previous project is one entry back so backspace returns to it.

- Projects whose `.prep` is gone are not listed (the registry already drops them). Name collisions show the remote to tell them apart.
- `prep tui --project <name>` opens on that project.
- Keys follow the TUI's conventions: a stable mnemonic letter sequence, arrows only as a fallback; the wheel scrolls the list, a click selects.

## Open questions
