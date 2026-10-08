---
type: component
title: Domain package
description: internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/domain
confirmed_commit: 6da3e44f576018f7d6ddd149a6492c1ae1c15183
---

# Domain package

Applies when changing lifecycle rules, gates, diagnostics or queries.

- `types.go`: `Issue`, records (`Baseline`, `Ready`, `Claim`, `Resolution`), `Entry`, kinds and states.
- `tree.go`: `Tree.State` derives state from records; `Stale`, `Blocked`, `Actionable`, `ChildProgress`, `EffectiveDoD`, ID suffix resolution (`Resolve`: a full ID or any unique suffix); `Tree.IDs` sorts IDs as strings, which is creation order for ULIDs. `ValidStamp` checks baseline names.
- `ulid.go`: issue IDs are ULIDs, implemented with the standard library (decided in 01M485SBR0WN8JDQSN7A3GT7CB). `ValidID` accepts only canonical ULIDs; `NewULID(now, entropy)` encodes the milliseconds and 80 bits from the reader (nil is `crypto/rand`); `ULIDTime` decodes the time. `PlanNew` calls `nextULID`, which increments the newest existing ID of the same millisecond when the new one would not sort after it, so IDs within one tree stay in creation order. `NewIssueInput.Entropy` and the `entropy` argument of `PlanBootstrap` inject the random source; tests seed it so IDs stay deterministic.
- `fsm.go`: `Gates` (unmet conditions with `G_*` codes; with nil input, argument-dependent gates become needs), `Plan` (records to write), `PlanNew`, `PlanEdit` (edits of title, kind, parent, dependencies, tags, priority and requirement on unresolved issues; tag and priority edits also on resolved ones; refuses dropped dependency targets and leaves existence and cycles to `CheckWrite`), `Apply` and `CheckWrite`.
- `records.go`: `PlanRecord` for the record writes (context, findings, decide, criterion, dod, log) on unresolved issues, with gates `G_CRITERION`, `G_DECISION`, `G_DOD` and `G_KIND`; the next decision ID is `D<max+1>`. Acceptance changes are `AcceptanceOp` operations, applied in memory by `ApplyAcceptance` (index operations refer to the numbering before the change, additions follow).
- `import.go`: `ImportTag`, `ImportOpen` (the unresolved import issue), `PlanImport` (one research issue "Import existing work" with a guiding context and three criteria, planned on a progressively applied tree like `PlanBootstrap`; `ErrInvalid` while an import issue is unresolved). `Filter.Text` matches the title or the requirement (`Issue.Body`).
- `bootstrap.go`: `Bootstrapped` (the overview entry exists), `BootstrapStart` (first open bootstrap-tagged leaf), `BootstrapAlert`, `PlanBootstrap` (parent and survey issues with requirement, context and criteria, planned on a progressively applied tree). `PlanKnowledge` refuses new entries other than the overview before bootstrapping (`G_BOOTSTRAP`); `Guide.Alerts` carries the alert.
- Priority: `Priority` (critical, high, medium, low; `Priorities` in rank order), `Issue.Priority` as stored (empty is medium, `Effective()` for output), `ParsePriority` (medium and empty unset), `Rank` (unknown levels rank as medium), `Filter.Priorities` (`--priority`, medium matches unset), validation code I027, gate `G_PRIORITY`. `Tree.ByPriority` orders IDs by rank then ID; `Query` returns its result through it and `TreeOrder` walks roots and siblings with it, while `Tree.IDs` stays chronological. Like tags, priority is outside the baselined requirement and a priority-only edit works on resolved issues (decided in 01M4874AR04JFPZ5KYW0402JC1).
- Tags: `Issue.Tags`, `NormalizeTags`/`ValidTag` (lowercase, `. _ - /`), `Filter.Tags` (repeated values OR), validation code I026. Tags are outside the baselined requirement, so they never make an issue stale; `PlanEdit` allows a tag-only edit on resolved issues.
- `knowledge.go`: `KnowledgeEdit` and `Tree.PlanKnowledge` (path form, no rename, existence, non-empty body, required fields for new entries, scope required for `confirmed_commit`); `Change.KnowledgeEntry` carries the adapter-rendered entry so `Tree.Apply` and `CheckWrite` validate it. The `KnowledgeStore` port has `Load`, `Render` and `Apply`.
- `guide.go`: each `Pointer` in the Write list carries the `command` that writes it.
- `validate.go` and `diag.go`: diagnostics with stable codes (`P*` project, `I*` issues, `K*` knowledge), severity and class (fixable, guided, manual). Codes are never reused. Codes for file contents the domain does not parse (I017 canonical format, I028 repeated section heading) come from the store adapter.

- `knowview.go`: `EntryIssues` (issues whose documentation decision names an entry, newest first), `IssueEntries` (an issue's context links plus, once resolved, its documented entries; existing ones, deduplicated), `KnowledgeFilter`/`ParseKnowledgeFilter` (`--type`, `--status`, `--scope` matched both ways with `ScopeMatch`, bare words as a title phrase) and `QueryKnowledge` (paths in order). Used by the TUI's knowledge screen; a CLI listing can share them.
- `config.go`: `PlanConfig` validates a whole project configuration (view names non-empty, trimmed and distinct; every view in `ViewOrder`; queries parse with `ParseFilter`; gate `G_CONFIG`) and returns `Change{Op: OpConfig, Config}` with queries whitespace-normalized; `Tree.Apply` swaps the project config, so `CheckWrite` sees it.
- `flags.go`: `Tree.QueryFlags` describes every `ParseFilter` flag in order (list, bool or text; description) with the values available now and issue counts: states, kinds, tags in use, priorities, parents for `--under` (ID, title, descendants), and counts for the switches. `prep flags` prints it and the TUI filter bar's picker offers its values; `FlagKind(name)` tells readers of typed queries which flags take a value. `TestQueryFlagsMatchParseFilter` reads `ParseFilter`'s `switch name` cases from the source and fails when one is not described.
- `query.go`: `ParseFilter` (different flags AND, repeated flags OR; a list flag's value with a leading `not-` or `!` (`NegationPrefixes`) goes to the `Not*` fields and excludes: an issue matches when it has a plain value, or there are none, and no excluded one; 01M4ANA516QWSG4BAQE0QCCR9F), `ValidTag` rejects tags starting with `not-` so negation is unambiguous; values use hyphens (the state is `in-progress`, the underscore spelling is not accepted), `Query`, tree ordering, `ViewNames` (saved views in config order). `validateKnowledge` accepts the OKF statuses draft, stable and deprecated.
- `guide.go`: `BuildGuide` step contracts and `Candidates` knowledge retrieval; `ScopeMatch` globbing.
- `knowread.go`: `Sections` (an entry body split at headings outside code fences; anchors are `Slug`s with -1, -2 for repeats; `Text` includes subsections, `Own` stops at the next heading), `Blocks` (paragraphs and top-level list items with what they nest; a fence stays whole) and `Tree.FindKnowledge` (blocks ranked by terms matched in the block or its heading, then score: exact, prefix, substring, one typo; entry title, description and scope add score only).
- `errors.go`: `E_*` error codes for commands.

Rules worth knowing: see [Issue lifecycle](/features/lifecycle.md).
