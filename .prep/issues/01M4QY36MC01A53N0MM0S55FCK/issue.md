---
title: Agents leave knowledge drifting after changing the code it describes
kind: code
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
depends_on:
  - 01M4QY2YM2T9X68RHQW6HTCH53
tags:
  - knowledge
  - adoption
---

After an agent changes files that a knowledge entry is scoped to, `prep check` reports K005 (scoped paths changed since the entry was confirmed). The agent should re-check the entry and update or confirm it (skill rule 10). Sonnet did; haiku left the warning in 3 of 3 runs of `natural-fix-matching-issue`. Make the drift visible to the agent at the moment it matters, without relying on it remembering rule 10.

## Open questions
