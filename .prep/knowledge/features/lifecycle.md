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
confirmed_commit: bacc2245668c64831389dced173514b2db6000df
---

# Issue lifecycle

Applies when working on gates, transitions or derived state.

States (never stored): resolution.md → done/dropped; claim.md with valid ready → in progress; ready.md referencing the newest baseline → ready; any baseline → defined; else open.

| Transition | From | Gates |
| --- | --- | --- |
| define | open, or defined/ready when stale | title, valid kind, non-empty prose, empty Open questions |
| ack | defined, ready, in progress when stale | empty Open questions; after a kind change on a signed-off issue, the context ready would require; rewrites ready.md to the new baseline |
| ready | defined, not stale | at least one criterion; context.md non-empty unless manual or parent |
| claim | ready, not stale | not a parent; dependencies done |
| release | in progress | records the release in history.md |
| complete | defined or in progress for non-parent decision/research; in progress for code/manual; parents: ready or in progress | not stale; all criteria checked when any exist; decision requires an outcome: true entry; research requires findings.md; documentation decision; parents require children resolved. Any human or agent can complete code; there is no actor gate. |
| reopen | done | reason; appends the removed resolution.md to history.md and returns the issue to ready |
| drop | any unresolved state | reason; a dropped (like a done) parent with unresolved children is warning I023 |

- **Record writes** are not transitions either: `prep context`, `decide`, `criterion`, `dod`, `findings` and `log` work on any unresolved issue and never change state; `prep guide` names them in its instructions and Write list.
- **Edit** is not a transition: `prep edit` changes title, kind, parent, dependencies, tags, priority or requirement of any unresolved issue (tags and priority also on resolved ones) and appends a history line; it never changes state, but a requirement or kind edit makes a defined issue stale.
- **Define and enrich** have different owners. Define settles what and why, and the user has authority over the requirement; it needs prose and an empty Open questions section. Enrich settles how: the agent drafts context, decisions and criteria, the user agrees, and `prep ready` signs off against the newest baseline. Code may be read during define only to test whether a requirement makes sense. Expected enrichment differs by kind: code needs context from the codebase, research states the question, scope and what counts as answered, decision states the question and options, manual is optional.
- **Defined decision and research issues** may record their result and complete directly: decision records its outcome, research records findings, then completes. Readiness and claiming remain available when either kind should be handed off for work.
- **Stale**: requirement (issue.md body) or kind differs from the newest baseline. Ack for trivial changes; define plus re-enrichment for real ones.
- **Actionable**: ready, not stale, dependencies satisfied, unclaimed, not a parent. **Blocked**: unresolved dependencies, including unknown cross-project observations. Missing/unavailable sources cannot imply completion. See [Cross-project work](/features/cross-project.md).
- **Completion evidence** for code issues is informational: without `--commit` it is the commit that adds `resolution.md`, so code and records land in one commit (01M4AVPH9A56WJ4JFMYZT476H0); `prep complete --commit` and `prep check` validate a given hash's format only, so evidence stays valid after a squash merge removes the branch commit (decided in 01M46ARQ78P8M40K4FSPVNK531).
- **Claims across branches**: prep never pushes. A claim made on a branch is visible elsewhere once the branch merges; working on main has no lag (decided in 01M46ARR6G4K7NZBYZQJZGBCR0).
- **Definition of Done** is a project-level list shown as guidance in the briefing and at the check step of `prep guide`; it is not a completion gate. `prep dod --add <item>` and `--remove <item>` edit that list.
- Parents are never claimed; their kind is ignored while they have children.

Format of each record: [Storage format](/conventions/storage-format.md).
