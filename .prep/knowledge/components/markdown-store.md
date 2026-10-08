---
type: component
title: Markdown store
description: internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/mdstore
confirmed_commit: 6a1a84958c37cb18db880dbb4f937214482a1f07
---

# Markdown store

Applies when changing file formats or write behavior. Format details: [Storage format](/conventions/storage-format.md).

- `Load` reads every file on every run (no cache) and records a SHA-256 per file. Issue directories must be ULIDs (`domain.ValidID`, else I001 naming the ULID shape); baseline files must be timestamps (`domain.ValidStamp`). Format-level problems become diagnostics (unknown files, frontmatter errors, non-canonical files, missing schema files). Names starting with `tempPrefix` (`.prep-tmp-`, the temp files of atomic writes) are skipped in `.prep` and issue directories: one may belong to a write in flight or be left by a crash.
- Writes go through a temp file plus rename; before writing, the current content hash must equal the one read at load (absent for new files), else `E_CONFLICT`. `Apply` first renders the whole change (`render`: paths and contents, nil content removes a file) and checks every path, then writes, so a conflict leaves the change unwritten rather than half applied; only an I/O error midway can leave part of it.
- `requirementBody` builds the issue.md body from requirement text for `prep new` and `prep edit`: it keeps the text's own `## Open questions` section and appends an empty one only when missing. An edit rewrites issue.md from the edited fields under the same compare-and-swap.
- Record writes: context.md and findings.md are rewritten; decisions are appended by `renderDecision` (`date:` directly under the heading); acceptance operations are applied line by line by `applyAcceptance`, so text between criteria survives. Appends and line edits read the file under compare-and-swap first.
- issue.md frontmatter carries `priority` after `tags`; `renderIssue` omits medium (and empty) and keeps an unknown level for `prep check` to report.
- `LoadConfig` also records the order of the `views:` keys (`Config.ViewOrder`) from the YAML node; issue loading keeps `history.md` as `Issue.History`.
- Config files: `LoadConfig` returns the config in effect, the user's own `config.yaml` (gitignored) or else the committed `config.yaml.dist` (constants `ConfigOwn`, `ConfigDist`), as a whole and never merged; with neither, defaults. `Load` validates both files with `loadConfigFile` and reports each invalid one as P003 under its own name (the message without the file prefix); a config in the prep 0.1.0 layout (`commit_mode`, `in_progress` in a `--state` value (other text containing it stays), own config without `config.yaml.dist`, no `.gitignore` entry) is read converted in memory (`convertLegacyConfig`, so views keep working) and reported as P006, fixable: `Fix` (`fixLegacyConfig`) writes both files converted, shares `config.yaml.dist` from the own config and adds the ignore entry, and `cmdFix` takes `config.yaml` out of git's index (01M4AVPGN5W7WZH6TQ5EKJG5KN); themes are validated with `palette.Validate` (unknown theme, base, token, invalid color, cycle).
- `Apply` writes `Change.Config` to the own config.yaml with `renderConfig`, starting from config.yaml.dist when there is no own file yet, after `palette.Validate` (an invalid theme is refused as E_USAGE): the comment lines leading the current file (or the default header) and `views` in `ViewOrder` through a `yaml.Node` mapping (an empty query as `""`; no views writes only the header), then `theme` and `themes` (custom themes by `palette.Names`, tokens in `palette.Tokens` order, values double-quoted), so an unchanged config reproduces `DefaultConfig` byte for byte. `DefaultConfig`, written by `prep init` as config.yaml.dist, has a header explaining the dist/own split and views, and two views: Unresolved (`--state open,defined,ready,in-progress`) and All (`""`).
- `Init` writes `project.md`, `config.yaml.dist`, `.gitignore` (ignoring `config.yaml` and `local/`; `Load` reports P007 when `.prep/local` exists unignored and `Fix` adds the line; `local` is a known entry of `.prep`) and `.gitkeep` files for `issues/` and `knowledge/`; it no longer writes a placeholder overview, because the overview marks a bootstrapped knowledge base.
- `Fmt` re-renders every parseable file canonically (`--check` reports only); `Fix` adds missing empty schema files, removes duplicate dependencies and merges repeated `##` sections whose copies hold content at most once (`mergeDuplicateSections` in `sections.go` keeps that copy, or the first; `loadIssue` reports repeats as I028 through `duplicateHeadings`, guided when several copies have content).

- Code fences: every parser that looks for headings or keeps code verbatim (`normalize`, `splitSections`, `splitOpenQuestions`, `hasOpenQuestions`, `parseDecisions`, `canonicalDecisions`, and `domain.Sections`/`Blocks`) follows blocks with `domain.Fence`, which closes a block only on a run of the same character at least as long as the opening one; a heading inside a fence is never a heading.
- The contract corpus in `testdata/contract` (valid and broken trees with an `expect` file of diagnostic codes) runs in `contract_test.go` and is the test suite for any future adapter.
