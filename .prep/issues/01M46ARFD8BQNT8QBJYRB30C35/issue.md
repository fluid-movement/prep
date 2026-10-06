---
title: 'TUI issue views: prep tui, tabs, list, detail pane, live reload'
kind: code
parent: 01M46AREE07YNNVPBA8F4MT56D
depends_on:
  - 01M46QA8TRCHXVESZEA603SJMY
tags:
  - tui
---

prep tui opens a terminal UI next to the agent harness, built only from the TUI design system's components. Query-first layout like lazygit or k9s: the saved views from config.yaml are tabs (Attention, Actionable, In progress, To enrich, To define, All), each showing its issues as a list with state, kind, progress or blocked status and title. Selecting an issue shows its records in a detail pane: requirement, open questions, context, decisions, acceptance criteria with their numbers, history and resolution, rendered as markdown. The TUI watches .prep and reloads within a moment when files change, so work an agent does in the harness appears without a keypress. It uses the same query engine and derived states as the CLI and only reads in this issue. Keyboard navigation covers tabs, list and pane focus, scrolling, copying the selected issue ID, and quitting; a key help footer lists the keys.

## Open questions
