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
confirmed_commit: 22a58e58c0cf17a7b64b7477344c30797dc116cf
---

# CLI contract guidance

Applies when changing lifecycle CLI responses. CLI remains the primary agent interface; domain planners decide validity and unmet conditions. MCP, dedicated external adapters and bidirectional synchronization need separate observed justification.

## Response compatibility

Lifecycle writes preserve their response fields (`ok`, `op`, optional `id`/`state` and `files`) and may add a bounded `next` command. Errors retain `ok: false`, stable `error.code`, `message`, optional `unmet` and `diagnostics`; domain-provided `error.next` supplies guidance when available. Existing clients may ignore additive fields. Exit codes remain 0 success, 1 unmet/check failure, 2 usage and 3 conflict. General list/check/help caps and guide section selection are separate work.

`prep guide <id>` is the shared lifecycle contract. Failed transitions point back to it. Completion checks current issue state and staleness, checked criteria, the kind's result and a documentation decision. Prep does not execute project checks; a report of a check is an agent assertion. See [Completion requirements](/features/completion-policy.md).

Knowledge reads expose the stored last explicit review marker as optional `confirmed_commit` in list/show JSON and qualified provenance in show outlines. It does not prove an entry is correct or current; agents inspect scoped code and assess documentation impact.
