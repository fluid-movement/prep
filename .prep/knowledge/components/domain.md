---
type: component
title: Domain package
description: internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/domain
confirmed_commit: 86c98f9e49a6fa46e114332c24e3fc2fb8b0f6f8
---

# Domain package

Applies when changing lifecycle rules, gates, diagnostics or queries.

- `types.go`: `Issue`, records (`Baseline`, `Ready`, `Claim`, `Resolution`), `Entry`, kinds and states.
- `tree.go`: `Tree.State` derives state from records; `Stale`, `Blocked`, `Actionable`, `ChildProgress`, `EffectiveDoD`, ID suffix resolution.
- `fsm.go`: `Gates` (unmet conditions with `G_*` codes; with nil input, argument-dependent gates become needs), `Plan` (records to write), `PlanNew`, `PlanEdit` (edits of title, kind, parent, dependencies and requirement on unresolved issues; refuses dropped dependency targets and leaves existence and cycles to `CheckWrite`), `Apply` and `CheckWrite`.
- `records.go`: `PlanRecord` for the record writes (context, findings, decide, criterion, dod, log) on unresolved issues, with gates `G_CRITERION`, `G_DECISION`, `G_DOD` and `G_KIND`; the next decision ID is `D<max+1>`. Acceptance changes are `AcceptanceOp` operations, applied in memory by `ApplyAcceptance` (index operations refer to the numbering before the change, additions follow).
- `guide.go`: each `Pointer` in the Write list carries the `command` that writes it.
- `validate.go` and `diag.go`: diagnostics with stable codes (`P*` project, `I*` issues, `K*` knowledge), severity and class (fixable, guided, manual). Codes are never reused.
- `query.go`: `ParseFilter` (different flags AND, repeated flags OR), `Query`, tree ordering, `ViewNames` (saved views in config order).
- `guide.go`: `BuildGuide` step contracts and `Candidates` knowledge retrieval; `ScopeMatch` globbing.
- `errors.go`: `E_*` error codes for commands.

Rules worth knowing: see [Issue lifecycle](/features/lifecycle.md).
