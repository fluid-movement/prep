---
title: Free-form tags on issues
kind: code
depends_on:
  - 20261005-152629
tags:
  - cli
  - tui
---

Issues get an optional tags list in issue.md frontmatter with no lifecycle meaning, so kind stays about the workflow. prep new and prep edit set tags, prep list and saved views filter by them with --tag, prep show and the TUI show them, and the TUI filter bar accepts --tag. Tags are free text, lowercased, without spaces.

## Open questions
