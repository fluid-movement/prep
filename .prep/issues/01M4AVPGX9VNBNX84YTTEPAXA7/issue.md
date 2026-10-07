---
title: Link mode swallows keys beyond the last link
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
---

In the detail's link mode, every letter in the link key list (b, c, d, f, g, m, n, p, r, s, t, …) counts as a link key. When the issue has fewer links than that key's position, the key does nothing: `s`, `f` or `n` neither follow a link nor do what they do outside link mode. A key without a link leaves link mode and acts as usual.

## Open questions
