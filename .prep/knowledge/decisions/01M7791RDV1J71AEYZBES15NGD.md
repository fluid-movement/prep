---
type: decision
title: Native request coverage is full when no recorded gaps exist
description: Keep native agent-scoped request accounting and durable bridge gaps; report full coverage for healthy observed sessions.
date: "2026-11-12"
issue: 01M75HAMNFR4SDWSNHX6585DSY
supersedes:
  - 01M75HSAMADHX34RQX9QPZ1P4X
---

Use Claude turn.step and Codex PostModelRequest only, never transcript totals or a SubagentStop cost hook. Preserve agent attribution, optional request identity, Codex cache-token splitting, prompt sends without retry loops, durable bridge-disable markers, and separate turn summaries. Per the user clarification, coverage is full when request usage is present for all observed sessions and no gap markers or requests missing usage exist; those missing observations make coverage partial, while an empty stream remains unknown. Other event kinds, including summaries and hooks, do not themselves reduce coverage. Full describes the recorded stream, not proof that every provider attempt was observed.
