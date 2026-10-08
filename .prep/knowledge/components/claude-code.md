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
confirmed_commit: 834cee63cedb618b3b8d548f5333ea1b26a9860c
---

# Claude Code integration

Applies when changing what Claude Code sessions get from prep or how the plugin is installed. Decided in 01M486FES0NQA47N0EH8T3NWCD and its children.

- **Plugin** (`plugins/claude-code`): `.claude-plugin/plugin.json` (name `prep`, version), `skills/prep/SKILL.md` (identical to `internal/cli/skill.md`; `just sync-skill` copies it and `TestClaudeCodePlugin` checks it), `commands/status.md`, `next.md`, `guide.md` (invoked as `/prep:status`, `/prep:next`, `/prep:guide`; read-only CLI calls through inline bash with `allowed-tools: Bash(prep:*)`), and `hooks/hooks.json` with a SessionStart hook: `command -v prep >/dev/null 2>&1 && prep prime --hook --plugin "${CLAUDE_PLUGIN_ROOT}" || true`. `hooks.json` also names the side panel's hooks module under `modules`.
- **Side panel** (01M46ARKA817KMB2C92WMNH2QP): the `prep` pane, `/prep:focus` and `/prep:pane`, in `hooks/`, `types/`, `tests/` and `commands/focus.md`, `pane.md` ([Claude Code side panel](/components/claude-code-panel.md)).
- **Hook behavior**: silent without the binary and, through `prime --hook`, outside a prep project. `--plugin` makes prime compare the plugin's `plugin.json` version with the binary and lead with a warning naming `prep setup --refresh` (older plugin) or `prep update` (older binary); development versions are not compared ([Agent context](/features/agent-context.md)).
- **Marketplace**: `.claude-plugin/marketplace.json` at the repository root makes this repository the marketplace `prep`, listing the plugin with source `./plugins/claude-code`. Both manifests carry the same version; `just release` sets it and the release workflow refuses a tag that differs ([Release, install and update](/components/release.md)).
- **Installation** (`internal/setup/claudecode`, a [setup](/components/setup.md) harness): detects `claude` on PATH, reads `claude plugin list --json` (user scope `prep@prep`), installs by adding the marketplace pinned to the binary's tag (`fluid-movement/prep#vX.Y.Z`; `PREP_PLUGIN_SOURCE` overrides, required for development builds) and `claude plugin install prep@prep --scope user --json`; updating removes and re-adds the marketplace to move the pin; removal uninstalls the plugin and removes the marketplace. Pinning exists because third-party marketplaces do not auto-update.
- **Cloud sessions** (Claude Code on the web) load no user plugins, so the plugin hook does not run there; this repository keeps `.claude/skills/prep/SKILL.md` as the static skill copy, which loads as a project skill. 01M46PWX38AZN55Q6T8TE93959 researched the environment and recommends a project SessionStart hook that installs the pinned release (follow-ups 01M48RV2FR0XQAGCKEXRJFH7AE, 01M48RV3F0ZZB2MXBFR2R62P9J, 01M48RV4E8J5T57FN2P09SHG2W).
- Verified with the real Claude Code CLI in an isolated `CLAUDE_CONFIG_DIR` (install, refresh, remove) and in a session started with `--plugin-dir`, which received the prime briefing from the hook and listed the commands.
