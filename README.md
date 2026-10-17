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

## Harness integrations

The binary is half of prep; the other half teaches your agent to use it. After installing, the script runs `prep setup`, which detects your agent harnesses and installs their integration per user. You can choose Codex explicitly with `prep setup --harness codex`. For Claude Code and Codex CLI, prep installs a plugin with the shared prep skill and lifecycle hooks. Session start identifies the prep project and supplies its general briefing; each prompt steers the agent to search existing issues and use prep tools; turn finish checks whether issue progress and knowledge need updating after the turn's changes. Supporting tool hooks redirect direct record access and report activity. Claude also reports request token usage and cost. Codex asks you to review and trust plugin hooks before they run; open `/hooks`, review the updated definitions, then start a new session. Hooks cover supported local tools and fail open on integration errors. These integrations act only in repositories initialized with `.prep/`; projects do not need an `AGENTS.md` file. Run `prep tui --agent` next to your agent, beside it or in a strip above or below, to watch what it does: its current issue in full, a live feed of its actions, and its usage. Run `prep setup` again at any time to add or remove harnesses; your choices live in `~/.config/prep/config.yaml`.

Prompt context is a global choice in `~/.config/prep/config.yaml`:

```yaml
hooks:
  prompt: small
```

`small` is the default concise workflow reminder; `large` also includes bounded guides for up to two full issue IDs; `off` skips prompt input and project lookup. Session context, record redirects, activity and finish remain enabled. Turn finish uses Git snapshots of dirty files and commits, ignoring unchanged pre-existing work; it asks for one issue/knowledge review pass when needed and respects the stop-loop flag. Git/snapshot failures leave the turn alone. With prompt context off, its boundary is session start or the previous finish. Codex session context tells the agent to set `PREP_ACTOR` and `PREP_SESSION` only on prep CLI invocations. Other shell commands keep their usual form so approved command prefixes can still match; hooks attribute their activity directly from event payloads and do not auto-approve or rewrite them. Keep plugin and binary versions in sync (`prep setup --refresh`).

## Update

```sh
prep update          # replace prep with the latest release, checksum verified
prep update --check  # only report whether a newer release exists
```

When prep was installed with Homebrew or `go install`, `prep update` names the command that updates it instead. Releases are cut with `just release vX.Y.Z` (it sets the plugin version, runs the tests, commits and tags) and published by pushing the tag; the release workflow builds and publishes them with GoReleaser.

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
prep complete <id> --docs /components/export.md   # or --no-impact <reason>
git commit                                   # code and the staged .prep records together
```

prep stages the files it writes and never commits; you (or your agent) commit them with the code they describe. A code issue completed without `--commit` takes as its evidence the commit that adds its `resolution.md`, so code, records and knowledge updates land in one commit; `--commit <hash>` names work committed earlier. A knowledge entry counts as current as of the last commit that changed it (or its `confirmed_commit`, if later), so updating or confirming it in the same commit as the code keeps it from drifting.

The project Definition of Done is guidance shown by `prep prime` and `prep guide` at the check step. Edit it with `prep dod --add "<item>"` or `prep dod --remove "<item>"`. Completion checks the issue’s state, requirement freshness, checked criteria, kind-specific result, and documentation decision (`--docs` or `--no-impact`).

Every read command (`prime`, `guide`, `list`, `next`, `show`, `check`, `views`) accepts `--json`; errors are JSON with stable codes under `--json`. Write commands (`new`, `define` (with `--trivial` for a trivial change), `ready`, `claim`, `release`, `complete`, `drop`, `fmt`, `fix`) perform one transition each, so harness permission rules can allow reads and ask before writes. IDs accept any unique suffix. Name the actor with `--by` or `PREP_ACTOR` (`<producer>/<version>` for agents, `human:<id>` for people).

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

## The TUI

`prep tui` has two rows of navigation. The top row is the screens: Issues (`i`), Agent (`w`) and Knowledge (`b`), with the check's counts and a gear for the settings at its right end. Below it are the views of the screen you are on: your saved issue lists on Issues; Issue, Activity and Usage on Agent. `tab` and `1`–`9` move between views, and backspace (or the mouse's back button) returns to where you were, across screens. Everything can be clicked. On the issue list, `a` opens the actions that apply to the selected issue, `e` edits it, `p` sets its priority and `n` creates one; `?` lists every key. In a terminal under 16 rows both navigation rows share one line.

## Views

The tabs of `prep tui` are the saved views in the project configuration, in the order the file lists them. A view is a name and a query, with the same flags as `prep list`. `prep init` writes two to `.prep/config.yaml.dist`: all unresolved work first, where each row's state glyph tells the states apart, then every issue.

The configuration is personal. The project commits `.prep/config.yaml.dist` as the shared default; your own `.prep/config.yaml` is gitignored and, when it exists, replaces the dist file as a whole (nothing is merged). Copy the dist file to start your own, or change a setting in the TUI, which creates it for you.

```yaml
views:
  Unresolved: --state open,defined,ready,in-progress
  All: ""
```

A query takes any number of flags, and an issue must match all of them; a comma-separated list within one flag matches any of its values (`--state ready --tag tui,cli --priority high,critical`). The flags are `--state`, `--kind`, `--tag`, `--priority`, `--under <id>` (descendants of an issue), `--text` (title or requirement), `--stale`, `--blocked`, `--actionable`, `--parent`, `--leaf` and `--top`; the boolean ones accept `=false`. A leading `not-` negates a value: `--state not-open` lists every issue that is not open, `--tag not-cli` the ones without the tag; within one flag, plain values include and negated ones exclude (`--state ready,in-progress,not-dropped`). `!` works the same and is quicker to type in the TUI filter bar and in views (`--state !open`); in a shell it needs quotes, since shells expand an unquoted `!`. Tags cannot start with `not-`. Values use hyphens throughout, such as the state `in-progress`. An empty query lists every issue. `prep flags` prints every flag with the values it accepts in your project right now (states, kinds, priorities, the tags in use, the parents for `--under`, the saved views) and how many issues each matches. In `prep tui`, the filter bar (`f`) offers the same values while you type: after `--` the flags, after a flag its values, after a finished value and a space the flags not used yet; ↑/↓ choose, enter or tab (or a click) inserts one, and enter with nothing left to pick applies the filter. The list follows the query as you type whenever it is complete; esc goes back to the filter the tab had before.

Change the views freely. In the TUI, `s` opens the settings screen, where views are added, edited, deleted and reordered. `prep views` lists them, and `prep list --view <name>` runs one on the command line.

## Themes

`prep tui` has two themes: `default` for dark terminals and `light` for light terminals. Without a configured theme, prep chooses from the terminal background. In settings (`s`), press enter or space on the Theme row, or `t`/`T`, to switch directly between them. `prep tui --gallery` shows every component; `t` and `T` switch between the two themes. You can also select a theme in `.prep/config.yaml`:

```yaml
theme: light
```

A theme has eight colors: the gray ramp `text`, `muted` and `subtle`, the `accent`, the `selection` background, and `success`, `warning` and `error`. Borders, states and kinds draw from them: states have their own glyphs, so only ready (`success`) and in progress (`accent`) stand out, and finished work fades to `subtle`. Terminals with 256 or 16 colors get the nearest colors, and `NO_COLOR` turns colors off.

## Development

```sh
just test      # go test ./..., including the contract corpus in testdata/contract
just check     # prep check and prep fmt --check on this repository's .prep
just ci        # lint, test and check: everything CI runs
```

Run `just` to list all recipes.

The agent skill lives in `.claude/skills/prep/SKILL.md` (also served by `prep skill`); a test keeps it identical to the copy embedded in the binary.
