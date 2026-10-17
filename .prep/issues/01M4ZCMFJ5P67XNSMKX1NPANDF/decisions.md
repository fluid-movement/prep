## D1: Direction: Markdown in git as truth, a derived local SQLite index as the read source
date: 2026-10-16

User direction with agent recommendation, 2026-10-16. Markdown stays the committed source of truth because it merges; SQLite cannot be resolved in a merge conflict. One file per record named by ULID, so new records never conflict; old records are never changed (a new decision states what it supersedes; status is derived). A gitignored SQLite index, covering issues and decisions with full-text search, becomes the read source. Prep keeps it fresh itself on every command (stat record files, re-parse changed ones, update on its own writes), not the TUI, because agents run prep without a TUI. Data flows one way, Markdown to index; the index is disposable and rebuilt on demand. Similar to Beads (JSONL in git plus local SQLite), but file-first, so nothing syncs back. Design the record format for it now; build it when measured read time warrants it (about 0.3 s once the contract machinery is gone; decisions may grow by thousands a year).

## D2: Record layout: one file per transition, no empty files, 0.3.0 records removed
date: 2026-10-16

User agreed, 2026-10-16. Keep one directory per issue with state derived from which transition files exist. Create a record file only when something is written to it (today 145 of 221 decisions.md and 63 history.md are empty). Remove the 0.3.0 record types (verification, approvals, knowledge-reviews, resolution-reviews, resolution-history, dependency-overrides, contract fields) and the unused artifacts frontmatter (zero uses). decisions.md and findings.md move to knowledge records. Minimal reopen moves the old resolution into history.md. context.md and acceptance.md stay separate files so a reader loads only what it needs. history.md gets merge=union in .gitattributes so parallel appends on branches do not conflict.

## D3: Write safety: one lock per write command, state file written last
date: 2026-10-16

User agreed, 2026-10-16: keep the system as simple as possible. Parallel subagents in one checkout make concurrent writes real, so keep the lock file, held for the whole write command (load, plan, write); a command holding it throughout cannot plan on stale data, so the hash-snapshot stale-plan check goes for the CLI. The long-running TUI re-reads under the lock before writing. Drop the transaction journal and roll-forward recovery (transaction.go and its tests): write the state-defining file last (for example resolution.md after its supporting records), so an interruption leaves the previous state plus harmless extra data that prep check reports.

## D4: No compatibility code: convert once, remove migrate and legacy paths
date: 2026-10-16

User agreed, 2026-10-16, following the rule of no compatibility before outside users. Convert this repository once to the new layout, then delete legacy-record reading paths and prep migrate. Keep prep fmt (canonical diffs) and a small prep fix.

## D5: Stay in OKF where possible; document any break
date: 2026-10-16

User direction, 2026-10-16. OKF is still beta; staying in the format is preferred, and a break is acceptable if documented and reconciled later. Known tension: OKF requires a generated index.md per directory listing every entry, so new decision records on parallel branches all change the same index and conflict on merge, defeating one-file-per-record. Since the index is derived, first try resolving conflicts by regenerating it (prep fix, or a git merge driver); if that is too clumsy, it is the first candidate for a documented break.

## D6: Verdicts carried into release 0.3.0
date: 2026-10-17
outcome: true

Verdicts recorded in this issue's decisions, 2026-10-16/17, are carried into release 0.3.0 (01M52ZNHXFXHNDATT9WWQ3C87E) by 01M52ZQKG3PZZ6YS5P4Z81FKGC, 01M52ZQSEP2EKTYKYFC79EGFS6, 01M52ZQZBAS48ZK64MW7H85QJ4, 01M52ZRB3VY0MSZ0HZWJ2SEC24. Open questions raised here were settled by the release's decision issues.
