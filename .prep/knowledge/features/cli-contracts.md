---
type: feature
title: CLI contract guidance and artifact references
description: Additive lifecycle response fields, actionable reconciliation and external workflow references that never count as verification.
generated:
  by: codex
  at: 2026-10-15T07:49:11Z
scope:
  - internal/cli
  - internal/domain/artifact.go
  - internal/domain/errors.go
confirmed_commit: d73c7a0eab6c048393bb0c2f9462fe34e3c655c7
---

# CLI contract guidance and artifact references

Applies when changing lifecycle/report/review CLI responses or linking external workflow documents. CLI remains the primary agent interface; domain planners decide validity and gates. MCP, dedicated external adapters and bidirectional synchronization need separate observed justification.

## Response compatibility

Lifecycle policy, evidence, knowledge-review and resolution-review reads retain their existing JSON fields and add `ok: true`. Issue lifecycle, report, approval and artifact writes preserve `ok`, `op`, optional `id`/`state` and `files`, adding a bounded `next` command. Maintenance writes keep their existing response shapes. Errors retain `ok: false`, stable `error.code`, `message`, optional `unmet` and `diagnostics`; domain-provided `error.next` supplies reconciliation guidance when available. Existing clients may ignore additive fields. Exit codes remain 0 success, 1 unmet/check failure, 2 usage and 3 conflict. General list/check/help caps and guide section selection are separate work.

`prep guide <id>` remains the shared lifecycle contract, not a client-specific gate implementation. Failed transitions point back to it; unmet messages identify missing policy, reports, knowledge entries or required human action. A report records checks the agent already ran. Prep does not execute those checks, and `human:` attribution does not authenticate a person. Never invent a human approval or uncertainty override.

## Reference-only artifacts

`prep artifact add <id> --source <identity> --location <URI-or-path> [--title <text>]` adds metadata. `prep artifact remove <id> --source <identity> --location <same-location>` removes that exact pair. Source and location are required, duplicate pairs fail, and all fields are single-line text. Locations are opaque: prep neither fetches nor executes them. No dedicated external system is required.

Optional `artifacts` frontmatter in `issue.md` contains `source`, `location`, and optional `title`. Legacy issues omit it. Normal edits, formatting and lifecycle transitions preserve it. `prep show --json` exposes `issue.artifacts`; text labels each artifact “reference only”. References may be maintained on resolved issues as metadata without rewriting the requirement or resolution.

A link is not imported authoritative content, a requirement baseline, verification, knowledge-review evidence or human approval. Existing completion gates still require applicable recorded reports and reviews. To record an actual externally produced check, use `prep evidence record` with the observed result, scopes and optional `--source`; that remains an agent-reported assertion, not authenticated proof.

Knowledge reads expose the stored last explicit review marker as optional `confirmed_commit` in list/show JSON and qualified provenance in show outlines. It does not prove a candidate entry is correct or current; agents still inspect scoped code and assess documentation impact.
