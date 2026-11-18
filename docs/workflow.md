# Workflow

[Documentation](../README.md#documentation)

Working alone, follow `prep guide <id>` through implementation and a separate check pass. To delegate, run `prep brief <id> --out` and give the implementer only this one-line prompt: "Read .prep/local/handoff/<id>/brief.md and follow it exactly. It tells you where to write your report." Run `prep handback <id>` to import its report. Then run `prep brief <id> --check --out` for a separate checker, pointing the same prompt at `check-brief.md`. Check knowledge updates and other tracker-only Definition of Done items yourself before `prep handback <id> --check`. Implementers and checkers follow their brief, not the workflow rules, and make no prep calls; the orchestrator keeps the records. To read the state of several issues at once, `prep list --under <parent>` prints one line each (state, kind, title); `prep show` prints a whole issue.

```sh
prep init                                    # create .prep/
prep new --title "CSV export" --kind code --body "Users can export their data as CSV."
prep guide <id>                              # what to do next, what blocks it, where outputs go; carries the requirement, criteria and context
prep define <id>                             # requirement settled: write a baseline
# enrich: context.md, knowledge decision records, acceptance.md
prep ready <id>
prep next                                    # actionable issues
prep claim <id>
prep complete <id>
git commit                                   # code and the staged .prep records together
```

The project Definition of Done is guidance shown by `prep prime` and `prep guide` at the check step. Edit it with `prep dod --add "<item>"` or `prep dod --remove "<item>"`. Completion checks the issue’s state, requirement freshness, checked criteria, kind-specific result: decision issues need a linked decision record with `outcome: true`, and research issues need a linked finding record. Completion has no documentation flags or gate.

`prep brief <id>` prints a hand-off text for an agent that has no prep: the requirement, the implementation context, the numbered acceptance criteria, the decisions that apply, the Definition of Done, the rules of the role and the format of the report to return. `--check` gives the checker's version. The text names no prep command; it ends with the absolute path of the file the agent writes its report to, `.prep/local/handoff/<id>/report.md` (`check-report.md` for a checker), and asks for a one-line answer. `--out` writes the brief to `brief.md` (`check-brief.md`) in the same directory and prints only that path, so an orchestrator hands over a path, not the text; `--out <file>` writes elsewhere. The directory is local to the machine and not versioned. Decisions linked to the issue are given in full; those linked to an ancestor, or whose scope matches a path the context names, by title and description; superseded ones are left out, and the section is bounded and says what it cut or left out. Definition of Done items that name a prep command stay with the orchestrator. An implementer's brief after a failed check carries the newest "Check failed" line of the work log. A brief is given for a code or research issue that is ready or in progress, not stale, not a parent, with its dependencies done. Code implementers report their work, decisions and findings without criterion verdicts. Research implementers must report findings and a verdict per criterion in a Criteria section. Every checker also reports Criteria. Keep that required section even if there are no criteria.

`prep handback <id>` imports that report in one command. It reads `.prep/local/handoff/<id>/report.md` by default, or `check-report.md` with `--check`; `--report <file>` reads elsewhere and `--report -` reads stdin. A code implementer's summary becomes an "Implemented, ready for check" log line, decisions and findings become linked knowledge records, problems are logged, and a ready issue is claimed. Code implementers report no criterion verdicts and never complete the issue. A decision's optional `Scope: <glob>[, <glob>]` line directly under its `###` title sets its scope. Checker and research reports require a Criteria section with numbered `met`/`not met` verdicts for every criterion, even if the section is empty. Research reports without `--check` write their findings and update criteria directly. These reports and checker reports check met criteria and uncheck unmet ones: all met with Problems empty completes the issue under the usual gates (`--commit <hash>` supplies evidence); otherwise a "Check failed" log records the unmet criteria and problems and the issue stays in progress. The whole report is validated before any records are written, under one lock with the resolution last. Heading case, blank lines, verdict punctuation and extra sections are tolerated; missing required content, unknown criterion numbers and decisions without reasons are refused. Successful file reports are renamed to `*.done.md` so they are not imported twice; stdin has no file to move. Output is one short line per record written, or a single JSON result with `--json`.

Every read command (`prime`, `guide`, `brief`, `list`, `next`, `show`, `check`, `views`) accepts `--json`; errors are JSON with stable codes under `--json`. Write commands (`new`, `define` (with `--trivial` for a trivial change), `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`) perform one transition each; `handback` batches the records of one report. Harness permission rules can allow reads and ask before writes. IDs accept any unique suffix. Name the actor with `--by` or `PREP_ACTOR` (`<producer>/<version>` for agents, `human:<id>` for people).

## Lifecycle

Lifecycle JSON responses preserve their existing fields and add success markers and actionable next commands where applicable.

State is derived from which records exist in `.prep/issues/<id>/`, never stored:

| State | Record | Written by |
| --- | --- | --- |
| open | `issue.md` | `prep new`, `prep edit` |
| defined | `baselines/<timestamp>.md` | `prep define`, `prep define --trivial` |
| ready | `ready.md` referencing the newest baseline | `prep ready` |
| in progress | `claim.md` | `prep claim` (`prep release` removes it and logs to `history.md`) |
| done / dropped | `resolution.md` | `prep complete`, `prep drop` (`prep reopen` removes a done resolution and logs it to `history.md`) |

An issue is **stale** when its requirement or kind differs from the newest baseline: `prep define --trivial` for a trivial change, `prep define` and re-enrichment for a real one. **Actionable** = ready, not stale, dependencies done, unclaimed, not a parent.

Dependencies refer to issues in the same project. `prep projects` lists projects registered on this machine; `--project <name>` runs a command in another registered project.

Prep serializes cooperating record writers within one checkout: each write command holds a lock from load to its last write, so a second writer waits and plans on the first one's result. Files are written one at a time through a temporary file and rename, the file that defines an issue's state last. Claims and activity remain local to each branch/worktree, so assigning parallel agents is the orchestrator's responsibility. This storage guarantee covers process interruption, with per-file atomic renames and no rollback; it does not coordinate other worktrees or promise power-loss atomicity.

## Checks

Run `prep check` to validate the tree and `prep fmt --check` to verify canonical formatting. Follow `prep guide <id>` for the current step and its gates; check the acceptance criteria and the project Definition of Done in a separate check pass before completion.

See the [knowledge base](knowledge-base.md) for decisions and findings and the [command reference](commands.md) for command usage.
