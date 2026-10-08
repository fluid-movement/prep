---
title: 'token-ledger: attribute issues per session and classify git show reads'
kind: code
tags:
  - tokens
---

The ledger's issue attribution and area classification misreport a measured session (fathom, 704a70a0), which a before/after comparison needs to get right.

- The issue being worked on resets when the session ID changes. A new session ID without `session.start` carried the previous session's issue over: 287 of 303 rows were attributed to an issue the session never named.
- `prep new` with `-h`/`--help` does not move to an issue (here the ID came from a `prep show` in the same command). The panel's inference has the same rule.
- `git show <rev>:<path>` reads are classified by their path like `cat` reads (about 75k tokens of docs read this way had no area).

## Open questions
