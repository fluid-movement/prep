---
type: component
title: Markdown store
description: internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/mdstore
confirmed_commit: d902990b28eaba324184b9331bc0b3b243104443
---

# Markdown store

Applies when changing file formats or write behavior. Format details: [Storage format](/conventions/storage-format.md).

- `Load` reads every file on every run (no cache) and records a SHA-256 per file. Format-level problems become diagnostics (unknown files, frontmatter errors, non-canonical files, missing schema files).
- Writes go through a temp file plus rename; before writing, the current content hash must equal the one read at load (absent for new files), else `E_CONFLICT`.
- `Fmt` re-renders every parseable file canonically (`--check` reports only); `Fix` adds missing empty schema files and removes duplicate dependencies.
- The contract corpus in `testdata/contract` (valid and broken trees with an `expect` file of diagnostic codes) runs in `contract_test.go` and is the test suite for any future adapter.
