---
title: 'Filter bar: live filtering while typing'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - views
---

The issue list follows the filter bar as the user types: whenever the text parses as a query, the tab's rows narrow to it right away, so the result is visible before pressing enter. Text that does not parse yet (a flag without its value, an unknown flag) keeps the last rows that matched, without an error; the error appears only on enter.

- enter keeps the filter (as today), esc goes back to the filter the tab had before the bar opened (instead of clearing it), so trying a query is safe.
- The pane title shows the live query while typing.

## Open questions

- Should esc restore the previous filter (as described) or clear it as today, with a separate key to restore?
