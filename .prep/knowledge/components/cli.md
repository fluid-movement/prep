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
confirmed_commit: 33d632fd0a3a2103b2fec6cedd2b47d7e8d5710d
---

# CLI

Applies when adding or changing commands.

- Commands are registered in `internal/cli/cli.go` with a read/write flag; read and write commands are strictly separate.
- Global flags anywhere before a `--`: `--json` (`--json=false` turns it off), `--root`, `--project` ([Cross-project work](/features/cross-project.md)), `--by` (or `PREP_ACTOR`, `PREP_ROOT`, `PREP_PROJECT`); from `--` on, arguments are passed as written and `parse` takes them all as positionals, so a text can be `--json`. The default actor is `cli/prep-<version>`; actors starting with `human:` cannot complete code issues.
- Under `--json` command output, including errors (`{ok:false,error:{code,message,unmet,diagnostics}}`), is JSON on stdout; `help`, `skill`, `watch`, `prime --hook` and `activity add` (silent) print as they always do. Exit codes: 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
- `prep edit <id>` changes any of `--title`, `--kind`, `--parent` (`''` removes it), `--depends-on` (repeatable, replaces the list; `''` clears it) and `--body`/`--body-file`. It uses `flag.Visit` to tell unset from empty, and reports `state` and `stale` after the write.
- Record writes (`recordCmd` in `write.go`): `context` and `findings` replace their record and require `--body` or `--body-file`; `decide` appends a decision; `criterion` (`--add`, `--check`/`--uncheck`/`--remove <n>`) and `dod` (`--add`, `--opt-out` with `--reason`, `--remove <item>`) change acceptance.md; `log <id> <text>` appends to history. `prep show` numbers criteria for `--check <n>`.
- `prep fmt` (and `--check`) and `prep fix` also regenerate the OKF index files of the knowledge bundle.
- `record` stages written paths except those git ignores (`gitx.Tracked`, so the user's own `.prep/config.yaml` is never staged); `afterWrite` prints its errors, the TUI's write function returns them. prep never commits: the user or agent commits the staged `.prep` files with the code (commit mode was removed in 01M4AJ9DN4CK5S4WX2N68VBR98).
- `prep knowledge list|find|show` (`knowread.go`) are reads, listed under read commands by a second `knowledge` row in the command table (dispatch takes the first; both run `cmdKnowledge`): `list` takes the knowledge filter, `find <query> [--max N]` prints matching blocks grouped by entry, `show <entry>[#anchor]... [--outline]` accepts several references in the forms `NormalizeEntryPath` takes; an unknown anchor fails naming the anchors. `prep guide` lists its knowledge candidates as entry paths to read with `prep knowledge show`.
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
- `prep watch` (`watch.go`, read, for harnesses and the [desktop app](/components/desktop.md)) watches `.prep` with `internal/watch` and prints one line per debounced project change (`changed`, `{"event":"changed"}` with `--json`; not the activity stream under `.prep/local`) until SIGINT/SIGTERM, its output closes, or its parent process is gone (`endWithParent`); tests replace `watchContext`.
- `prep show --json` carries the issue's fields plus `history` (the work log text), `context`, `findings`, derived state, relations, progress and the effective Definition of Done.
- `prep projects`, `init`'s identity and `--new-id`: [project registry](/components/projects.md); `prep links`, `prep link`, `validate` (check plus registry checks): [Cross-project work](/features/cross-project.md).

- `prep tui` requires a terminal; `--agent` opens the Agent screen and `--gallery` shows the [design system](/components/tui-design-system.md) with project themes. `prep theme list|new` lists or creates/selects themes; see [Palette](/components/palette.md). `prep flags` lists query values/counts, including saved views and tree mode; `prep views` lists saved views in config order.
- `prep skill` prints the embedded `internal/cli/skill.md`; `.claude/skills/prep/SKILL.md` must stay identical (tested).
- **Completion evidence**: a code issue completed without `--commit` has no evidence in `resolution.md`; `resolveEvidence` (used by `prep show` and the guide's touched files) fills it in memory with the commit that added `resolution.md` (`gitx.AddedIn`), and `prep show` marks it pending (`evidence_pending` in JSON) until that commit exists. A given `--commit` is validated as a hash and remains optional. Policy-bound issues also require passing applicable reports under `G_EVIDENCE`, independently of Git references. `prep evidence record/show` records results and inspects applicability; `prep policy show/set/adopt/retain` manages completion methods and bindings; `prep exception approve/show` and `prep dod approve` record explicit human approval.  See [Completion policy](/features/completion-policy.md).
- **Eval**: test seams `PREP_TEST_NOW`, `PREP_TEST_SEED` (`testSeams`); [prep eval](/components/eval.md).
`prep new` returns bounded advisory `similar_issues` candidates when titles overlap unresolved records; terminal/self records are omitted. Hints never prevent creation or choose a match. See [Tracking](/decisions/tracking.md).
