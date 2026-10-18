---
type: decision
title: Prep offers handoffs and never runs them; implementer and checker are passes
description: prep guide offers implement and review handoff prompts; solo sessions check in a separate pass and complete
date: "2026-10-18"
issue: 01M56ZKEHAQE5C4G8J0NQA0KPX
scope:
  - internal/domain/guide.go
  - internal/cli/rules.md
supersedes:
  - 01M54VQCGG9212J72216RJ1J8V
---

User decision, 2026-10-18, refining the handoff decision of 01M52ZR8564ECW8GP3ZHSXEM0A. Prep still never runs, schedules or requires other agents; the harness does that. Prep is a helpful guide: prep guide offers the handoff when an agent wants it ("start a subagent or another harness with the prompt: run prep guide <id> and follow it"; for a review: "run prep guide <id> and follow its checker section") and does not insist. Whether to delegate is the orchestrating agent's judgment, worth it for several issues, a cheaper implementer or a long session, and usually not for one small issue.

Implementer and checker are passes, not necessarily different agents. When another agent will check, the implementer logs 'Implemented, ready for check' and stops. In a solo session the same agent implements, then checks in a separate pass against the criteria and the Definition of Done, then completes. Once the latest work log line says ready for check, prep guide gives the checker brief, so a review subagent works from the issue alone, the same way an implementer does; a reviewer that cannot judge from the issue shows the issue is underspecified.

Why: the release's handoffs and evals showed the brief is enough (Haiku 5/5 from prep guide alone) and the separate check catches something on most medium issues; but casting every reader as an implementer made a solo session stop before completion, and a main session that reads every diff fills its context.
