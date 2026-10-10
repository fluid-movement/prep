---
title: Claude Code hooks make sure prep calls are made
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
depends_on:
  - 01M4J8780KZRKWHJR4NN8AD85M
tags:
  - integration
  - adoption
priority: high
---

prep now reaches the person through the TUI rather than harness-specific panes, so the one harness-specific piece left is hooks that make sure the agent actually uses prep where it should. In Claude Code, add the hooks the ledger analysis calls for: for example a PreToolUse hook that catches reads, greps and edits under `.prep` and points the agent to the prep command that does the job, and briefing or skill-loading hooks that make sure the skill is in context when an issue is worked on.

Iterate experimentally: change a hook, run real sessions, and measure in the ledger how often prep is used where expected, before and after. A hook never fails or blocks a call because of its own error.

## Open questions

- Which misses from the ledger analysis are worth a hook, and which are better fixed in the briefing or skill wording?
- Should a hook block the misdirected call with a pointer, or let it through and add a reminder to the context?
- Does the Jev research change how a hook decides?
