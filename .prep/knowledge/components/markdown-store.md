---
type: component
title: Markdown store
description: internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/mdstore
confirmed_commit: 0a4f01a06a9df5c42a1051c92986c8d2ee2731ba
---

# Markdown store

Applies when changing file formats or write behavior. Format details: [Storage format](/conventions/storage-format.md).

- `Load` reads every file on every run (no cache) and records a SHA-256 per file. Format-level problems become diagnostics (unknown files, frontmatter errors, non-canonical files, missing schema files).
- Writes go through a temp file plus rename; before writing, the current content hash must equal the one read at load (absent for new files), else `E_CONFLICT`.
- `requirementBody` builds the issue.md body from requirement text for `prep new` and `prep edit`: it keeps the text's own `## Open questions` section and appends an empty one only when missing. An edit rewrites issue.md from the edited fields under the same compare-and-swap.
- Record writes: context.md and findings.md are rewritten; decisions are appended by `renderDecision` (`date:` directly under the heading); acceptance operations are applied line by line by `applyAcceptance`, so text between criteria survives. Appends and line edits read the file under compare-and-swap first.
- `LoadConfig` also records the order of the `views:` keys (`Config.ViewOrder`) from the YAML node; issue loading keeps `history.md` as `Issue.History`.
- `Init` writes `project.md`, `config.yaml` and `.gitkeep` files for `issues/` and `knowledge/`; it no longer writes a placeholder overview, because the overview marks a bootstrapped knowledge base.
- `Fmt` re-renders every parseable file canonically (`--check` reports only); `Fix` adds missing empty schema files and removes duplicate dependencies.
- The contract corpus in `testdata/contract` (valid and broken trees with an `expect` file of diagnostic codes) runs in `contract_test.go` and is the test suite for any future adapter.
