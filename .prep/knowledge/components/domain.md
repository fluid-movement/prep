---
type: component
title: Domain package
description: internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/domain
confirmed_commit: 2f70ffa1d6383f5b075d48b0d0a817ca1fea00d9
---

# Domain package

Applies when changing lifecycle rules, diagnostics or queries.

- `types.go`: issues, lifecycle records and entries. Optional `Error.Next` carries [CLI guidance](/features/cli-contracts.md).
- `tree.go`: `Tree.State` derives state from records; `Stale`, `Blocked`, `Actionable`, `ChildProgress`, ID suffix resolution (`Resolve`: a full ID or any unique suffix, in either case since Crockford base32 ignores it); `Tree.IDs` sorts IDs as strings, which is creation order for ULIDs. `ValidStamp` checks baseline names.
- `ulid.go`: canonical ULID validation/generation and project-name validation. PlanNew increments the newest ID for the same millisecond to preserve order; injected clock/entropy make tests deterministic. ID format was chosen in 01M485SBR0WN8JDQSN7A3GT7CB.
- `fsm.go`: `Gates`, `Plan`, `PlanNew`, `PlanEdit`, `Apply` and `CheckWrite` serve CLI and TUI. Completion permits non-parent decision/research issues from defined or in-progress, requiring a linked decision record with outcome (`G_OUTCOME`) or finding record (`G_FINDINGS`); criteria must be checked and requirements current. Code/manual issues require in-progress; parents require ready or in-progress and resolved children. Completion has no actor or documentation gate.
- `records.go`: `PlanRecord` writes context, criteria and log on unresolved issues. `ApplyAcceptance` applies operations against the original criterion numbering before appending additions. Decisions and findings are knowledge records rather than issue files.
- Priority: effective medium when unset; priority is outside the baseline. Query/next/prime use rank then ID, while Tree.IDs stays chronological. Priority-only edits work on resolved issues.
- Tags: `Issue.Tags`, `NormalizeTags`/`ValidTag` (lowercase, `. _ - /`), `Filter.Tags` (repeated values OR), validation code I026. Tags are outside the baselined requirement, so they never make an issue stale; `PlanEdit` allows a tag-only edit on resolved issues.
- `knowledge.go`: `KnowledgeRecord`, `PlanKnowledgeRecord`, `ResolveRecord`, `IssueRecords` and `HasRecord`. New decisions/findings require a ULID, title and one-line description, with optional issue, scopes, body, and decision outcome/supersedes. References accept record IDs or unique suffixes across the bundle. `DeriveSuperseded` marks older records referenced by newer ones. `Change.KnowledgeEntry` lets `CheckWrite` validate adapter-rendered fields and links. The KnowledgeStore port exposes Load, Render and Apply; records cannot be edited.
- `guide.go`: each `Pointer` in the Write list carries the `command` that writes it.
- `validate.go` and `diag.go`: diagnostics with stable codes (`P*` project, `I*` issues, `K*` knowledge), severity and class (fixable, guided, manual). Codes are never reused. Codes for file contents the domain does not parse (I017 format, I028 repeated heading) come from the store adapter. Resolved issues' knowledge references are history, unchecked; I025 covers open issues.

- `knowview.go`: `EntryIssues` and `IssueEntries` include linked records alongside existing context links and legacy resolution references. KnowledgeFilter supports type, status, scope and title text. A record without scopes applies project-wide.
- `config.go`: `PlanConfig` validates a whole project configuration (view names non-empty, trimmed and distinct; every view in `ViewOrder`; queries parse with `ParseFilter`; gate `G_CONFIG`) and returns `Change{Op: OpConfig, Config}` with queries whitespace-normalized; `Tree.Apply` swaps the project config, so `CheckWrite` sees it.
- `flags.go`: `Tree.QueryFlags` describes every `ParseFilter` flag in order (list, bool or text; description) with the values available now and issue counts: states, kinds, tags in use, priorities, parents for `--under` (ID, title, descendants), and counts for switches. `prep flags` prints it and the TUI filter bar's picker offers its values; `FlagKind(name)` tells readers of typed queries which flags take a value. `TestQueryFlagsMatchParseFilter` reads `ParseFilter`'s `switch name` cases from the source and fails when one is not described.
- `query.go`: `ParseFilter` (different flags AND, repeated flags OR; list values may use a leading `not-` or `!` to exclude; an issue matches when it has a plain value or there are none, and no excluded one; 01M4ANA516QWSG4BAQE0QCCR9F), `ValidTag` rejects tags starting with `not-`; values use hyphens (state is `in-progress`), `Query`, tree ordering, `ViewNames` (saved views in config order). `validateKnowledge` accepts OKF statuses draft, stable and deprecated.
- `guide.go`: `BuildGuide` step contracts and `Candidates` knowledge retrieval; `ScopeMatch` globbing.
- `knowread.go`: `Sections` and `Blocks` respect code fences; `FindKnowledge` ranks matching blocks by terms, title, description and scope, and includes derived superseded status. Searches match record title and description even when body words differ.
- `errors.go`: `E_*` error codes for commands.

Rules worth knowing: see [Issue lifecycle](/features/lifecycle.md).
- Completion behavior is described in [Completion requirements](/features/completion-policy.md). Definition of Done remains guidance. Knowledge starts empty and has no bootstrap, code-drift or confirmation machinery.
