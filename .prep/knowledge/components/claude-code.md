---
type: component
title: Claude Code integration
description: 'The prep plugin for Claude Code: skill, SessionStart briefing and /prep:status, /prep:next, /prep:guide, served from this repository as its own marketplace and installed per user by prep setup.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - plugins/claude-code
  - .claude-plugin
  - internal/setup/claudecode
  - .claude/skills
confirmed_commit: 4ab8b4e482cdb4a9f26dc676acbaeacf47ba7c05
---

# Claude Code integration

Applies when changing what Claude Code sessions get from prep or how the plugin is installed. Decided in 20261006-084956 and its children.

- **Plugin** (`plugins/claude-code`): `.claude-plugin/plugin.json` (name `prep`, version), `skills/prep/SKILL.md` (identical to `internal/cli/skill.md`; `just sync-skill` copies it and `TestClaudeCodePlugin` checks it), `commands/status.md`, `next.md`, `guide.md` (invoked as `/prep:status`, `/prep:next`, `/prep:guide`; read-only CLI calls through inline bash with `allowed-tools: Bash(prep:*)`), and `hooks/hooks.json` with a SessionStart hook: `command -v prep >/dev/null 2>&1 && prep prime --hook --plugin "${CLAUDE_PLUGIN_ROOT}" || true`. A `mod/` directory can take the TUI pane later (20261005-152621).
- **Hook behavior**: silent without the binary and, through `prime --hook`, outside a prep project. `--plugin` makes prime compare the plugin's `plugin.json` version with the binary and lead with a warning naming `prep setup --refresh` (older plugin) or `prep update` (older binary); development versions are not compared ([Agent context](/features/agent-context.md)).
- **Marketplace**: `.claude-plugin/marketplace.json` at the repository root makes this repository the marketplace `prep`, listing the plugin with source `./plugins/claude-code`. Both manifests carry the same version; `just release` sets it and the release workflow refuses a tag that differs ([Release, install and update](/components/release.md)).
- **Installation** (`internal/setup/claudecode`, a [setup](/components/setup.md) harness): detects `claude` on PATH, reads `claude plugin list --json` (user scope `prep@prep`), installs by adding the marketplace pinned to the binary's tag (`fluid-movement/prep#vX.Y.Z`; `PREP_PLUGIN_SOURCE` overrides, required for development builds) and `claude plugin install prep@prep --scope user --json`; updating removes and re-adds the marketplace to move the pin; removal uninstalls the plugin and removes the marketplace. Pinning exists because third-party marketplaces do not auto-update.
- **Cloud sessions** load no plugins, so this repository keeps `.claude/skills/prep/SKILL.md` as the static skill copy (20261005-185825 decides further).
- Verified with the real Claude Code CLI in an isolated `CLAUDE_CONFIG_DIR` (install, refresh, remove) and in a session started with `--plugin-dir`, which received the prime briefing from the hook and listed the commands.
