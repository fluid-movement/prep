---
title: git describe builds count as development builds
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
priority: high
---

Binaries built with `just install` carry `git describe` versions such as `v0.1.0-14-gaa2b6fc` or `v0.1.0-14-gaa2b6fc-dirty`. `update.IsRelease` treats them as releases with a pre-release part, so the SessionStart briefing claims the binary is older than the v0.1.0 plugin, `prep update` would replace a development build, and `prep setup` would pin the Claude Code plugin to a tag that does not exist. Versions with a `git describe` suffix (`-<n>-g<hash>`, optionally `-dirty`) or only `-dirty` are development builds, like Go pseudo-versions.

## Open questions
