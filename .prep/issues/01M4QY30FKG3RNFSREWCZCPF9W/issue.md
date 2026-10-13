---
title: Agents skip the tracker for tasks that do not mention prep
kind: code
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
depends_on:
  - 01M4QY2XNA86D2MSFZXDT982K6
  - 01M4SX94VFDXCRYWNKCJG1H01J
tags:
  - adoption
  - integration
priority: high
---

When a user asks for work in plain words and a tracked issue covers it, the agent should find that issue, claim it, log progress and complete it with the knowledge entries in step (the outcome chosen in 01M4QY2XNA86D2MSFZXDT982K6). Haiku does not: on `natural-fix-matching-issue` it fixed the code and never looked at the tracker (0/3 runs), leaving the issue ready and unclaimed and the knowledge entry about the changed files drifting (`prep check` K005). Sonnet did the full loop. Make the plugin produce the behavior on haiku too, measured with the suite.

## Open questions
