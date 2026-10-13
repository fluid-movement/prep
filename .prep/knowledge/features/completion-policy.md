---
type: feature
title: Explicit project completion policy
description: Project verification and documentation methods, configured through policy set and required for new readiness transitions.
generated:
  by: codex
  at: 2026-10-13T16:47:43Z
---

A project explicitly records how work is verified and how documentation impact is reviewed with `prep policy set --verification <method> --documentation <method>`. `prep policy show` reports those methods, attribution and configured state; JSON also includes the existing project Definition of Done. prep records the contract and never executes check commands.

The policy is stored as `completion_policy` in project.md frontmatter, with verification, documentation, by and at fields. Attribution is not authenticated human approval. Existing project DoD bullets continue to cascade; they do not count as an explicit completion policy. Missing policy blocks new ready transitions with G_POLICY, but does not block capture or definition. Legacy ready and resolved records are not rewritten or assigned fabricated evidence. A malformed configured policy is P010.

New ready records snapshot a completion contract: verification and documentation methods, effective DoD items and opt-outs, acceptance criterion text, and ancestor IDs. Requirement IDs and the contract ID are deterministic content hashes. Criterion checkboxes, policy attribution/time, UI configuration, and Git commit identity do not change contract identity. Changed methods, effective requirements, or ancestry invalidate the binding. Claim and completion then require explicit reconciliation; policy changes do not rewrite historical resolutions. `prep policy show <id>` compares the bound and current contract; `prep policy adopt <id>` explicitly adopts the current one without changing the requirement baseline or claiming work. Claimable-work queries exclude changed-policy issues. I029 warns about changed bindings; inconsistent stored contract hashes are I015 errors.

These are initial increments of issue 01M4TSWX8HF105ZQQKS50MH6RP. Per-requirement evidence validation and human exceptions are still pending, as are completion approval/reopening and the TUI review bundle. A contract snapshot alone does not prove that checks ran or documentation was updated. Legacy ready records have no contract until explicitly adopted; nothing infers human approval from legacy DoD text.
