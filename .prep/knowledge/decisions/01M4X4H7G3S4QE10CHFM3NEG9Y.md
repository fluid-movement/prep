---
type: decision
title: "Separate resolution review from existing approvals and gate reopen through ready"
description: "Separate resolution review from existing approvals and gate reopen through ready"
date: "2026-10-14"
issue: 01M4TSWXEPBGMQG271CZJBBZJP
---

Store issue-resolution reviews in a dedicated append-only record chain and archive completed resolutions when reopening, keeping both separate from approvals for policy/verification exceptions and from knowledge-entry reviews. Bind each review to the full resolution content ID, monotonic generation and current fingerprint over explicitly recorded work/configuration scopes plus selected knowledge-review scopes. This prevents reuse after a supported reopen/re-complete cycle and invalidates approval when scoped content changes. Require explicit scopes when legacy evidence does not supply any so approval cannot silently cover an empty snapshot. Keep the human actor/reason as an explicit local-trust assertion, with no signature or commit prerequisite.

Reopen archives the current resolution, clears only active resolution state and returns the same issue to ready. A subsequent ordinary claim applies existing stale, policy and dependency gates; creating an implicit claim during reopen could bypass those checks. Make approval required for a dependency only through a separate list constrained to existing ordinary local dependencies, preserving current defaults and cycle checks. Pending review remains informational and does not alter global actionability.

Alternatives considered: extending the existing `approvals.md` chain would conflate three approval meanings; storing only a boolean would lose history and allow old approval to revive; reopening directly into in-progress would bypass ordinary claim gates. Use separate resolution review/history records, generation-bound applicability, and ready-state reopening.
