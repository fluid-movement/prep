- 2026-10-16T18:40:49Z claude-code: Note 2026-10-16: go.mod lists golang.org/x/sys as indirect although the lock code imports it directly; go mod tidy moves it. Settle it with this change.

- 2026-10-17T08:03:39Z edited by claude-code/opus-5.5: depends_on

- 2026-10-17T12:52:44Z claude-code: Implemented, ready for check: write commands hold the checkout lock from load to last write (Store.Hold, app.lockWrite, TUI), journal/roll-forward/hash snapshots removed, files written atomically with the state-defining file last

- 2026-10-17T12:59:01Z claude-code: Checker: knowledge confirm now holds the lock across its drift load; removals are ordered after all writes (complete writes resolution.md before removing claim.md, so an interruption never leaves a completed issue in ready); added a test that an unchanged edit leaves no temp file and removed 11 .prep-tmp files the old journaled writer had left and staging had committed.
