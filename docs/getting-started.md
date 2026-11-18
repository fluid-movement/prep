# Getting started

[Documentation](../README.md#documentation)

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh
```

The script downloads the release for your machine (macOS or Linux, amd64 or arm64), verifies it against the release checksums and installs `prep` into `~/.local/bin`. `PREP_VERSION=v0.1.0` picks a version, `PREP_BINDIR` another directory. Windows builds are on the [releases page](https://github.com/fluid-movement/prep/releases).

Other ways:

```sh
go install github.com/fluid-movement/prep/cmd/prep@latest
# or, from a checkout (needs just): builds with the version stamped in,
# copies the binary to ~/.local/bin (override with PREP_BINDIR or bindir=...)
# and points the harness integrations (the Claude Code plugin) at the checkout
just install
```

## Update

```sh
prep update          # replace prep with the latest release, checksum verified
prep update --check  # only report whether a newer release exists
prep cloud script    # print the cloud environment setup script for this version
```

When prep was installed with Homebrew or `go install`, `prep update` names the command that updates it instead. Releases are cut with `just release vX.Y.Z` (it sets the plugin version, runs the tests, commits and tags) and published by pushing the tag; the release workflow builds and publishes them with GoReleaser.

## Quick start

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

Follow the [workflow](workflow.md) to define, ready, claim and complete issues. Record decisions and findings in the [knowledge base](knowledge-base.md).

## Development

```sh
just test      # go test ./...
just check     # prep check and prep fmt --check on this repository's .prep
just ci        # lint, test and check: everything CI runs
```

Run `just` to list all recipes.

The agent skill lives in `.claude/skills/prep/SKILL.md` (also served by `prep skill`); a test keeps it identical to the copy embedded in the binary.

Run `just docs-commands` to regenerate the [command reference](commands.md). `just ci` checks that it is current and that relative documentation links resolve.
