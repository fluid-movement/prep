---
type: finding
title: Cloud environment setup script gives a 1:1 prep install
description: 'Verified in a cloud session: install.sh plus prep setup --harness claude-code gives binary, plugin, hooks, skill and activity bridge.'
date: "2026-11-18"
---

# Cloud environment setup script gives a 1:1 prep install

Verified 2026-11-18 in a Claude Code cloud session (Claude Code 2.1.327, Ubuntu, root, `CLAUDE_CODE_REMOTE=true`) on a custom environment with network access Limited and package managers allowed, against the daypack repository with prep v0.3.0.

## The setup script

```bash
#!/bin/bash
export PREP_VERSION=v0.3.0
curl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh || true
echo "claude on PATH: $(command -v claude || echo no)"
"$HOME/.local/bin/prep" setup --harness claude-code || true
echo "prep version: $("$HOME/.local/bin/prep" version 2>&1)"
```

Plain `prep setup` cannot be used: it opens /dev/tty for its checklist and a setup script has no terminal. `--harness claude-code` is the non-interactive form.

## Evidence

- `claude` is on the PATH during the setup phase (`/opt/node22/bin/claude`), so `prep setup` detects the harness and installs the plugin.
- The plugin installed inside the VM loads in the session: `prep@prep 0.3.0`, scope user, enabled, from marketplace `fluid-movement/prep@v0.3.0`. This contradicts the earlier assumption that cloud sessions never run plugins; the documented limit covers plugins a repository declares, not ones the setup script installs into the user scope.
- SessionStart hook: the briefing arrived before the first message. UserPromptSubmit guidance arrived with prompts. Bash tool calls appear as activity events.
- The activity bridge (`register.ts`) works: `prep activity` lists the session's events including model requests.
- The skill loads as `prep:prep` only; the repository's copy is not listed separately.
- Not verified: PreToolUse specifically, and the cause of one `request` event with a blank value and one `hook skipped` event.

## Consequences

- The earlier plain-file hook fallback and the project SessionStart hook route are not needed for this environment. They stay relevant for collaborators who do not use the same environment.
- The script is per account and per environment, so it is invisible to the repository.
- The environment caches the filesystem after setup, rebuilding when the script changes or after about seven days. A pinned `PREP_VERSION` keeps the cloud binary and plugin at one version.
- A repository whose `.prep` schema is newer than the cloud binary is not checked; a version check is a follow-up.
