---
type: feature
title: Resolution review and reopening
description: Human approvals bound to completed resolution history and scoped snapshots.
generated:
  by: codex
  at: 2026-10-14T12:33:05Z
scope:
  - internal/domain
  - internal/mdstore
  - internal/cli
  - internal/tui/inbox.go
  - internal/tui/review_snapshot.go
confirmed_commit: c4976b6ee5a7ffd2c68c46078332a05ad733edb7
---

Human approval is a separate post-completion action; it does not replace verification waivers, policy retention or Definition of Done opt-outs. Run `prep review show <id>` to inspect the active resolution and status (`pending`, `approved` or `stale`). `prep review approve <id> --reason <text> --by human:<name>` approves the done resolution. Prep adds the work and check-configuration scopes recorded by verification, exception approvals and selected knowledge-review reports, including the named knowledge entry files; legacy resolutions with no recorded work scopes require explicit repeatable `--scope <path>`. `prep guide <id>` and `prep show <id>` surface the required next action. This is local human attribution, not authentication, and agents must not invent approval.

The approval binds the full resolution content, its generation, and fingerprints of the selected work/configuration scopes including selected knowledge-review work/configuration scopes and entry files. A changed scoped snapshot or a later resolution makes an older review stale. Fingerprints compare current scoped content; unobserved edits away and back between reads cannot be detected. Supported reopen/re-complete transitions archive the earlier resolution and increment generation so historical approval cannot revive. Reopen with `prep reopen <id> --reason <text>` returns the issue to its prior ready or defined state and removes its active claim; normal claim gates still apply. Reopening does not require a commit.

Human review is opt-in. Pending review does not block unrelated ready work. To require human approval for a local or cross-project dependency, include it in both `--depends-on <id>` and `--depends-on-approval <id>` when editing the dependent issue; the ordinary dependency must already be done and the explicit approval prerequisite must also be approved before claiming. Approval prerequisites must be ordinary dependencies. External targets must be freshly found/done with applicable approved review; an unknown-source implementation override cannot waive this review requirement. No commit is required before completion or review.

The append-only `resolution-history.md` preserves earlier completed resolution records; `resolution-reviews.md` preserves every review action. Review records name the resolution content ID and generation, human actor, reason, scopes, snapshot, timestamp and prior review ID. Legacy resolution records remain absent of fabricated generation metadata; review binding uses a compatibility generation without rewriting those records.

Derived Git evidence is presentation-only: committing unchanged code and records after approval does not change the resolution identity or invalidate that approval. Named knowledge files are fingerprinted with their metadata; coverage remains the recorded scope, not proof that every relevant path was identified. The selected knowledge report identities and entry revisions are part of the full resolution hash.
The TUI Inbox (`u`) offers a bundled review and one explicit Approve action with a reason. It captures the displayed resolution identity, effective current completion contract, requirement/current checklist, ready baseline, review status, selected knowledge-report applicability and work/configuration fingerprint. The CLI adapter reloads a fresh tree inside its serialized write plan and refuses approval if these displayed facts changed, including another reviewer acting in the meantime; reload and inspect the fresh bundle. The scope picker can add explicit paths and refreshes the bundle before approval; legacy completions require at least one work scope. Git evidence/diffs are display metadata; absent commits, missing artifacts, untracked files or truncated diffs are reported rather than treated as verification. Leaving review keeps it pending; copy the issue ID to discuss feedback with the agent and instruct reopening explicitly.
