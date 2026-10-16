---
title: 'Rethink: agent guidance (skill, hooks, guide, prompt context)'
kind: decision
parent: 01M4ZCK8C17RDGBVCTMQPGMEK7
tags:
  - rethink
---

Scope: the prep skill, session start, prompt, finish and tool hooks, prep prime and prep guide output, record redirects, tracking policy. Starting observations: the implement-step guide grew from 2.3 KB to 5.1 KB; the prompt hook adds text to every turn; plugin A/B on what the skill and the hook each buy is still open. What is the smallest guidance that keeps agents on the workflow?

## Open questions

- The mechanism for handing off a ready issue to a subagent or another harness (Codex): how the orchestrator starts it, what it passes besides prep guide, how the result comes back for checking. Raised by the user 2026-10-16; not settled.
