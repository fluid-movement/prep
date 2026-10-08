---
type: component
title: Claude Code side panel
description: 'The prep pane in Claude Code: activation, following the agent, data, drawing, the Live, Project and Usage views, /prep:focus and /prep:pane, opening and developing.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T06:36:05Z
scope:
  - plugins/claude-code/hooks
  - plugins/claude-code/types
  - plugins/claude-code/tests
  - plugins/claude-code/commands/focus.md
  - plugins/claude-code/commands/pane.md
confirmed_commit: 1befaf298ace4ab6b58f598ad21dc03209bcc6e7
---

# Claude Code side panel

Applies when changing the prep pane in Claude Code sessions. Decided in 01M46ARKA817KMB2C92WMNH2QP; the plugin around it is [Claude Code integration](/components/claude-code.md).

- **Module**: a Claude Code function-hooks module named under `modules` in `plugins/claude-code/hooks/hooks.json`. `hooks/register.tsx` wires it, `hooks/infer.ts` infers the current issue, `hooks/panel.tsx` draws it, `types/index.d.ts` declares its `$.state` (`prep.focus`, `prep.pin`, `prep.snapshot`, `prep.tab`, `prep.project`, `prep.usage`; `plugin.json` names it under `types`), `tests/panel.test.tsx` runs under `claude plugin test`.
- **Following**: `tool.call` hooks run after the call and act only when it succeeded: a Bash command's last `prep <write|guide> <ref>` (a ULID or suffix, made the full ID when the output shows it; shell variables assigned earlier in the command expanded, a reference left as a variable falling back to the last issue ID in the output, heredoc bodies ignored, value flags skipped) or the ID in `prep new`'s output (not for `prep new --help`), and Edit/Write paths under `.prep/issues/<id>/`, set the focus; reads (show, list, next) do not. Focus lives in the session's `$.state`, so each session follows its own agent. The shown issue is the pin, else the focus, else the first claim in `prep prime`, else the project overview (bootstrap alert, in progress, actionable, attention, top-level parents).
- **Data**: `prep show`, `prep guide`, `prep list` (titles and states of relations) and `prep prime`, all `--json` through `$.process.run`; overlapping refreshes keep only the newest. `prep watch` runs through `$.process.spawn` for the session and each line refreshes. `activate` (`prep prime`, then load and watch) runs at `session.start` and, while idle, after the agent's `prep init` and on `/prep:pane` and `/prep:focus`, which answer `Not a prep project: <prime's error>`.
- **Drawing**: a `ui.render` hook on the `Pane` `prep` with Box, Text and Markdown: header (title, id, kind, tags, state, step, source, stale/blocked), next transition with unmet gates (ack when stale, else the first besides ack, release, drop), requirement (14 lines), open questions, acceptance with progress, DoD, decisions, surroundings (parent, dependencies, blocks, children, knowledge from guide), last 3 history lines. Colors are Claude Code theme keys mapped from the TUI tokens (states: open inactive, defined suggestion, ready ide, in progress warning, done success, dropped subtle).
- **Priority**: the live header shows a non-medium priority after the state, and project, overview and relation rows mark it before the title (`!crit` error, `!high` warning, `low` subtle), from `priority` in prep's JSON.
- **Views** (01M48DTN4RTN3YN7QM6WJV94V7): `Live`, `Project` and `Usage` tabs (Buttons `tab-live`, `tab-project`, `tab-usage` drawn as buttons, `[ Live ]` `[ Project ]` `[ Usage ]`, the active one `variant: 'primary'`, no hotkeys; 01M4AW8EXXDSBR64ERQVWY0ZMN) head the pane; `prep.tab` holds the choice for the session. Live is everything above. Project shows counts, check, bootstrap alert, in progress and stale issues, then `prep list --tree --state open,defined,ready,in-progress` as an indented tree with state, parent progress and stale/blocked/actionable marks; display only. Its data (`prep.project`) loads only while it shows and on each refresh then.
- **Usage view** (`hooks/usage.ts`): the Claude Code session's context fill, cost and id from `$.session.usage()` and `$.session.id()`, then, when the token-ledger plugin (`plugins/token-ledger`, also in the marketplace) has written `~/.claude/token-ledger/<session id>.jsonl`, request tokens (input, cache read and write, output), tool output read back per area (knowledge base, including `prep knowledge list|find|show`; issue files; other prep commands; code; chars/4) and the largest knowledge reads. Without it, the view names the install command. `prep.usage` loads when the tab opens and, while it shows, 500 ms after each `turn.complete`, after the ledger's write. Measurement for agent-context decisions; prep itself never reads the ledger. The ledger keeps a copy of `infer.ts` to stamp its rows with the issue being worked on (reset when the session ID changes); change both together.
- **Commands**: `commands/focus.md` and `pane.md` declare `/prep:focus [id]` and `/prep:pane [live|project|usage]` (registered commands cannot carry the `prep:` namespace); `command.run` hooks answer them without reaching the model, and the markdown body only tells the user the hooks did not load. `/prep:focus <ref>` resolves through `prep show`, pins and opens the pane; without an id it unpins. `/prep:pane live|project|usage` opens the pane on that view; without an argument it toggles.
- **Opening**: `userConfig.pane` (remember, default; open; closed). `$.store` keeps `open:<cwd>` per project: `/prep:pane` and a close by the person write it; remember opens unless it is false. Claude Code seats a pane opened unasked from 144 columns only.
- **Developing**: the installed plugin is read from this repository's folder marketplace, so `/reload-plugins` (or a new session) loads edits; `claude plugin validate plugins/claude-code` and `claude plugin test plugins/claude-code`. Hooks on tool calls and `ui.close` carry `.catch(passOn)`: a panel failure never changes the call.
- The panel was tried live in a session: opening at start, /prep:pane, following prep guide and writes, ignoring reads, and live updates through prep watch.
