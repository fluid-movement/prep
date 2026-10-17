---
type: feature
title: Local record write safety
description: Checkout-local locks, stale-plan rejection and roll-forward recovery for interrupted Markdown record transactions.
generated:
  by: codex
  at: 2026-10-14T18:20:03Z
scope:
  - internal/mdstore
  - internal/domain/ports.go
confirmed_commit: c4976b6ee5a7ffd2c68c46078332a05ad733edb7
---

# Local record write safety

Applies when changing Markdown record persistence, planning against loaded state, or recovering interrupted lifecycle writes.

Cooperating mdstore readers and writers in one checkout use an OS advisory lock at `.prep/local/write.lock`. Process death releases it. Current projects ignore `local/`; a lock-only legacy directory does not trigger the activity-ignore warning. A read outside a prep project does not create `.prep`. Lock access failures are surfaced, with no unlocked fallback. Other branches/worktrees have separate records, claims and activity; assigning parallel agents remains the orchestrator's responsibility.

`Load` resets and snapshots file hashes and relevant directory-entry manifests, including the valid issue-directory inventory while locked. `Apply` holds the lock across rendering, conflict validation and committing: frozen loaded hashes protect every record used by the plan, including untouched verification/dependency inputs, and directory membership catches newly added baselines/records or a child that would invalidate a leaf claim. Successful writes refresh the same Store snapshot for legitimate reuse. A stale writer returns actionable `E_CONFLICT` before mutation; reload and re-plan. Render rereads cannot erase older snapshot expectations.

Multi-file `Apply` changes stage complete new contents in sibling `.prep-tmp-*` files and install a versioned `.prep/.write-transaction.json` journal before the first target mutation. Each target uses atomic rename or removal. Locked `Load`, `LoadConfig` and writes recover a pending journal by rolling forward before exposing records; there is no whole-operation rollback. Targets already containing the expected new hash count as done. Unexpected target/stage content or malformed journals fail visibly with the journal retained: preserve the journal, staged data and unexpected edits while reconciling the reported paths, rather than deleting recovery metadata to bypass the error. The complete manifest is validated before any recovery install, with schema path/hash/duplicate checks and rejection of local/lock/temp targets and symlink escapes.

This is a bounded process-interruption guarantee for cooperating prep processes with access to the same checkout lock, not simultaneous atomic visibility to direct editors, power-loss atomicity, or distributed coordination. Direct edits between check and rename are outside advisory locking; recovery refuses unexpected target content it observes. Initialization, formatting and repair paths retain serialized per-file CAS writes, but are not one multi-file Apply transaction. Markdown stays authoritative; domain policy stays outside the storage adapter.

Tests in `internal/mdstore/transaction_test.go` reproduce competing stale writers, interruptions after journaling and after a target install, direct-edit and malformed-journal refusal, baseline paths, a newly created child invalidating an older planned claim, and changed verification invalidating a claim even when claim.md itself is untouched. `dependency_transaction_test.go` checks override/history recovery and append-chain conflicts across the two lifecycle slices. Contract fixtures are copied to temporary roots before locked loading. The race suite and Windows package compilation cover locking integration.
