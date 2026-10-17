---
type: decision
title: "Write safety: one lock per write command, state file written last"
description: "Write safety: one lock per write command, state file written last"
date: "2026-10-16"
issue: 01M4ZCMFJ5P67XNSMKX1NPANDF
---

User agreed, 2026-10-16: keep the system as simple as possible. Parallel subagents in one checkout make concurrent writes real, so keep the lock file, held for the whole write command (load, plan, write); a command holding it throughout cannot plan on stale data, so the hash-snapshot stale-plan check goes for the CLI. The long-running TUI re-reads under the lock before writing. Drop the transaction journal and roll-forward recovery (transaction.go and its tests): write the state-defining file last (for example resolution.md after its supporting records), so an interruption leaves the previous state plus harmless extra data that prep check reports.
