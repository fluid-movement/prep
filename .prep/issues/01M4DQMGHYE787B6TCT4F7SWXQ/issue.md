---
title: prep setup installs the Pi integration
kind: code
parent: 01M46ARM9G3492V298TM82KD45
depends_on:
  - 01M4DQMY0R6D02TSYS40YATN29
tags:
  - integration
  - pi
---

`prep setup` gets a Pi harness (`internal/setup/pi`, registered like `internal/setup/claudecode`, see [Harness setup](/components/setup.md)): it detects `pi` on the PATH, reads from `pi list` whether the prep package is installed and at which ref, installs `git:github.com/fluid-movement/prep@vX.Y.Z` per user (D4), moves the pin on refresh, and removes it.

- `PREP_PLUGIN_SOURCE` (or its equivalent) points development builds at the checkout as a local source; `just install` refreshes the Pi integration as it does the Claude Code plugin.
- `just release` sets the version in the root `package.json` and the release workflow refuses a tag that differs, as for the plugin manifests.
- Other Pi packages and settings are left untouched; tests use a fake `pi`.
- Verified with the real Pi CLI in an isolated agent directory: install, refresh, remove.

## Open questions
