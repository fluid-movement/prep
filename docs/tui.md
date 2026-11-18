# The TUI

[Documentation](../README.md#documentation)

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

See [harness integrations](harness-integrations.md) for the agent activity report.
