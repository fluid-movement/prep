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
confirmed_commit: acc7486bb5972c218f85e206ca2eafae31736992
---

# CLI

- Commands are registered in `internal/cli/cli.go` with a read/write flag; read and write commands are strictly separate.
- Global flags anywhere before a `--`: `--json` (`--json=false` turns it off), `--root`, `--project` ([Cross-project work](/features/cross-project.md)), `--by` (or `PREP_ACTOR`, `PREP_ROOT`, `PREP_PROJECT`); from `--` on, arguments are passed as written and `parse` takes them all as positionals, so a text can be `--json`. The default actor is `cli/prep-<version>`; any actor, a `human:` one included, can complete any issue.
- Under `--json` command output, including errors (`{ok:false,error:{code,message,unmet,diagnostics}}`), is JSON on stdout; `help`, `skill`, `watch`, `prime --hook` and `activity add` (silent) print as they always do. Exit codes: 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
- `prep edit <id>` changes any of `--title`, `--kind`, `--parent` (`''` removes it), `--depends-on` (repeatable, replaces the list; `''` clears it) and `--body`/`--body-file`. It uses `flag.Visit` to tell unset from empty, and reports `state` and `stale` after the write.
- Record writes (`recordCmd` in `write.go`): `context` replaces its record and requires `--body` or `--body-file`; `criterion` changes acceptance.md, `dod` changes project Definition of Done guidance, and `log` appends history.md. `prep show` numbers criteria. `prep decide [<issue>] --title T --description D` and `prep finding [<issue>] --title T --description D` use the OKF write pipeline to create immutable ULID-named knowledge records; body and repeatable scopes are optional. Decisions support repeatable `--supersedes <record-id-or-suffix>` and `--outcome`.
- `prep fmt` (and `--check`) and `prep fix` also regenerate the OKF index files of the knowledge bundle.
- `record` stages written paths except those git ignores (`gitx.Tracked`, so the user's own `.prep/config.yaml` is never staged); `afterWrite` prints its errors, the TUI's write function returns them. prep never commits: the user or agent commits the staged `.prep` files with the code.
- `prep knowledge list|find|show` (`knowread.go`) are reads: `list` takes the knowledge filter, `find <query> [--max N]` prints matching blocks grouped by entry, `show <entry>[#anchor]... [--outline]` accepts several references in the forms `NormalizeEntryPath` takes; an unknown anchor fails naming the anchors. `prep guide` lists its knowledge candidates as entry paths to read with `prep knowledge show`.
- Knowledge writes (`knowledge.go`): `knowledgeRecordCmd` runs `PlanKnowledgeRecord`, renders through the OKF store, checks the prospective tree, applies and stages. There are no mutable entry or confirmation commands. List, find and show expose derived superseded status; existing descriptive entries remain readable.

- Write pipeline: lock → load → `Plan` → `CheckWrite` → `store.Apply` → stage. Each write command calls `lockWrite` before its load and defers the release, so the checkout lock covers load through the last write and `load` reuses the store that holds it.
- Priority: new/edit accept critical, high, medium or low; medium unsets. Priority-only edits work on resolved issues. list/next/prime order by priority then ID; JSON includes effective priority and text marks non-medium values.
- Tags: `prep new --tag` and `prep edit --tag` (repeatable or comma-separated; edit replaces the list, `--tag ''` clears it, and a tag-only edit also works on resolved issues); `prep list --tag` filters; `prep list` shows tags as `#tag` after the title and `prep show` as a `tags:` line.
- `prep init` creates `.prep` with empty knowledge and no bootstrap issues or prerequisite overview. It writes union-merge attributes for history.md and knowledge indexes.
- `prep list --text` (and `--text` in views and the TUI filter) matches a case-insensitive substring of the title or the requirement.
- `prep setup [--harness h,...|--remove h|--refresh]` installs harness integrations; see [Harness setup](/components/setup.md).
- `prep activity`, `activity add` and `prep focus`: [Agent activity](/features/agent-activity.md); commands note their event when they succeed and `Main` appends it.
- `prep update [--check]` replaces the binary with the latest release; see [Release, install and update](/components/release.md).
- `prep show --json` carries issue fields plus history, context, derived state, relations, progress, and `knowledge` containing linked records. Show and guide list linked records by title.
- `prep projects`, `init`'s identity and `--new-id`: [project registry](/components/projects.md).

- `prep tui` requires a terminal; `--agent` opens the Agent screen and `--gallery` shows the [design system](/components/tui-design-system.md) with project themes. `prep flags` lists query values/counts, including saved views and tree mode; `prep views` lists saved views in config order.
- `prep skill` prints the embedded `internal/cli/skill.md`; `.claude/skills/prep/SKILL.md` must stay identical (tested).
- **Completion** checks issue state, staleness, checked criteria and the kind result: a linked decision record with outcome, or a linked finding record. There are no documentation flags or completion gate. The project Definition of Done remains guidance; see [Completion requirements](/features/completion-policy.md).
- **Eval**: test seams `PREP_TEST_NOW`, `PREP_TEST_SEED` (`testSeams`); [prep eval](/components/eval.md).
`prep new` returns bounded advisory `similar_issues` candidates when titles overlap unresolved records; terminal/self records are omitted. Hints never prevent creation or choose a match. See [Tracking](/decisions/tracking.md).

- `prep reopen <id> --reason` moves a done resolution into the work log. See [Issue lifecycle](/features/lifecycle.md).

Additive lifecycle JSON/error guidance: [CLI contracts](/features/cli-contracts.md).
