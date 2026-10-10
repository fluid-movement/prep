---
title: Release 0.3.0
kind: manual
depends_on:
  - 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - release
---

The release after 0.2.0: prep calls in Claude Code that happen when they should. prep reaches the person through the TUI rather than harness-specific panes, so the only harness-specific work left is hooks that make sure the agent calls prep. This release analyses real sessions in the ledger to find where prep was not used when expected and why, researches whether Jev can judge when a call should be a prep call, and iterates on the Claude Code hooks until the ledger shows prep used where expected. Its children are the work that must ship with it.

## Open questions
