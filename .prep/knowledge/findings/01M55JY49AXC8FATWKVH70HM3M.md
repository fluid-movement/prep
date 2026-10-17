---
type: finding
title: Eval run A after the rethink guidance
description: Handoff and checker roles pass 3/3; agents miss user-stated decisions and do not look up decisions before coding; three cases are obsolete
date: "2026-10-17"
issue: 01M55JY43XT800M50Y6YFC5YJ0
---

Eval run 20261017-183812-135e (Sonnet, 3 runs per case, $1.73) after the rethink's guidance changes.

Pass 3/3: implement-handoff-from-guide, implement-checker-completes, knowledge-write-new-entry, natural-fix-no-issue, natural-fix-opt-out, natural-read-only. The handoff from prep guide alone and the checker role work, and tracking holds with the prompt hook silent on repeats.

Failures, read from the transcripts:
- implement-complete 0/3 and implement-log-and-stop 0/3 are obsolete cases, not regressions: in the shop fixture Order emails has no code and its context names an event the stub code lacks, so the checker correctly refused to complete unverified work and the implementer correctly logged the context conflict and stopped. checker-completes and handoff-from-guide cover both roles with fixtures that have the code.
- implement-record-code-decision 0/3 is a real gap: agents ran prep decide, but recorded their own small choice (cache keyed by rows identity) instead of the decision the user stated in the prompt (cache instead of recompute, with the reason). The rule 'record a decision when it is made' reads as the agent's own choices.
- knowledge-read-respect-decision 1/3 is a real gap: no agent looked up decisions about src/cart.go before acting; all stopped to settle the requirement with the user, safe but blind to the recorded decision that forbids the cache.
- natural-fix-matching-issue 0/3: obsolete rubric; agents found the issue, claimed, changed the code, recorded the new limit with prep decide and logged ready for check, but the rubric expects a descriptive knowledge entry to be edited.
