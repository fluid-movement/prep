---
title: 'Briefing: find instead of grep, smaller models for surveys'
kind: code
tags:
  - tokens
  - knowledge
---

In a fathom audit (session 63e30b20) the briefing hint reached the agent and the subagents read knowledge through `prep knowledge show`, but they still searched entries with `grep -r .prep/knowledge`, and every subagent ran on the parent's large model: that advice is only in the skill, which the agent never invoked.

- The `prep prime` hint says `prep knowledge find` answers where something is documented, instead of grep.
- The hint says to give wide surveys to subagents on a smaller model, telling them the knowledge-read rule.
- The skill's knowledge-read rule says the same about find and grep.
- The hint stays short: it is paid on every request of every session.

## Open questions
