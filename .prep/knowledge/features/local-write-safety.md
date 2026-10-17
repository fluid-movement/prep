---
type: feature
title: Local record write safety
description: Checkout-local lock held for a whole write command, atomic per-file writes and the order that writes the state-defining file last.
generated:
  by: codex
  at: 2026-10-14T18:20:03Z
scope:
  - internal/mdstore
  - internal/domain/ports.go
confirmed_commit: 28f150511800c7e93ff7bcd5e1ed4f499c1f8651
---

# Local record write safety

Applies when changing Markdown record persistence, planning against loaded state, or changing the order in which a command writes its files.

Cooperating mdstore readers and writers in one checkout use an OS advisory lock at `.prep/local/write.lock`. Process death releases it. Current projects ignore `local/`; a lock-only legacy directory does not trigger the activity-ignore warning. A read outside a prep project does not create `.prep`. Lock access failures are surfaced, with no unlocked fallback. Other branches/worktrees have separate records, claims and activity; assigning parallel agents remains the orchestrator's responsibility.

One lock per write command. `Store.Hold` (or `Store.Locked(fn)`) takes the exclusive lock and keeps it until released; while it is held, `Load`, `LoadConfig` and `Apply` on that store skip locking, because flock on a second descriptor of the same lock file blocks even inside one process. The CLI calls `(*app).lockWrite` before a write command's load and defers the release, so load, plan (`Plan`, `CheckWrite`), `Apply` and the staging that follows run under one lock; a second concurrent writer waits and then plans on the first one's result. That is why there is no hash snapshot or stale-plan check: nothing can change between load and write. Reads take the lock only for the duration of `Load`. `Init`, `Fix`, `SetIdentity` and a non-dry-run `Fmt` hold the lock across their per-file writes. The TUI's `write` (internal/cli/tui.go) builds a fresh store per write, holds the lock, loads again, plans and applies, so a long-running TUI never plans on a stale screen.

Every file is written atomically: a synced temporary file `.prep-tmp-*` in the same directory, then a rename (removals use `os.Remove`), then a directory sync where the platform supports it. There is no journal and no rollback. `Apply` writes the files of a change one at a time in `writeRank` order: supporting records first, then the files that define the issue's state (baseline, ready.md, claim.md), then resolution.md, and removals after every write. So define writes its baseline last, ready ready.md last, claim claim.md last; complete writes resolution.md and then removes claim.md (resolution wins over a claim, so an interruption leaves the issue done with a stale claim, never back in ready); release and reopen remove their state file last. An interruption leaves the previous state, or the new state with a leftover file, plus harmless extra content that `prep check` reports; run the command again.

This is a bounded process-interruption guarantee for cooperating prep processes with access to the same checkout lock, not simultaneous atomic visibility to direct editors, power-loss atomicity, or distributed coordination. Markdown stays authoritative; domain policy stays outside the storage adapter.

Tests in `internal/mdstore/hold_test.go` cover a held lock spanning load and write (a second writer blocks and then sees the first writer's result), the write order for complete, and `orderForWrite`; `lock_test.go` covers reads outside a project and baseline writes.
