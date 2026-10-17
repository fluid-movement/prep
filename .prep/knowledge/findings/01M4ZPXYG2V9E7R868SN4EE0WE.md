---
type: finding
title: "Findings: Lean knowledge policy and comparative usefulness experiments"
description: "Findings: Lean knowledge policy and comparative usefulness experiments"
date: "2026-10-15"
issue: 01M4TSWY40Z4473D9SPRTT711C
---

# Bounded Codex pilot findings

This is a task-specific pilot, not a general knowledge-quality estimate. It completed 12 Codex sessions: three each for code-only, current, lean, and deliberately stale knowledge. Every session used the pinned `bacc2245668c64831389dced173514b2db6000df` fixture, issue `01M4ZCMNCGTDQDYH1PSTS70B5H`, task prompt, common overview, and Codex skill (SHA-256 `85f9cfca3ae57c7c4b67dfc59d44af41bc64eba86120e814b0816829519f0fbd`). Codex CLI was 0.162.1 and `gpt-6.1-sol` was requested; the CLI event stream did not report the model name. The differently-manifested preflight smoke is excluded. `results.json` assigns distinct run IDs and preserves original run IDs and source streams.

All 12 sessions reached `done` without a stale issue. Each recorded four applicable pass evidence reports (12 per condition), but these are agent reports, not semantic proof. Resolution review remained pending because no human approval was created. `evidence_pending=true` reflects absent commit provenance, which was not a completion prerequisite.

The captured command events show one successful Node helper/test check in every run; the actual outer command strings, output excerpts, and outer shell statuses are in `results.json`. There was exactly one explicit `sh test.sh` invocation: in `lean-2` it exited 1 with `FAIL: totals are rounded down, expected nearest cent`. The other 11 runs did not execute `test.sh`; they must not be counted as passing it. Direct `prep check` invocations exited 0. Where an output first reported a K005 knowledge warning, a later direct check reported zero errors and warnings. Compound shell events retain their whole command and outer status; no subcommand exit status is inferred.

All 12 post-hoc Node v24.22.0 runtime checks imported the actual `profileHeadline` helper and confirmed a 200-character input is returned unchanged. The original fixture helper failed the same 200-character check. This is helper-level runtime evidence, not browser rendering or deployed database verification. One actual SQLite check confirmed the schema used by that trial, but an over-width `VARCHAR(n)` insert/select alone would not prove a capacity limit because SQLite accepts such values. The old static proxy was syntactic and unreliable (it marked two lean runs false due to any slice expression); it is retained only for audit and must not be used as a correctness score.

| Condition | n | Mean elapsed | CLI-reported input / output tokens | Applicable knowledge reports |
| --- | ---: | ---: | ---: | --- |
| Code-only | 3 | 104.9 s | 734,172 / 10,042 | 3 unchanged |
| Current | 3 | 137.4 s | 1,188,636 / 13,959 | 6 updated, 1 unchanged |
| Lean | 3 | 118.5 s | 850,847 / 10,996 | 3 updated |
| Stale | 3 | 97.4 s | 733,790 / 9,434 | 3 unchanged |

Token totals include repeated and cached context; they do not measure unique knowledge size. Knowledge-report counts are a maintenance proxy; human maintenance time was not measured. No price estimate or cross-model comparison was made. With only three trials per cell, these observations are descriptive.

There is a treatment confound: current knowledge includes `/conventions/testing.md` and its required test instruction; the code-only, lean, and stale overlays omit it. Therefore, the as-run lean condition drops a global testing constraint and cannot support a lean-superiority claim. A future lean overlay should retain that mandatory check instruction verbatim or equivalently, and compress only advisory details. For example, keep the common overview plus: “Before completion, run `./test.sh`; it must pass. Record its actual result and stop for reconciliation on failure. For the headline change, inspect `db/schema.sql` and `src/lib/profile.ts`; keep storage and display behavior aligned, and prefer current code over stale notes.”

The stale overlay starts by claiming `VARCHAR(200)` and no rendering truncation. The final helper has no truncation, so that behavioral claim agrees with the code; the final schema is `TEXT`, so the entry's declared-type claim remains inaccurate. The agent reported it unchanged in all three stale runs. This supports only that the stale note was not reported as updated, not that the model followed or misunderstood it. Current runs reported updating profile-fields and player-list entries; code-only reported its overview unchanged; lean reported its profile-fields entry updated. These are human/agent-authored assessments, not semantic correctness checks.

No general usefulness ranking follows from this pilot. It demonstrates the need to make global required checks visible in every treatment and to evaluate actual command traces separately from reported evidence and knowledge-review records.
