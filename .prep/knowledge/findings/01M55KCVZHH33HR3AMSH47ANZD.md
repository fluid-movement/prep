---
type: finding
title: Haiku is enough for handed-off small and medium issues
description: Haiku passes the handoff evals 5/5 at a tenth of Sonnet's cost; use Sonnet for subtle code; the check keeps quality
date: "2026-10-17"
issue: 01M4QY3H1NH122T4DZ6XCJJ55M
---

Question: which model is enough to implement a handed-off ready issue from prep guide alone, and at what cost per passing run.

Evals (2026-10-17, five runs per cell, local): implement-handoff-from-guide and implement-checker-completes passed 5/5 on Haiku at $0.004 per run (runs 20261017-184512-3a80) and 5/5 on Sonnet at $0.042 to $0.046 per run (20261017-184752-b4a2). Five runs cannot separate the models (the 95% intervals are 57 to 100% for both), but they show Haiku at parity on the handoff path at about a tenth of the cost. Sonnet also passed both 3/3 in run 20261017-183812-135e.

Real handoffs during release 0.3.0 (logged on this issue): Codex 9 issues, 4 needing checker fixes, mostly on the larger issues; Haiku 6 issues, 2 small fixes (an empty-file staging warning, per-row columns that misaligned a list), the rest clean, including two implemented from the guide brief alone; Sonnet 5 issues, 2 fixes (a lock gap and write order in store concurrency; two dropped rules). Every implementer needed the check on some issues; the check is what keeps quality, whatever the model.

Answer: Haiku is enough for small and medium issues with a mapped context, at a tenth of Sonnet's cost; use Sonnet for subtle code (concurrency, ordering, layout rules) and for drafting or judgment work; Codex is comparable to Sonnet on large removals. The checker always verifies each criterion itself.
