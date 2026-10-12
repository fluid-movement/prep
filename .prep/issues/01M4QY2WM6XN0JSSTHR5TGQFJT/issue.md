---
title: Findings from the first real runs of the agent behavior suite
kind: manual
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - testing
  - adoption
priority: high
---

The agent behavior suite (01M4QKTQVR3J6PJZAJM0MVWQR0) ran for real for the first time on 2026-10-12, on haiku and sonnet. It found where agents do not use prep the way the skill and guide intend, and a few rough edges of the CLI. This parent collects what is still open; each child states the evidence, the options and how to check a fix with the suite. Fixed on the way and not tracked here: `prep list --max N` and `--tags` (agents guessed both), `prep guide` naming `.prep` record files instead of `prep show`, the `project:id` hint when a dependency is not found, and three faults in the eval cases and the transcript parser (heredoc pipelines).

## Open questions
