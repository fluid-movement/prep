---
title: Claude Code hooks make sure prep calls are made
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
depends_on:
  - 01M4J8780KZRKWHJR4NN8AD85M
  - 01M4SX94VFDXCRYWNKCJG1H01J
tags:
  - integration
  - adoption
priority: high
---

prep now reaches the person through the TUI rather than harness-specific panes, so the one harness-specific piece left is hooks that make sure the agent actually uses prep where it should. In Claude Code, add the hooks the ledger analysis (01M4J8780KZRKWHJR4NN8AD85M) calls for:

- a PreToolUse hook that catches Read, Grep, Glob, Edit and Write calls and shell reads (`cat`, `sed`, `head`, `grep`, `ls`) on `.prep/issues` and `.prep/knowledge`, and points the agent to the prep command that does the job (`prep show`, `prep knowledge find|show`, the write commands);
- making sure the skill is in context when an issue is worked on, without the person typing `/prep`: in 12 working sessions the skill was loaded 11 times, nearly always because the person typed it, and the one session without it had the most bypasses.

Iterate experimentally: change a hook, run real sessions, and measure in the activity stream how often prep is used where expected, before and after. A hook never fails or blocks a call because of its own error.

## Open questions

- Should a hook block the misdirected call with a pointer, or let it through and add a reminder to the context?
- Does the Jev research change how a hook decides?
- Agents run an issue's whole lifecycle in a minute without `prep guide` (3 of 24 completions in the stream were guided), mostly for issues they filed themselves from the conversation. Is that fine, or should a hook (or the transition output itself) bring the guide's knowledge candidates and gates in at claim?
