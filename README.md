# prep

prep is a workflow engine and memory for coding agents, with a human supervising. It enforces the steps an issue goes through (define → enrich → implement → resolve) and keeps the context agents need in the repository, versioned with the code. How prep works and why is documented in its own knowledge base, `.prep/knowledge/` (start at `overview.md`); open work is in `.prep/issues/` (`prep list`, or `prep tui`).

## Install

```sh
go install github.com/fluid-movement/prep/cmd/prep@latest
# or, from a checkout (needs just): builds with the version stamped in and
# copies the binary to ~/.local/bin (override with PREP_BINDIR or bindir=...)
just install
```

## Quick start

```sh
prep init                                    # create .prep/
prep new --title "CSV export" --kind code --body "Users can export their data as CSV."
prep guide <id>                              # what to do next, what blocks it, where outputs go
prep define <id>                             # requirement settled: write a baseline
# enrich: context.md, decisions.md, acceptance.md
prep ready <id>
prep next                                    # actionable issues
prep claim <id>
prep complete <id> --commit <hash> --docs /components/export.md   # or --no-impact <reason>
```

Every read command (`prime`, `guide`, `list`, `next`, `show`, `check`, `views`) accepts `--json`; errors are JSON with stable codes under `--json`. Write commands (`new`, `define`, `ack`, `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`, `migrate`) perform one transition each, so harness permission rules can allow reads and ask before writes. IDs accept any unique suffix. Name the actor with `--by` or `PREP_ACTOR` (`<producer>/<version>` for agents, `human:<id>` for people).

## Lifecycle

State is derived from which records exist in `.prep/issues/<id>/`, never stored:

| State | Record | Written by |
| --- | --- | --- |
| open | `issue.md` | `prep new`, then edited by hand |
| defined | `baselines/<timestamp>.md` | `prep define`, `prep ack` |
| ready | `ready.md` referencing the newest baseline | `prep ready` |
| in progress | `claim.md` | `prep claim` (`prep release` removes it and logs to `history.md`) |
| done / dropped | `resolution.md` | `prep complete`, `prep drop` |

An issue is **stale** when its requirement or kind differs from the newest baseline: `prep ack` for a trivial change, `prep define` and re-enrichment for a real one. **Actionable** = ready, not stale, dependencies done, unclaimed, not a parent.

## Development

```sh
just test      # go test ./..., including the contract corpus in testdata/contract
just check     # prep check and prep fmt --check on this repository's .prep
just ci        # lint, test and check: everything CI runs
```

Run `just` to list all recipes.

The agent skill lives in `.claude/skills/prep/SKILL.md` (also served by `prep skill`); a test keeps it identical to the copy embedded in the binary.
