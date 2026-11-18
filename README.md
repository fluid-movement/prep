# prep

prep is a workflow engine and memory for coding agents, with a human supervising. It enforces the steps an issue goes through (define → enrich → implement → resolve) and keeps the context agents need in the repository, versioned with the code. How prep works and why is documented in its own knowledge base, `.prep/knowledge/` (start at `overview.md`); open work is in `.prep/issues/` (`prep list`, or `prep tui`).

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

## Quick start

Working alone, follow `prep guide <id>` through implementation and a separate check pass. To delegate, run `prep brief <id> --out` and give the implementer only this one-line prompt: "Read .prep/local/handoff/<id>/brief.md and follow it exactly. It tells you where to write your report." Run `prep handback <id>` to import its report. To delegate to Claude Code cloud sessions instead, run `prep handoff start <id>...`, then `prep handoff fetch --merge`; see [cloud sessions](docs/cloud-sessions.md). Then run `prep brief <id> --check --out` for a separate checker, pointing the same prompt at `check-brief.md`. Check knowledge updates and other tracker-only Definition of Done items yourself before `prep handback <id> --check`. Implementers and checkers follow their brief, not the workflow rules, and make no prep calls; the orchestrator keeps the records. To read the state of several issues at once, `prep list --under <parent>` prints one line each (state, kind, title); `prep show` prints a whole issue.

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

## Documentation

Browse the docs at **[fluid-movement.github.io/prep](https://fluid-movement.github.io/prep/)**, or read the pages in [`docs/`](docs/index.md) directly:

| Page | Contents |
| --- | --- |
| [Getting started](docs/getting-started.md) | Installation, updates and development |
| [Workflow](docs/workflow.md) | Define, ready, claim, complete and checks |
| [Commands](docs/commands.md) | Generated command reference |
| [Harness integrations](docs/harness-integrations.md) | Claude Code, Codex, setup and activity |
| [Cloud sessions](docs/cloud-sessions.md) | Set up and run Claude Code cloud sessions |
| [Knowledge base](docs/knowledge-base.md) | Decisions and findings |
| [TUI](docs/tui.md) | Navigation, views and themes |
