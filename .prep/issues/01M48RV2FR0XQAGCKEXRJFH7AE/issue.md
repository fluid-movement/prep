---
title: Install prep in cloud sessions through a project SessionStart hook
kind: code
tags:
  - adoption
---

Cloud sessions (Claude Code on the web) load no user plugins, so the plugin's SessionStart hook never runs there and agents fall back to hand-editing `.prep`. The research in 01M46PWX38AZN55Q6T8TE93959 found that a cloud session has network access to GitHub release downloads and proxy.golang.org, and has `~/.local/bin` on the PATH; Claude Code documents project SessionStart hooks (`.claude/settings.json` in the repository) as the way to prepare cloud sessions.

`prep setup` gains a project-scoped option that writes a SessionStart hook into the repository (`.claude/settings.json` plus a small script under `.claude/hooks/`) which, only when `CLAUDE_CODE_REMOTE` is `true` and `prep` is missing, installs the release pinned to the binary that wrote the hook (install.sh with `PREP_VERSION`, checksum-verified, into `~/.local/bin`), then runs `prep prime --hook` so the session gets the same briefing as a local one. When the release cannot be fetched it tries `go install github.com/fluid-movement/prep/cmd/prep@<version>` if Go is available, and otherwise prints one line saying the binary is missing, so the agent knows it is in the fallback. The hook must never fail the session start.

## Open questions

- Should `prep update` and `prep setup --refresh` move the pinned version in the project hook, and how does a project see that its pin is older than the local binary?
- Does a copy of the skill still belong in `.claude/skills/prep` once the hook installs the binary, or does `prep prime --hook` cover it?
