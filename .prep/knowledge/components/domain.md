---
type: component
title: Domain package
description: internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/domain
confirmed_commit: 0a4f01a06a9df5c42a1051c92986c8d2ee2731ba
---

# Domain package

Applies when changing lifecycle rules, gates, diagnostics or queries.

- `types.go`: `Issue`, records (`Baseline`, `Ready`, `Claim`, `Resolution`), `Entry`, kinds and states.
- `tree.go`: `Tree.State` derives state from records; `Stale`, `Blocked`, `Actionable`, `ChildProgress`, `EffectiveDoD`, ID suffix resolution.
- `fsm.go`: `Gates` (unmet conditions with `G_*` codes; with nil input, argument-dependent gates become needs), `Plan` (records to write), `PlanNew`, `PlanEdit` (edits of title, kind, parent, dependencies and requirement on unresolved issues; refuses dropped dependency targets and leaves existence and cycles to `CheckWrite`), `Apply` and `CheckWrite`.
- `records.go`: `PlanRecord` for the record writes (context, findings, decide, criterion, dod, log) on unresolved issues, with gates `G_CRITERION`, `G_DECISION`, `G_DOD` and `G_KIND`; the next decision ID is `D<max+1>`. Acceptance changes are `AcceptanceOp` operations, applied in memory by `ApplyAcceptance` (index operations refer to the numbering before the change, additions follow).
- `bootstrap.go`: `Bootstrapped` (the overview entry exists), `BootstrapStart` (first open bootstrap-tagged leaf), `BootstrapAlert`, `PlanBootstrap` (parent and survey issues with requirement, context and criteria, planned on a progressively applied tree). `PlanKnowledge` refuses new entries other than the overview before bootstrapping (`G_BOOTSTRAP`); `Guide.Alerts` carries the alert.
- Tags: `Issue.Tags`, `NormalizeTags`/`ValidTag` (lowercase, `. _ - /`), `Filter.Tags` (repeated values OR), validation code I026. Tags are outside the baselined requirement, so they never make an issue stale; `PlanEdit` allows a tag-only edit on resolved issues.
- `knowledge.go`: `KnowledgeEdit` and `Tree.PlanKnowledge` (path form, no rename, existence, required fields for new entries, scope required for `confirmed_commit`); `Change.KnowledgeEntry` carries the adapter-rendered entry so `Tree.Apply` and `CheckWrite` validate it. The `KnowledgeStore` port has `Load`, `Render` and `Apply`.
- `guide.go`: each `Pointer` in the Write list carries the `command` that writes it.
- `validate.go` and `diag.go`: diagnostics with stable codes (`P*` project, `I*` issues, `K*` knowledge), severity and class (fixable, guided, manual). Codes are never reused.
- `query.go`: `ParseFilter` (different flags AND, repeated flags OR), `Query`, tree ordering, `ViewNames` (saved views in config order).
- `guide.go`: `BuildGuide` step contracts and `Candidates` knowledge retrieval; `ScopeMatch` globbing.
- `errors.go`: `E_*` error codes for commands.

Rules worth knowing: see [Issue lifecycle](/features/lifecycle.md).
