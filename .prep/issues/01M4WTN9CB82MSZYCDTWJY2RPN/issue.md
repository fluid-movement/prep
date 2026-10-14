---
title: Codex attribution environment prefixes break approved command rules
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - integration
  - codex
  - hooks
priority: critical
---

The Codex session-start hook instructs agents to set PREP_ACTOR and PREP_SESSION for every shell command. Agents consequently prefix unrelated commands such as go test with inline environment assignments. In the affected Codex permission policy, commands containing environment assignments are not evaluated against approved command prefixes, so an existing go approval no longer applies and users receive unnecessary approval requests. Narrow attribution guidance to prep CLI invocations so ordinary commands retain their approved command shape, while prep activity still records the correct actor and session. Update the shared hook, embedded and packaged skill guidance, relevant documentation and regression coverage consistently. Verify that session guidance no longer requests attribution prefixes on unrelated commands, and that prep invocations retain correct attribution. Do not broaden command permissions or disable attribution to solve this. Scope is the prep integration guidance; implementation remains separate from this issue creation.

## Open questions
