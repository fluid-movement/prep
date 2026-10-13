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

This is the first increment of issue 01M4TSWX8HF105ZQQKS50MH6RP: evidence validation, policy binding/reconciliation, human exceptions, and completion review are still pending. The current policy record alone does not prove that checks ran or that documentation was updated.
