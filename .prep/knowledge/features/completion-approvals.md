---
type: feature
title: Human approval for completion contracts
description: Explicit scoped waivers, policy retention and inherited DoD approval, including history and local trust limits.
generated:
  by: codex
  at: 2026-10-14T10:37:10Z
scope:
  - internal/domain/approval.go
  - internal/cli/approval.go
  - internal/mdstore/approval.go
---

# Completion approvals

Applies when changing verification waivers, policy retention or inherited DoD opt-outs. Prep records explicit human actions; it does not run checks or require a commit before completion.

- `prep exception approve <id> --requirement <hash> --reason <text> --scope <path> --by human:<name>` waives a bound requirement for explicitly scoped work. Repeat --scope and include --config-scope for check configuration. The record binds the contract, requirement, latest report ID (or its absence), work/configuration fingerprint, approver and reason. Changed scoped content, a later report or a different contract requires renewed approval. `prep exception show <id>` reports applicability. Missing or failed reports can be waived; an unmet acceptance criterion stays visibly unchecked and its waiver is preserved.
- `prep policy retain <id> --reason <text> --by human:<name>` preserves the bound policy after review of the current one. Approval stores the bound ID, full reviewed current contract and policy_revision. Another policy or inherited DoD revision requires new approval, including a change away and back. Re-recording identical policy methods does not expire retention. Changed acceptance text, ancestry or opt-out choices require adoption instead. `prep policy show <id>` exposes changed, review_required, retained and retention_approval in JSON, distinguishing current from bound requirements. Reports and resolution DoD continue to use bound requirements.
- `prep dod <id> --opt-out <item> --reason <text>` requests an inherited DoD exemption; it does not approve it. `prep dod approve <id> --item <item> --reason <approval reason> --by human:<name>` approves the exact request on its originating issue, applying to its subtree. Approval binds request_revision, so changing the reason away and back or removing/recreating a request requires renewed approval. New readiness and contract adoption require approval for applied opt-outs; bound unresolved issues cannot proceed with unapproved opt-outs. Legacy records without a contract remain readable and gain no invented approval.

All actions require a nonempty declared human:<name> actor and reason. Approval is an explicit action, never inferred from attribution on policy, ready or evidence records. Actor names and Markdown records form a local trust model, not authenticated identities or signatures. Agents must not invent human approval. Resolution review, reopening and Inbox belong to later issues.

approvals.md contains an approvals frontmatter list with no body, plus optional dod_revision and opt_out_revisions heads. Each approval has id, previous, kind, reason, by and at. Exceptions add contract, requirement, optional evidence, scopes, optional config_scopes and snapshot. Retention adds contract, reviewed and policy_revision. Opt-out approvals add requirement (the exact item/reason hash), item, opt_out_reason and request_revision. Parsers validate shapes, human attribution, content IDs, chain order and stored revision heads. Work applicability is derived by the adapter. Resolutions copy only approvals actually used, including inherited opt-out approvals.

CLI-mediated DoD edits advance chained revision heads only when additions or opt-out requests change; checking acceptance criteria does not. Removed request heads are retained to prevent approval revival on recreation. Policy revisions include method history and the ancestor/issue DoD heads. Project DoD has no CLI edit command or revision history; unobserved manual edits away and back cannot be detected. Direct manual history rewriting is outside this local trust guarantee. Compare-and-swap rejects writers already stale at validation; it does not serialize simultaneous processes, which remains local write-safety work.

Policy changes archive previous methods/attribution in project.md completion_policy_history. Contract adoption and new readiness preserve earlier bindings in ready.md previous_contracts; repeating an unchanged binding does not duplicate it. Legacy omissions remain absent. These are versioned Markdown records independent of commit timing.
