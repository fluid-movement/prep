---
title: 'Filter bar: suggest the next flag after a finished value'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - views
---

After a finished value and a space (`--state done `), the filter bar suggests the flags not used yet, so a query can be built with enter and arrows alone, without typing `--`. Flags already in the query are left out, except those that can repeat (`--under`, `--text`), which come last; plain words still search titles and get no suggestions.

## Open questions
