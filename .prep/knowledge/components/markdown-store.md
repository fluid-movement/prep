---
type: component
title: Markdown store
description: internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope: internal/mdstore
confirmed_commit: 8030cc0c6ca5ae85dd8d4b1e186d0aab1b963960
---

# Markdown store

Applies when changing file formats or write behavior. Format details: [Storage format](/conventions/storage-format.md). Optional artifact references round-trip as issue metadata: [CLI contracts](/features/cli-contracts.md).

- `Load` takes the checkout lock, recovers unfinished Apply transactions, resets its snapshot and reads every file (no cache), recording SHA-256 hashes and relevant directory-entry manifests, including issue and baseline membership. Issue directories must be ULIDs (`domain.ValidID`, else I001 naming the ULID shape); baseline files must be timestamps (`domain.ValidStamp`). Format-level problems become diagnostics (unknown files, frontmatter errors, non-canonical files, missing schema files). `splitFrontmatter` tolerates a leading BOM, which `prep fmt` then drops. Names starting with `tempPrefix` (`.prep-tmp-`, the temp files of atomic writes) are skipped in `.prep` and issue directories: one may belong to a write in flight or be left by a crash.
- Writes use checkout-local advisory serialization and per-file atomic rename. `Apply` validates frozen loaded record hashes and loaded directory-entry manifests, then stages a recoverable multi-file journal; stale plans return E_CONFLICT before mutation. Locked loads recover by rolling forward, or retain conflicting data and fail visibly. Maintenance writes remain per-file CAS operations. See [Local write safety](/features/local-write-safety.md) for protocol, recovery and limits.
- `requirementBody` builds the issue.md body from requirement text for `prep new` and `prep edit`: it keeps the text’s own `## Open questions` section and appends an empty one only when missing. An edit rewrites issue.md from the edited fields under the same compare-and-swap.
- Record writes: context.md and findings.md are rewritten; decisions are appended by `renderDecision` (`date:` directly under the heading); acceptance operations are applied line by line by `applyAcceptance`, so text between criteria survives. Appends and line edits read the file under compare-and-swap first.
- issue.md frontmatter carries `priority` after `tags`; `renderIssue` omits medium (and empty) and keeps an unknown level for `prep check` to report.
- `LoadConfig` also records the order of the `views:` keys (`Config.ViewOrder`) from the YAML node; issue loading keeps `history.md` as `Issue.History`.
- Config files: `LoadConfig` returns the config in effect, the user’s own `config.yaml` (gitignored) or else committed `config.yaml.dist`, as a whole and never merged; with neither, defaults. `Load` validates both files with `loadConfigFile` and reports each invalid one as P003 under its own name (the message without the file prefix); a config in the prep 0.1.0 layout (`commit_mode`, `in_progress` in a `--state` value (other text containing it stays), own config without `config.yaml.dist`, no `.gitignore` entry) is read converted in memory (`convertLegacyConfig`, so views keep working) and reported as P006, fixable: `Fix` writes both files converted, shares `config.yaml.dist` from the own config and adds the ignore entry, and `cmdFix` takes `config.yaml` out of git’s index; theme names are validated with `palette.Validate` (only `default` and `light` are accepted).
- `Apply` writes `Change.Config` to the own config.yaml with `renderConfig`, starting from config.yaml.dist when there is no own file yet, after `palette.Validate`: leading comment lines (or default header) and `views` in `ViewOrder` through a `yaml.Node` mapping (an empty query as `""`; no views writes only the header), then `theme`, so unchanged config reproduces `DefaultConfig` byte for byte. `DefaultConfig`, written by `prep init` as config.yaml.dist, has a header explaining the dist/own split and views, and two views: Unresolved (`--state open,defined,ready,in-progress`) and All (`""`).
- `Init(id, name)` writes `project.md` (with identity computed by the CLI; mdstore never calls git), `config.yaml.dist`, `.gitignore` (ignoring `config.yaml` and `local/`) and `.gitkeep` files for `issues/` and `knowledge/`; it no longer writes a placeholder overview, because the overview marks a bootstrapped knowledge base.
- `Fmt` re-renders every parseable file canonically (`--check` reports only); `Fix` adds missing empty schema files, removes duplicate dependencies and merges repeated `##` sections whose copies hold content at most once. Headings inside code fences are ignored.
- The contract corpus in `testdata/contract` (valid and broken trees with an `expect` file of diagnostic codes) runs in `contract_test.go` and is the test suite for any future adapter.
- Project identity: `parseProject` reads `id` and `name` into `domain.Project`; `ReadProject(root)` parses one project.md without opening a store (the [project registry](/components/projects.md) uses it). `SetSchema` and `SetIdentity` rewrite frontmatter through `editProject`, keeping the body. P008 (no id) and P009 (invalid id or name) come from `domain.Validate`.

Issue completion records contain ready.md, claim.md and resolution.md. Resolution records store the outcome, actor and time, reason, optional evidence, and documentation decision (`entries` or `no_impact`). See [Storage format](/conventions/storage-format.md). Reopen appends the removed resolution.md, fenced, to history.md after the reopen line.
