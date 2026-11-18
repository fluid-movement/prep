---
title: 'prep handoff: hand ready issues to cloud sessions'
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
depends_on:
  - 01M7Q2AQV2EXF21424KDFS7CNY
tags:
  - cloud
---

One command hands ready issues to cloud sessions: `prep handoff <id>...` (name open; hand-off exists as `prep brief` and `prep handback` today).

For each issue it checks the hand-off preconditions that `prep brief` already enforces (ready, not stale, dependencies done), checks that the records and the branch are pushed because the cloud session clones from GitHub, and starts one `claude --cloud` session per issue with the standard prompt, or with the `prep brief` text. It prints the session link per issue. It starts nothing when a check fails and names what is missing.

It must not push or start sessions without the user asking for it, and it must report usage cost before starting several sessions. Which environment is used comes from `/remote-env`, not from prep.

Depends on the spike: its findings decide the branch handling, the prompt and how results return.

## Open questions
