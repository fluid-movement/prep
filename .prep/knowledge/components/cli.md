---
type: component
title: CLI
description: internal/cli — command table, global flags, JSON output and errors, write pipeline and git staging.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/cli
  - cmd/prep
confirmed_commit: d73c7a0eab6c048393bb0c2f9462fe34e3c655c7
---

# CLI

- Commands are registered in `internal/cli/cli.go` with a read/write flag; read and write commands are strictly separate.
- Global flags anywhere before a `--`: `--json` (`--json=false` turns it off), `--root`, `--project` ([Cross-project work](/features/cross-project.md)), `--by` (or `PREP_ACTOR`, `PREP_ROOT`, `PREP_PROJECT`); from `--` on, arguments are passed as written and `parse` takes them all as positionals, so a text can be `--json`. The default actor is `cli/prep-<version>`; actors starting with `human:` cannot complete code issues.
- Under `--json` command output, including errors (`{ok:false,error:{code,message,unmet,diagnostics}}`), is JSON on stdout; `help`, `skill`, `watch`, `prime --hook` and `activity add` (silent) print as they always do. Exit codes: 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
- `prep edit <id>` changes any of `--title`, `--kind`, `--parent` (`''` removes it), `--depends-on` (repeatable, replaces the list; `''` clears it) and `--body`/`--body-file`. It uses `flag.Visit` to tell unset from empty, and reports `state` and `stale` after the write.
- Record writes (`recordCmd` in `write.go`): `context` and `findings` replace their record and require `--body` or `--body-file`; `decide` appends a decision; `criterion` (`--add`, `--check`/`--uncheck`/`--remove <n>`) and `dod` (`--add`, `--remove <item>`) change acceptance.md; `log <id> <text>` appends to history. `prep show` numbers criteria for `--check <n>`. `prep dod --add <item>` and `--remove <item>` edit the project-level Definition of Done.
- `prep fmt` (and `--check`) and `prep fix` also regenerate the OKF index files of the knowledge bundle.
- `record` stages written paths except those git ignores (`gitx.Tracked`, so the user's own `.prep/config.yaml` is never staged); `afterWrite` prints its errors, the TUI's write function returns them. prep never commits: the user or agent commits the staged `.prep` files with the code.
- `prep knowledge list|find|show` (`knowread.go`) are reads: `list` takes the knowledge filter, `find <query> [--max N]` prints matching blocks grouped by entry, `show <entry>[#anchor]... [--outline]` accepts several references in the forms `NormalizeEntryPath` takes; an unknown anchor fails naming the anchors. `prep guide` lists its knowledge candidates as entry paths to read with `prep knowledge show`.
- `prep knowledge new|update|confirm` (`knowledge.go`): knowledge entry writes run plan (`Tree.PlanKnowledge`), render through the store, `CheckWrite`, write, stage; `confirm` sets `confirmed_commit` to HEAD for the named entries or `--drifted` ones and needs git and a scope.

- Write pipeline: load → `Plan` → `CheckWrite` → `store.Apply` → stage.
- Priority: new/edit accept critical, high, medium or low; medium unsets. Priority-only edits work on resolved issues. list/next/prime order by priority then ID; JSON includes effective priority and text marks non-medium values.
- Tags: `prep new --tag` and `prep edit --tag` (repeatable or comma-separated; edit replaces the list, `--tag ''` clears it, and a tag-only edit also works on resolved issues); `prep list --tag` filters; `prep list` shows tags as `#tag` after the title and `prep show` as a `tags:` line.
- `prep init [--no-bootstrap]` creates `.prep` (no placeholder overview) and, unless skipped, the bootstrap issues via `Tree.PlanBootstrap`, written change by change; `prep knowledge bootstrap` does the same later when the knowledge base is not bootstrapped and no bootstrap issue is open. `prep prime` (JSON `bootstrap`) and `prep guide` (`alerts`) print the bootstrap alert first.
- `prep import` creates the "Import existing work" research issue via `Tree.PlanImport`, written change by change like the bootstrap (`applyAll`), and reports it as open with `next: prep guide <id>` (text: a `Next:` line); while an import issue is unresolved it fails with `E_INVALID` naming that issue. Sources are project-specific, so there are no per-source importers: the issue's context guides the agent (decided in 01M47YPC2G69C52S5FWP6H42PH).
- `prep list --text` (and `--text` in views and the TUI filter) matches a case-insensitive substring of the title or the requirement, so an imported issue is found by its `Source:` line.
- `prep setup [--harness h,...|--remove h|--refresh]` installs harness integrations; see [Harness setup](/components/setup.md).
- `prep activity`, `activity add` and `prep focus`: [Agent activity](/features/agent-activity.md); commands note their event when they succeed and `Main` appends it.
- `prep update [--check]` replaces the binary with the latest release; see [Release, install and update](/components/release.md).
- `prep show --json` carries the issue's fields plus `history` (the work log text), `context`, `findings`, derived state, relations and progress.
- `prep projects`, `init`'s identity and `--new-id`: [project registry](/components/projects.md); `prep links`, `prep link`, `validate` (check plus registry checks): [Cross-project work](/features/cross-project.md).

- `prep tui` requires a terminal; `--agent` opens the Agent screen and `--gallery` shows the [design system](/components/tui-design-system.md) with project themes. `prep flags` lists query values/counts, including saved views and tree mode; `prep views` lists saved views in config order.
- `prep skill` prints the embedded `internal/cli/skill.md`; `.claude/skills/prep/SKILL.md` must stay identical (tested).
- **Completion**: the transition checks issue state and staleness, checked criteria, the kind's result and a documentation decision. The project Definition of Done is guidance shown in the briefing and `prep guide`, not a gate; see [Completion requirements](/features/completion-policy.md).
- **Eval**: test seams `PREP_TEST_NOW`, `PREP_TEST_SEED` (`testSeams`); [prep eval](/components/eval.md).
`prep new` returns bounded advisory `similar_issues` candidates when titles overlap unresolved records; terminal/self records are omitted. Hints never prevent creation or choose a match. See [Tracking](/decisions/tracking.md).

- `prep reopen <id> --reason` moves a done resolution into the work log. See [Issue lifecycle](/features/lifecycle.md).
`prep dependency show`: [source observations](/features/cross-project.md).

Artifact metadata and additive lifecycle JSON/error guidance: [CLI contracts](/features/cli-contracts.md).
