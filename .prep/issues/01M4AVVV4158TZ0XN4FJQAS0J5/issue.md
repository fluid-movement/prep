---
title: Edit tags in the TUI
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

The TUI shows tags but cannot change them; only `prep edit <id> --tag` can, so removing a tag (for example `release` from a leaf issue so a view shows only release parents) needs the CLI. The edit menu gets a tags entry: a dialog with the issue's tags as text (space or comma separated), offering the tags in use as suggestions like the filter bar, validated like `prep edit` (lowercase, no `not-` prefix). Saving replaces the issue's tags; an empty field removes them all. Works on resolved issues too, which may change their tags.

## Open questions
