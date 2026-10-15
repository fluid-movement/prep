---
type: component
title: TUI Inbox and completion bundles
description: Inbox attention classification, bundled completion review, non-blocking notifications and captured approval checks.
generated:
  by: codex
  at: 2026-10-14T17:22:58Z
scope:
  - internal/tui/inbox.go
  - internal/tui/inbox_test.go
  - internal/tui/review_snapshot.go
  - internal/cli/tui_review.go
  - internal/cli/tui_review_test.go
  - internal/domain/inbox.go
  - internal/domain/inbox_test.go
confirmed_commit: 8030cc0c6ca5ae85dd8d4b1e186d0aab1b963960
---

# TUI Inbox and completion bundles

Applies when changing Inbox classification, attention navigation, completion notifications or captured review approval.

- **Inbox** (`inbox.go`, `u`): the read-only `Tree.Inbox` classifier prioritizes pending/stale completion reviews, eligible ready signoffs, open questions, stale requirements or policy, and unknown external dependencies. Ordinary backlog stays in Issues. List/detail panes retain selection across reloads, support mouse/keyboard navigation and adaptive or single-pane layout. A successful reload announces newly completed resolutions without moving focus; the initial backlog is not announced. The review bundle contains the ready requirement baseline, current acceptance checklist, recorded completion contract/DoD, selected reported verification and knowledge assessments with applicability, exception approvals, captured scopes/fingerprint, and available Git evidence/changes. `a` asks for a reason and explicitly approves the captured bundle; `e` adds work scopes and refreshes it, required for legacy completions without recorded scopes. Opening, scrolling, canceling or leaving does not approve. `y` copies the issue ID and `o` opens its normal lifecycle actions; reopening remains an explicit agent action after a user instruction. `Options.Review` and `Options.ApproveReview` keep adapter reads/writes outside `internal/tui`; stale async loads are ignored, and an open approval dialog retains its captured snapshot across reloads.

## Classification

`inbox.go`: `Tree.Inbox` returns one highest-priority attention category/reason per issue: pending/stale completion review, eligible ready signoff, open questions, stale requirement, stale policy, or unresolved unknown external dependency. A current implementation override removes ordinary unknown attention, but an unknown required resolution review remains in Inbox and blocks claiming. It omits ordinary backlog and dropped issues, and sorts category, priority, completion time and stable ID without changing lifecycle or dependency gates. It shares `ResolutionReviewStatus`, `Contract` and `KnowledgeReviewApplicable` read APIs with the TUI review adapter; the adapter owns filesystem snapshots and persistence.

## Review adapter

`internal/cli/tui_review.go` supplies the Inbox review/approval callbacks. Reads load fresh mdstore/OKF trees and resolve external relations, capture the raw resolution identity and mandatory work/config/knowledge-entry scopes, then check the fingerprint before/after collecting bounded Git display evidence. The specialized approval callback uses the existing serialized writer, compares displayed contract/requirement/checklist/baseline/status/knowledge applicability and scoped content with fresh values, and persists the shared domain review through Plan/CheckWrite/Apply. A commit is not required.

## Verification

`inbox_test.go` covers focus-preserving notifications, captured approval across reloads, stale async responses, legacy scope selection, reported/missing/stale artifacts, mouse/back navigation, scrolling and 80×24/110×28 list/bundle goldens. `tui_review_test.go` tests commit-free approval and refusal after code, configuration, resolution, knowledge entry, current contract, requirement or another human review changes. See [Resolution review](/features/resolution-review.md) for shared lifecycle semantics.
