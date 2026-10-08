---
type: component
title: Claude Code integration
description: 'The prep plugin for Claude Code: skill, SessionStart briefing and the activity bridge to prep tui''s Agent screen, served from this repository as its own marketplace and installed per user by prep setup.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - plugins/claude-code
  - .claude-plugin
  - internal/setup/claudecode
  - .claude/skills
confirmed_commit: 4d7af952dcf2a8cd5a0ce0c4c865932f46625c79
---

# Claude Code integration

Applies when changing what Claude Code sessions get from prep or how the plugin is installed. Decided in 01M486FES0NQA47N0EH8T3NWCD and its children; slimmed in 01M4DTR02HTJGCNAM4XZ14MWQM, when the side panel and the commands gave way to prep tui's [Agent screen](/components/tui-agent.md).

- **Plugin** (`plugins/claude-code`): `.claude-plugin/plugin.json` (name `prep`, version), `skills/prep/SKILL.md` (identical to `internal/cli/skill.md`; `just sync-skill` copies it and `TestClaudeCodePlugin` checks it), and `hooks/hooks.json` with a SessionStart hook, `command -v prep >/dev/null 2>&1 && prep prime --hook --plugin "${CLAUDE_PLUGIN_ROOT}" || true`, and the activity bridge module under `modules`. No commands, no options.
- **Hook behavior**: silent without the binary and, through `prime --hook`, outside a prep project. `--plugin` makes prime compare the plugin's `plugin.json` version with the binary and lead with a warning naming `prep setup --refresh` (older plugin) or `prep update` (older binary); development versions are not compared ([Agent context](/features/agent-context.md)).
- **Activity bridge** (`hooks/register.ts`, `hooks/classify.ts`, from the former token-ledger plugin): at `session.start` and whenever the session ID changes (checked at `turn.start`) it sets `PREP_SESSION` and `PREP_ACTOR=claude-code` with `$.env.set`, so the agent's Bash calls carry them, its prep commands and the bridge's events meet as one agent, and its reads are recorded even without `--by` (an explicit `--by` wins) ([Agent activity](/features/agent-activity.md)). It reports to `prep activity add` (JSON lines on stdin, actor `claude-code`): each tool call after it ran (`toolEvent`: verb, target from `classify`, area knowledge/issue/code/prep/other, a shell command naming no file counts as code, output size, failure), except Bash calls that run prep, which prep records itself; each `turn.step`'s token usage; and at `turn.complete` the context fill and the turn's cost (the difference to the session's last total). Sends are queued and never awaited by a call; the first failed send (outside a prep project, no binary) turns the bridge off for the session; every hook but `turn.step` carries a `.catch` that lets the call through; a streaming hook that fails after `next` is skipped and its `next` result stands, so it needs none.
- **Tests**: `tests/bridge.test.ts` under `claude plugin test plugins/claude-code` (fake `process.run`, `env.set`, `session.id`); `claude plugin validate plugins/claude-code`.
- **Marketplace**: `.claude-plugin/marketplace.json` at the repository root makes this repository the marketplace `prep`, listing the plugin with source `./plugins/claude-code`. Both manifests carry the same version; `just release` sets it and the release workflow refuses a tag that differs ([Release, install and update](/components/release.md)).
- **Installation** (`internal/setup/claudecode`, a [setup](/components/setup.md) harness): detects `claude` on PATH, reads `claude plugin list --json` (user scope `prep@prep`), installs by adding the marketplace pinned to the binary's tag (`fluid-movement/prep#vX.Y.Z`; `PREP_PLUGIN_SOURCE` overrides, required for development builds) and `claude plugin install prep@prep --scope user --json`; updating removes and re-adds the marketplace to move the pin, which uninstalls every plugin from it, so `Install` installs the other user-scope `*@prep` plugins again and returns warnings for those that fail and for those at other scopes (01M4DMXST6T9VKMVF46H9DFR40); if adding the new pin fails, `restore` adds the marketplace back from the source `claude plugin marketplace list --json` showed (`marketplaceSource`) and reinstalls the plugins, and the error says what was kept; removal uninstalls the plugin and removes the marketplace. Pinning exists because third-party marketplaces do not auto-update.
- **Cloud sessions** (Claude Code on the web) load no user plugins, so the plugin hook does not run there; this repository keeps `.claude/skills/prep/SKILL.md` as the static skill copy, which loads as a project skill. 01M46PWX38AZN55Q6T8TE93959 researched the environment and recommends a project SessionStart hook that installs the pinned release (follow-ups 01M48RV2FR0XQAGCKEXRJFH7AE, 01M48RV3F0ZZB2MXBFR2R62P9J, 01M48RV4E8J5T57FN2P09SHG2W).
