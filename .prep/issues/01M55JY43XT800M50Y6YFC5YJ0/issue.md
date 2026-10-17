---
title: 'Act on eval run A: record the user''s decisions, look up decisions before coding'
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - eval
---

Eval run 20261017-183812-135e showed two real gaps and three obsolete cases: agents record their own small choices instead of the decision the user states, and do not look up recorded decisions before changing code; implement-complete and log-and-stop no longer fit (agents correctly refuse unverified work and stop on context conflicts), and fix-matching-issue still expects a descriptive entry edit. Change the two rules and the finish hook wording, delete the two obsolete cases and update the rubric, then rerun the affected cases with the user's OK.

## Open questions
