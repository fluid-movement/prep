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
confirmed_commit: 0a4f01a06a9df5c42a1051c92986c8d2ee2731ba
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

- **Record writes** are not transitions either: `prep context`, `decide`, `criterion`, `dod`, `findings` and `log` work on any unresolved issue and never change state; `prep guide` names them in its instructions and Write list.
- **Edit** is not a transition: `prep edit` changes title, kind, parent, dependencies or requirement of any unresolved issue and appends a history line; it never changes state, but a requirement or kind edit makes a defined issue stale.
- **Define and enrich** have different owners. Define settles what and why, and the user has authority over the requirement; it needs prose and an empty Open questions section. Enrich settles how: the agent drafts context, decisions and criteria, the user agrees, and `prep ready` signs off against the newest baseline. Code may be read during define only to test whether a requirement makes sense. Expected enrichment differs by kind: code needs context from the codebase, research states the question, scope and what counts as answered, decision states the question and options, manual is optional.
- **Stale**: requirement (issue.md body) or kind differs from the newest baseline. Ack for trivial changes; define plus re-enrichment for real ones.
- **Actionable**: ready, not stale, dependencies done, unclaimed, not a parent. **Blocked**: unresolved dependencies.
- **Completion evidence** for code issues is informational: `prep complete --commit` and `prep check` validate the hash format only, so evidence stays valid after a squash merge removes the branch commit (decided in 20261005-152625).
- **Claims across branches**: prep never pushes. A claim made on a branch is visible elsewhere once the branch merges; working on main has no lag (decided in 20261005-152626).
- **Definition of Done** cascades project → ancestors → issue with opt-outs and is snapshotted into resolution.md.
- Parents are never claimed; their kind is ignored while they have children.

Format of each record: [Storage format](/conventions/storage-format.md).
