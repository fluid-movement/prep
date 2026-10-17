---
type: component
title: Codex integration
description: 'Codex CLI plugin: setup installation and lifecycle hooks for prep briefing, task steering, record safety and activity reporting.'
generated:
  by: cli/prep-v0.2.0-20-g7487215-dirty
  at: 2026-10-13T05:31:54Z
scope:
  - internal/setup/codex/**
  - plugins/codex/**
confirmed_commit: dc8a789739da2f94fe9274590cdd0d9223976c35
---

# Codex integration

Applies when changing the Codex plugin, Codex hook behavior, or its setup and release path. The integration is a Codex CLI plugin; prep projects do not need to add `AGENTS.md`.

- **Package** (`plugins/codex`): `.codex-plugin/plugin.json`, `skills/prep/SKILL.md` (generated from `internal/cli/skill.md` by `just sync-skill`, then extended with Codex-specific attribution notes), and `hooks/hooks.json` calling the shared `prep hook` commands. The repo marketplace is `.agents/plugins/marketplace.json`; its `source.path` points to `./plugins/codex`. Codex plugins can bundle hooks, and the manifest names their config with `hooks`.
- **Context and steering**: on `SessionStart`, the hook finds the nearest `.prep/` project and runs `prep prime --hook`, adding its brief output as developer context. On every `UserPromptSubmit` in a prep project, it adds a short instruction to look up existing work with `prep list --text`, check `prep next`, and run `prep guide <id>` before coding. Global `hooks.prompt` defaults to small; large adds bounded guides for full issue IDs and off skips prompt context. See [Shared lifecycle hooks](/components/hooks.md). This is reliable context injection when the hook is trusted; it still depends on the agent following the guidance for ambiguous tasks.
- **Records and activity**: `PreToolUse` blocks direct edits to `.prep` through supported Bash and `apply_patch` paths, without approving or rewriting normal shell calls. Session context instructs the agent to set `PREP_SESSION` and `PREP_ACTOR=codex` only on prep CLI invocations. Unrelated shell commands retain their usual form so approved command prefixes can match. Ordinary tool events take actor and session from the hook event payload, without attribution environment assignments on those commands. `PostToolUse` runs `prep hook after-tool` to append supported tool events; prep CLI calls are omitted because prep records them itself. Codex hooks do not expose the token usage/cost data available to the Claude Code bridge. Hook errors fail open; intentional direct-record denials are the exception. Tool coverage has documented gaps, so this is a guardrail, not a complete enforcement boundary.
- **Turn finish**: Stop runs `prep hook finish`, comparing turn snapshots to request issue/knowledge reconciliation when needed. `stop_hook_active` prevents loops; Git/snapshot failures fail open.
- **Installation** (`internal/setup/codex`): `prep setup` detects the `codex` CLI and installs `prep@prep` from the marketplace at the binary's release tag, sparsely checking out `.agents/plugins` and `plugins/codex`. Refresh moves that pin; plugins installed from the same prep marketplace are reinstalled afterward, with failures reported as warnings. Removal keeps the marketplace when another plugin from it remains. For development builds, `PREP_CODEX_PLUGIN_SOURCE` points to a prep checkout containing the marketplace. Plugin hook definitions must be reviewed and trusted in Codex (`/hooks`) before they execute; installing the plugin alone does not grant hook trust. Start a new session after install. This is a local Codex CLI integration; cloud-orchestrated ChatGPT Work does not run command hooks.
## Compatibility and live verification

Keep `hooks/prep_hook.py` as a fail-open compatibility launcher for sessions that loaded the older Python commands before a plugin refresh. It forwards old session/prompt/before-tool/after-tool modes to the shared CLI and suppresses missing-binary, timeout and nonzero-exit failures. Removing it while a session still refers to it makes Python exit 2, which Codex interprets as an intentional tool denial. `TestCodexLegacyHookLauncher` checks forwarding and error behavior.

A restart alone does not trust changed hook definitions: Codex stores trust against each definition hash. Review the updated commands in `/hooks`, including the new Stop hook. A live run with old trust hashes can change code while producing no prep activity or turn snapshots; that does not validate the integration. Check actual session-attributed activity and snapshots, issue progress and knowledge outcomes, not only whether the model completed its code task.
Attribution regressions cover prep-only session guidance (including quoted/empty session IDs), prep read/write actor and session with explicit --by precedence, ordinary tool attribution without env assignments, unchanged before-tool permissions, and embedded/packaged skill consistency including the sync-skill generator. New guidance reaches sessions through the updated binary and skill; already injected session instructions are not replaced retroactively.
