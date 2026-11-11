---
title: Reading an issue's state takes one line, not a full show
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
tags:
  - usage
priority: low
---

An orchestrator can read the state of one or several issues in one short line each, without the full `prep show`.

Why: in the sessions analysed on 2026-11-10 main sessions ran `prep show` 70 times without a guide, usually cut with `head` or `grep`, only to see an issue's state (finding 01M72A9TBJ5JQXT2VGT15N2HW1). That is more than the 48 shows the guide change removes (01M6V6R1H9P8F53ESDJTB9VSE9), and each prints the whole issue into the orchestrator's large context unless the agent remembers to cut it. With hand-off by brief the orchestrator checks state more often, not less.

## Open questions

- Is a new form needed, or does `prep list --under <parent>` already give this and only the rules have to name it? Check what the 70 shows were looking for (state, criteria progress, the last log line).
- If a new form: `prep list <id>...` accepting IDs, or `prep show <id>... --state`?
- Does `prep handback` printing the resulting state make most of these reads unnecessary? Measure again with the activity report after the hand-off changes before building anything.
