---
type: component
title: Claude Code integration
description: 'The prep plugin for Claude Code: skill, SessionStart briefing, /prep:status, /prep:next, /prep:guide and the side panel (/prep:focus, /prep:pane), served from this repository as its own marketplace and installed per user by prep setup.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - plugins/claude-code
  - .claude-plugin
  - internal/setup/claudecode
  - .claude/skills
confirmed_commit: 51db3fd5336871f85130e1345cc3e52d12de26db
---

# Claude Code integration

Applies when changing what Claude Code sessions get from prep or how the plugin is installed. Decided in 20261006-084956 and its children.

- **Plugin** (`plugins/claude-code`): `.claude-plugin/plugin.json` (name `prep`, version), `skills/prep/SKILL.md` (identical to `internal/cli/skill.md`; `just sync-skill` copies it and `TestClaudeCodePlugin` checks it), `commands/status.md`, `next.md`, `guide.md` (invoked as `/prep:status`, `/prep:next`, `/prep:guide`; read-only CLI calls through inline bash with `allowed-tools: Bash(prep:*)`), and `hooks/hooks.json` with a SessionStart hook: `command -v prep >/dev/null 2>&1 && prep prime --hook --plugin "${CLAUDE_PLUGIN_ROOT}" || true`. `hooks.json` also names the side panel's hooks module under `modules`.
- **Side panel** (20261005-152621; a Claude Code function-hooks module): `hooks/register.tsx` wires it, `hooks/infer.ts` infers the current issue, `hooks/panel.tsx` draws it, `types/index.d.ts` declares its `$.state` (`prep.focus`, `prep.pin`, `prep.snapshot`; `plugin.json` names it under `types`), `tests/panel.test.tsx` runs under `claude plugin test`.
  - **Following**: `tool.call` hooks run after the call and act only when it succeeded: a Bash command's last `prep <write|guide> <ref>` (env assignments skipped, heredoc bodies ignored, value flags skipped) or the ID in `prep new`'s output, and Edit/Write paths under `.prep/issues/<id>/`, set the focus; reads (show, list, next) do not. Focus lives in the session's `$.state`, so each session follows its own agent. The shown issue is the pin, else the focus, else the first claim in `prep prime`, else the project overview (bootstrap alert, in progress, actionable, attention, top-level parents).
  - **Data**: `prep show`, `prep guide`, `prep list` (titles and states of relations) and `prep prime`, all `--json` through `$.process.run`; overlapping refreshes keep only the newest. `prep watch` runs through `$.process.spawn` for the session and each line refreshes. At `session.start` a failing `prep prime` (no binary, no project) leaves the module idle.
  - **Drawing**: a `ui.render` hook on the `Pane` `prep` with Box, Text and Markdown: header (title, id, kind, tags, state, step, source, stale/blocked), next transition with unmet gates (ack when stale, else the first besides ack, release, drop), requirement (14 lines), open questions, acceptance with progress, DoD, decisions, surroundings (parent, dependencies, blocks, children, knowledge from guide), last 3 history lines. Colors are Claude Code theme keys mapped from the TUI tokens (states: open inactive, defined suggestion, ready ide, in progress warning, done success, dropped subtle).
  - **Priority**: the live header shows a non-medium priority after the state, and project, overview and relation rows mark it before the title (`!crit` error, `!high` warning, `low` subtle), from `priority` in prep's JSON.
  - **Views** (20261006-105823): `Live` and `Project` tabs (plain Buttons `tab-live`, `tab-project`, hotkeys `l` and `p`) head the pane; `prep.tab` holds the choice for the session. Live is everything above. Project shows counts, check, bootstrap alert, in progress and stale issues, then `prep list --tree --state open,defined,ready,in_progress` as an indented tree with state, parent progress and stale/blocked/actionable marks; display only. Its data (`prep.project`) loads only while it shows and on each refresh then.
  - **Commands**: `commands/focus.md` and `pane.md` declare `/prep:focus [id]` and `/prep:pane [live|project]` (registered commands cannot carry the `prep:` namespace); `command.run` hooks answer them without reaching the model, and the markdown body only tells the user the hooks did not load. `/prep:focus <ref>` resolves through `prep show`, pins and opens the pane; without an id it unpins. `/prep:pane live|project` opens the pane on that view; without an argument it toggles.
  - **Opening**: `userConfig.pane` (remember, default; open; closed). `$.store` keeps `open:<cwd>` per project: `/prep:pane` and a close by the person write it; remember opens unless it is false. Claude Code seats a pane opened unasked from 144 columns only.
  - **Developing**: the installed plugin is read from this repository's folder marketplace, so `/reload-plugins` (or a new session) loads edits; `claude plugin validate plugins/claude-code` and `claude plugin test plugins/claude-code`. Hooks on tool calls and `ui.close` carry `.catch(passOn)` so a panel failure never changes the call.
- **Hook behavior**: silent without the binary and, through `prime --hook`, outside a prep project. `--plugin` makes prime compare the plugin's `plugin.json` version with the binary and lead with a warning naming `prep setup --refresh` (older plugin) or `prep update` (older binary); development versions are not compared ([Agent context](/features/agent-context.md)).
- **Marketplace**: `.claude-plugin/marketplace.json` at the repository root makes this repository the marketplace `prep`, listing the plugin with source `./plugins/claude-code`. Both manifests carry the same version; `just release` sets it and the release workflow refuses a tag that differs ([Release, install and update](/components/release.md)).
- **Installation** (`internal/setup/claudecode`, a [setup](/components/setup.md) harness): detects `claude` on PATH, reads `claude plugin list --json` (user scope `prep@prep`), installs by adding the marketplace pinned to the binary's tag (`fluid-movement/prep#vX.Y.Z`; `PREP_PLUGIN_SOURCE` overrides, required for development builds) and `claude plugin install prep@prep --scope user --json`; updating removes and re-adds the marketplace to move the pin; removal uninstalls the plugin and removes the marketplace. Pinning exists because third-party marketplaces do not auto-update.
- **Cloud sessions** load no plugins, so this repository keeps `.claude/skills/prep/SKILL.md` as the static skill copy (20261005-185825 decides further).
- The panel was tried live in a session: opening at start, /prep:pane, following prep guide and writes, ignoring reads, and live updates through prep watch.
- Verified with the real Claude Code CLI in an isolated `CLAUDE_CONFIG_DIR` (install, refresh, remove) and in a session started with `--plugin-dir`, which received the prime briefing from the hook and listed the commands.
