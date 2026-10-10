---
title: 'Ledger analysis: is prep used in Claude Code sessions when we expect it'
kind: research
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - adoption
  - activity
priority: high
---

The activity stream (the ledger: `prep activity`, fed by the Claude Code activity bridge and by prep itself) records every tool call and prep command of past sessions. Analyse those sessions to find where an agent should have called prep and did not: reading `.prep/knowledge` files or grepping them instead of `prep knowledge find|show`, editing `.prep` records by hand, starting work on an issue without `prep guide`, finishing without `prep complete`, or skipping `prep prime`/the skill altogether.

For each kind of miss, work out why: the skill was never loaded, the briefing did not mention the call, the wording did not trigger, the call was costlier than the alternative, or the agent did not know the command existed. Findings list the misses with their frequency and cause, and propose the hooks, briefing or skill changes that would turn each into a prep call; those proposals become the code issues of this release.

## Open questions
