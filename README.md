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
# or, from a checkout (needs just): builds with the version stamped in and
# copies the binary to ~/.local/bin (override with PREP_BINDIR or bindir=...)
just install
```

## Harness integrations

The binary is half of prep; the other half teaches your agent to use it. After installing, the script runs `prep setup`, which detects your agent harnesses and installs their integration per user. For Claude Code that is the prep plugin: the prep skill, a briefing at the start of every session in a prep project, and `/prep:status`, `/prep:next` and `/prep:guide`. Run `prep setup` again at any time to add or remove harnesses; your choices live in `~/.config/prep/config.yaml`.

## Update

```sh
prep update          # replace prep with the latest release, checksum verified
prep update --check  # only report whether a newer release exists
```

When prep was installed with Homebrew or `go install`, `prep update` names the command that updates it instead. Releases are cut by pushing a version tag (`git tag v0.1.0 && git push origin v0.1.0`); the release workflow builds and publishes them with GoReleaser.

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

Every read command (`prime`, `guide`, `list`, `next`, `show`, `check`, `views`) accepts `--json`; errors are JSON with stable codes under `--json`. Write commands (`new`, `import`, `define`, `ack`, `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`, `migrate`) perform one transition each, so harness permission rules can allow reads and ask before writes. IDs accept any unique suffix. Name the actor with `--by` or `PREP_ACTOR` (`<producer>/<version>` for agents, `human:<id>` for people).

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

## Views

The tabs of `prep tui` are the saved views in the project configuration, in the order the file lists them. A view is a name and a query, with the same flags as `prep list`. `prep init` writes two to `.prep/config.yaml.dist`: all unresolved work first, where each row's state glyph tells the states apart, then every issue.

The configuration is personal. The project commits `.prep/config.yaml.dist` as the shared default; your own `.prep/config.yaml` is gitignored and, when it exists, replaces the dist file as a whole (nothing is merged). Copy the dist file to start your own, or change a setting in the TUI, which creates it for you.

```yaml
views:
  Unresolved: --state open,defined,ready,in_progress
  All: ""
```

A query takes any number of flags, and an issue must match all of them; a comma-separated list within one flag matches any of its values (`--state ready --tag tui,cli --priority high,critical`). The flags are `--state`, `--kind`, `--tag`, `--priority`, `--under <id>` (descendants of an issue), `--text` (title or requirement), `--stale`, `--blocked`, `--actionable`, `--parent`, `--leaf` and `--top`; the boolean ones accept `=false`. An empty query lists every issue. `prep flags` prints every flag with the values it accepts in your project right now (states, kinds, priorities, the tags in use, the parents for `--under`, the saved views) and how many issues each matches. In `prep tui`, the filter bar (`f`) offers the same values while you type: after `--` the flags, after a flag its values; ↑/↓ and tab or a click insert one.

Change the views freely. In the TUI, `s` opens the settings screen, where views are added, edited, deleted and reordered. `prep views` lists them, and `prep list --view <name>` runs one on the command line.

## Themes

The themes of `prep tui` are made for dark terminals: `default`, `high-contrast`, `monochrome`, `pastel`, `catppuccin`, `nord` and `gruvbox`. For light terminals there is one theme, `light`, to use as is or to build your own on. Without a configured theme, prep shows `default` on a dark terminal and `light` on a light one. `prep tui --gallery` shows every component; `t` and `T` switch to the next and previous theme, so you can compare them. Pick one on the settings screen (`s`, then `t`/`T`) or in your `.prep/config.yaml`:

```yaml
theme: nord
```

A custom theme starts from a base and overrides any of its colors:

```yaml
theme: mine
themes:
  mine:
    base: nord
    accent: "#FF9E5E"
    selection: "#3A4254"
```

A theme has eight colors: the gray ramp `text`, `muted` and `subtle`, the `accent`, the `selection` background, and `success`, `warning` and `error`. Borders, states and kinds draw from them: states have their own glyphs, so only ready (`success`) and in progress (`accent`) stand out, and finished work fades to `subtle`. A theme based on `light` is a light theme. `prep theme new <name> [--from <theme>]` writes a custom theme with all eight colors filled in from the active theme (or `--from`) and selects it. `prep theme list` shows the themes and marks the active one; `prep check` reports unknown tokens, invalid colors and unknown bases. Terminals with 256 or 16 colors get the nearest colors, and `NO_COLOR` turns colors off.

## Development

```sh
just test      # go test ./..., including the contract corpus in testdata/contract
just check     # prep check and prep fmt --check on this repository's .prep
just ci        # lint, test and check: everything CI runs
```

Run `just` to list all recipes.

The agent skill lives in `.claude/skills/prep/SKILL.md` (also served by `prep skill`); a test keeps it identical to the copy embedded in the binary.
