---
type: decision
title: "Checkout-local serialization and recoverable multi-file transactions"
description: "Checkout-local serialization and recoverable multi-file transactions"
date: "2026-10-14"
issue: 01M4TSWYF45MFAAWHFS4PMMB8B
---

Implement the defined local safety scope using Markdown store, Storage format, Architecture and Agent activity. Historical 01M4E7XR44YR718HVTPD5N1BWT and 01M46ARR6G4K7NZBYZQJZGBCR0 established per-file rename/CAS and branch-local claim limitations. Survey reproduced the protocol race: two writers pass CAS then the later rename silently wins; interruption between sequential writes leaves mixed records. Repository tests must reproduce these actual mdstore boundaries deterministically before stronger claims.

Use a cooperating checkout-local OS advisory exclusive lock in ignored `.prep/local`, released by process death, around validation/render/commit and coherent loads. Validate loaded read-set hashes before render so rereads cannot replace a stale snapshot; keep actionable E_CONFLICT. Stage fully rendered transaction files and a durable validated journal before first target mutation; on next locked Load/Apply roll forward idempotently. Already-new content counts done; unexpected target/stage content must fail visibly without overwriting. Explicitly bound guarantees to cooperating prep processes, one checkout/local filesystem, process interruption and recoverable journal state; do not claim distributed coordination, direct-editor serialization or power-loss atomic visibility.

Own transaction/lock helpers, Apply/Load wrappers and storage support fields, with actual deterministic barrier/fault/restart tests preserving approval/evidence/requirement/history chains. Audit other public mdstore write paths for compatible locking/recovery and avoid recursive lock deadlocks or creating .prep on failed non-project reads. The parallel dependency worker owns narrow mdstore override parse/render additions; keep diff areas disjoint. Root owns all issue and knowledge writes. No domain lifecycle policy, remote lock/scheduler or branch merger.
