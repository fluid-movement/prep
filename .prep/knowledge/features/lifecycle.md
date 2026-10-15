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
| ready | defined, not stale | explicit project completion policy; human-approved inherited DoD opt-outs; at least one criterion; context.md non-empty unless manual or parent |
| claim | ready, not stale | not a parent; dependencies done, with applicable human approval for explicitly declared local or cross-project approval prerequisites |
| release | in progress | records the release in history.md |
| complete | in progress (parents: ready or in progress) | not stale; criteria checked or explicitly waived; applicable reports or approved exceptions; documentation decision; parents: children resolved; code: non-human actor, and a given --commit must be a hex hash; research: findings.md; decision: an outcome: true entry |
| review approve | done | explicit human actor and reason; full resolution/generation and scoped work fingerprint; no commit prerequisite |
| reopen | done | explicit agent actor and reason; archive earlier resolution and remove active claim; return to ready/defined with ordinary claim gates |
| drop | any unresolved state | reason; a dropped (like a done) parent with unresolved children is warning I023 |

- **Record writes** are not transitions either: `prep context`, `decide`, `criterion`, `dod`, `findings` and `log` work on any unresolved issue and never change state; `prep guide` names them in its instructions and Write list.
- **Edit** is not a transition: `prep edit` changes title, kind, parent, dependencies, tags, priority or requirement of any unresolved issue (tags and priority also on resolved ones) and appends a history line; it never changes state, but a requirement or kind edit makes a defined issue stale.
- **Define and enrich** have different owners. Define settles what and why, and the user has authority over the requirement; it needs prose and an empty Open questions section. Enrich settles how: the agent drafts context, decisions and criteria, the user agrees, and `prep ready` signs off against the newest baseline. Code may be read during define only to test whether a requirement makes sense. Expected enrichment differs by kind: code needs context from the codebase, research states the question, scope and what counts as answered, decision states the question and options, manual is optional.
- **Stale**: requirement (issue.md body) or kind differs from the newest baseline. Ack for trivial changes; define plus re-enrichment for real ones.
- **Actionable**: ready, not stale, dependencies satisfied, unclaimed, not a parent. **Blocked**: unresolved dependencies, including unknown cross-project observations. Missing/unavailable sources cannot imply completion; a current explicitly human-approved uncertainty override can permit an unknown implementation dependency, but never waive a required resolution review. See [Cross-project work](/features/cross-project.md).
- **Completion evidence** for code issues is informational: without `--commit` it is the commit that adds `resolution.md`, so code and records land in one commit (01M4AVPH9A56WJ4JFMYZT476H0); `prep complete --commit` and `prep check` validate a given hash's format only, so evidence stays valid after a squash merge removes the branch commit (decided in 01M46ARQ78P8M40K4FSPVNK531).
- **Claims across branches**: prep never pushes. A claim made on a branch is visible elsewhere once the branch merges; working on main has no lag (decided in 01M46ARR6G4K7NZBYZQJZGBCR0).
- **Definition of Done** cascades project → ancestors → issue with opt-outs and is snapshotted into resolution.md.
- Parents are never claimed; their kind is ignored while they have children.

Format of each record: [Storage format](/conventions/storage-format.md).
Policy-bound issues require passing applicable reports or explicit human-approved exceptions for each requirement, and explicit reconciliation when the contract changes. A waived acceptance criterion remains unchecked with its approval preserved. Human-approved policy retention keeps the bound requirements for one reviewed policy change; another change requires renewed review. The informational Git commit reference above is separate from these reports; commits remain optional. Legacy records are not assigned fabricated contracts or reports. See [Completion policy and verification](/features/completion-policy.md).

Policy-bound documentation decisions require current per-entry knowledge review reports for named docs or a report matching the no-impact reason; a reference or prose alone is insufficient. See [Knowledge lifecycle](/features/knowledge-lifecycle.md). Implementation completion and human review are separate: pending review does not block ordinary dependencies or unrelated work. Explicit local or cross-project approval prerequisites require applicable human resolution approval. Reopening preserves resolution, evidence and review history; derived Git evidence is presentation-only and does not change approval identity. See [Resolution review](/features/resolution-review.md).
