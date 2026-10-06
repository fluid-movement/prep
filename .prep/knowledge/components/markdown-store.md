---
type: component
title: Markdown store
description: internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/mdstore
confirmed_commit: bf76661eace9892a8409fcaff992925b6eff8252
---

# Markdown store

Applies when changing file formats or write behavior. Format details: [Storage format](/conventions/storage-format.md).

- `Load` reads every file on every run (no cache) and records a SHA-256 per file. Issue directories must be ULIDs (`domain.ValidID`, else I001 naming the ULID shape); baseline files must be timestamps (`domain.ValidStamp`). Format-level problems become diagnostics (unknown files, frontmatter errors, non-canonical files, missing schema files).
- Writes go through a temp file plus rename; before writing, the current content hash must equal the one read at load (absent for new files), else `E_CONFLICT`.
- `requirementBody` builds the issue.md body from requirement text for `prep new` and `prep edit`: it keeps the text's own `## Open questions` section and appends an empty one only when missing. An edit rewrites issue.md from the edited fields under the same compare-and-swap.
- Record writes: context.md and findings.md are rewritten; decisions are appended by `renderDecision` (`date:` directly under the heading); acceptance operations are applied line by line by `applyAcceptance`, so text between criteria survives. Appends and line edits read the file under compare-and-swap first.
- issue.md frontmatter carries `priority` after `tags`; `renderIssue` omits medium (and empty) and keeps an unknown level for `prep check` to report.
- `LoadConfig` also records the order of the `views:` keys (`Config.ViewOrder`) from the YAML node; issue loading keeps `history.md` as `Issue.History`.
- `Apply` writes `Change.Config` to config.yaml with `renderConfig`: the comment lines leading the current file (or the default header), `commit_mode`, and `views` in `ViewOrder` through a `yaml.Node` mapping (an empty query as `""`), so an unchanged config reproduces `DefaultConfig` byte for byte. `DefaultConfig`, written by `prep init`, has commit mode off and two views: Unresolved (`--state open,defined,ready,in_progress`) and All (`""`).
- `Init` writes `project.md`, `config.yaml` and `.gitkeep` files for `issues/` and `knowledge/`; it no longer writes a placeholder overview, because the overview marks a bootstrapped knowledge base.
- `Fmt` re-renders every parseable file canonically (`--check` reports only); `Fix` adds missing empty schema files, removes duplicate dependencies and merges repeated `##` sections whose copies hold content at most once (`mergeDuplicateSections` in `sections.go` keeps that copy, or the first; `loadIssue` reports repeats as I028 through `duplicateHeadings`, guided when several copies have content).

- The contract corpus in `testdata/contract` (valid and broken trees with an `expect` file of diagnostic codes) runs in `contract_test.go` and is the test suite for any future adapter.
