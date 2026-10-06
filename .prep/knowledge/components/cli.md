---
type: component
title: CLI
description: internal/cli — command table, global flags, JSON output and errors, write pipeline, git staging and commit modes.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/cli
  - cmd/prep
confirmed_commit: e3437dad10f299c56025845594f0874fad99a668
---

# CLI

Applies when adding or changing commands.

- Commands are registered in `internal/cli/cli.go` with a read/write flag; read and write commands are strictly separate.
- Global flags anywhere: `--json`, `--root`, `--by` (or `PREP_ACTOR`, `PREP_ROOT`). The default actor is `cli/prep-<version>`; actors starting with `human:` cannot complete code issues.
- Under `--json` everything, including errors (`{ok:false,error:{code,message,unmet,diagnostics}}`), is JSON on stdout. Exit codes: 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
- `prep edit <id>` changes any of `--title`, `--kind`, `--parent` (`''` removes it), `--depends-on` (repeatable, replaces the list; `''` clears it) and `--body`/`--body-file`. It uses `flag.Visit` to tell unset from empty, and reports `state` and `stale` after the write.
- Record writes (`recordCmd` in `write.go`): `context` and `findings` replace their record and require `--body` or `--body-file`; `decide` appends a decision; `criterion` (`--add`, `--check`/`--uncheck`/`--remove <n>`) and `dod` (`--add`, `--opt-out` with `--reason`, `--remove <item>`) change acceptance.md; `log <id> <text>` appends to history. `prep show` numbers criteria for `--check <n>`.
- `prep fmt` (and `--check`) and `prep fix` also regenerate the OKF index files of the knowledge bundle.
- `record` stages written paths or commits them in commit mode `all`; `afterWrite` prints its errors, the TUI's write function returns them.
- `prep knowledge new|update|confirm` (`knowledge.go`): knowledge entry writes run plan (`Tree.PlanKnowledge`), render through the store, `CheckWrite`, write, stage; `confirm` sets `confirmed_commit` to HEAD for the named entries or `--drifted` ones and needs git and a scope.
- Commit mode defaults to `off`: prep stages the `.prep` files it writes and the user or agent commits them with the code; `all` commits each operation (decided in 01M46ARS5RNN4X2TMVHAKWA3KW).
- Write pipeline: load → `Plan` → `CheckWrite` → `store.Apply` → stage (commit mode `off`) or commit only `.prep` paths (`all`).
- Priority: `prep new --priority` and `prep edit --priority` (critical, high, medium, low; medium unsets; a priority-only edit works on resolved issues), `prep list --priority` filters (comma-separated OR, also in saved views). `summary` and `prep show --json` always carry the effective `priority`; text output marks non-medium levels before the title (`!crit`, `!high`, `low`) and `show` prints `priority:`. `prep list`, `prep next` and prime's actionable and stale lists order by priority then ID (`Tree.ByPriority`, `byPriority` for prime summaries).
- Tags: `prep new --tag` and `prep edit --tag` (repeatable or comma-separated; edit replaces the list, `--tag ''` clears it, and a tag-only edit also works on resolved issues); `prep list --tag` filters; `prep list` shows tags as `#tag` after the title and `prep show` as a `tags:` line.
- `prep init [--no-bootstrap]` creates `.prep` (no placeholder overview) and, unless skipped, the bootstrap issues via `Tree.PlanBootstrap`, written change by change; `prep knowledge bootstrap` does the same later when the knowledge base is not bootstrapped and no bootstrap issue is open. `prep prime` (JSON `bootstrap`) and `prep guide` (`alerts`) print the bootstrap alert first.
- `prep setup [--harness h,...|--remove h|--refresh]` installs harness integrations; see [Harness setup](/components/setup.md).
- `prep update [--check]` replaces the binary with the latest release; see [Release, install and update](/components/release.md).
- `prep watch` (`watch.go`) is a read command for harness integrations: it watches `.prep` with `internal/watch` and prints one line per debounced change (`changed`, or `{"event":"changed"}` with `--json`) until SIGINT/SIGTERM or until its output closes; tests stop it by replacing `watchContext`. The [Claude Code integration](/components/claude-code.md)'s panel refreshes on each line.
- `prep show --json` carries the issue's fields plus `history` (the work log text), `context`, `findings`, derived state, relations, progress and the effective Definition of Done.
- `prep tui` is a read command that needs a terminal (stdin and stdout); it opens the [TUI](/components/tui.md), and `--gallery` shows the [TUI design system](/components/tui-design-system.md). `prep views` lists saved views in config order.
- `prep skill` prints the embedded `internal/cli/skill.md`; `.claude/skills/prep/SKILL.md` must stay identical (tested).
- Tests in `cli_test.go` drive whole lifecycles through `Main` with a fake clock; the package variables `clock` and `entropy` (nil: `crypto/rand`) feed `prep new` and the bootstrap issues their IDs, and tests read IDs from the commands' output.
