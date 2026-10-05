---
type: feature
title: Issue lifecycle
description: States derived from records, transitions with their gates per kind, stale drift handling, parents, dependencies and Definition of Done.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/domain/fsm.go
  - internal/domain/tree.go
confirmed_commit: db27d6b1996f239fe8e18dc951b950ad5cab77ec
---

# Issue lifecycle

Applies when working on gates, transitions or derived state.

States (never stored): resolution.md → done/dropped; claim.md with valid ready → in progress; ready.md referencing the newest baseline → ready; any baseline → defined; else open.

| Transition | From | Gates |
| --- | --- | --- |
| define | open, or defined/ready when stale | title, valid kind, non-empty prose, empty Open questions |
| ack | defined, ready, in progress when stale | empty Open questions; rewrites ready.md to the new baseline |
| ready | defined, not stale | at least one criterion; context.md non-empty unless manual or parent |
| claim | ready, not stale | not a parent; dependencies done |
| release | in progress | records the release in history.md |
| complete | in progress (parents: ready or in progress) | not stale; all criteria checked; documentation decision; parents: children resolved; code: commit evidence and non-human actor; research: findings.md; decision: an outcome: true entry |
| drop | any unresolved state | reason |

- **Edit** is not a transition: `prep edit` changes title, kind, parent, dependencies or requirement of any unresolved issue and appends a history line; it never changes state, but a requirement or kind edit makes a defined issue stale.
- **Stale**: requirement (issue.md body) or kind differs from the newest baseline. Ack for trivial changes; define plus re-enrichment for real ones.
- **Actionable**: ready, not stale, dependencies done, unclaimed, not a parent. **Blocked**: unresolved dependencies.
- **Definition of Done** cascades project → ancestors → issue with opt-outs and is snapshotted into resolution.md.
- Parents are never claimed; their kind is ignored while they have children.

Format of each record: [Storage format](/conventions/storage-format.md).
