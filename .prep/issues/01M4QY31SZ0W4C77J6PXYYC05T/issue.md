---
title: Agents file a new issue before looking for one that already covers it
kind: code
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
tags:
  - cli
  - adoption
priority: high
---

When a user reports a bug or asks for work that an open issue already describes, the agent should find that issue and extend it (or say it exists), not file a duplicate. In the suite haiku filed a duplicate in 3 of 3 runs of `real-bug-report-to-issue` (two of them never searched at all: `prep new` first, `prep show` afterwards if ever), and sonnet in 1 of 2. The fix should not depend on the model remembering to search: make `prep new` itself surface similar open issues.

For 0.3.0, verify the CLI safeguard deterministically and with a targeted Codex duplicate-report fixture. Original named-model thresholds are deferred to 01M4J8789E81QAJHG8R8ZPEMRD.

## Open questions
