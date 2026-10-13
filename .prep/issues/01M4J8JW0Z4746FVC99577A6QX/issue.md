---
title: prep knowledge list shows each entry's confirmation state
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
tags:
  - knowledge
  - cli
priority: low
---

To find entries that were never confirmed or whose scope changed, agents grep the frontmatter under `.prep/knowledge` (`grep -L '^confirmed_commit'`, `head` of an entry), because `prep knowledge list` prints scope and status but not `confirmed_commit` or whether the entry has drifted. These were the only knowledge-file reads left after `prep knowledge find|show` existed (ledger analysis 01M4J8780KZRKWHJR4NN8AD85M).

`prep knowledge list` (text and JSON) shows the confirmed commit and the drift state, and can filter by them.

## Open questions
