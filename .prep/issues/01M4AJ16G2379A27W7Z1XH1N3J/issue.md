---
title: TUI filter bar offers flag values
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
depends_on:
  - 01M4AHPY9MSW3EVBPV5BEJHFQ1
tags:
  - tui
  - views
---

While typing a query in the TUI filter bar, the user sees the values each query flag accepts, from the same source as `prep flags`: after `--state ` the states, after `--tag ` the tags in use, after `--under ` the parents with their titles, and so on. The candidates appear in a picker list under the bar that narrows as the user types; choosing one (keys or click) inserts it.

## Open questions
