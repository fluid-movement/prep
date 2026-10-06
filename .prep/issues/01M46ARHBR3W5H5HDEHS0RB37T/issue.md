---
title: TUI knowledge view
kind: code
parent: 01M46AREE07YNNVPBA8F4MT56D
depends_on:
  - 01M46ARFD8BQNT8QBJYRB30C35
tags:
  - tui
  - knowledge
---

Knowledge is written by agents for agents. The TUI shows it read-only, so a person can see what the agents know, where it came from and whether it still holds, without opening the files.

- Issue detail lists the knowledge entries in play: the ones linked from the issue's context and, for finished issues, the ones its resolution changed. Each is a link like the other relations and opens the entry.
- A knowledge view, reached like the check and settings screens, lists the entries by title, description, type and status, filterable like issues (by type, status, scope and words in the title), and shows the selected entry rendered in the detail pane with its metadata and the issues that changed it (computed backlinks).
- Entries that need an agent's attention are marked and can be listed on their own: drifted (their scoped code changed since they were confirmed), broken links and entries over the size limit, as prep check reports them.
- Nothing in the TUI edits knowledge.

## Open questions
