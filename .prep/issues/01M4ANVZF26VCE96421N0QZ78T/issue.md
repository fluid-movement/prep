---
title: Enter in the filter bar takes the highlighted suggestion
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - views
---

While the filter bar shows suggestions, enter inserts the highlighted one, as tab does; with no suggestions showing, enter applies the filter. Typing `--sta` and pressing enter gives `--state ` instead of the error "unknown query flag --sta".

- A word that already equals a candidate (a value just picked, such as `open`) shows no suggestions, so the next enter applies the filter.
- An error from a previous enter disappears as soon as the input changes.
- The key help names enter as "pick or apply".

## Open questions
