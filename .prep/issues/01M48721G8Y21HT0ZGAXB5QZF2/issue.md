---
title: prep setup, harness interface and the user configuration
kind: code
parent: 01M486FES0NQA47N0EH8T3NWCD
tags:
  - install
  - integration
---

prep setup installs, refreshes and removes harness integrations through one interface that every harness implements: detect whether the harness is present, report what is installed and at which version, install or update to the binary's version, and remove. Claude Code is the first implementation (its plugin is a sibling issue); the interface is what later harnesses plug into.

- Interactive: prep setup lists the supported harnesses, preselects the ones detected on the machine and the ones already chosen, and applies the selection after confirmation; deselecting a harness removes its integration. Prompts read from the terminal, so prep setup works when started from curl | sh.
- Non-interactive: prep setup --harness claude-code[,...] installs exactly those, prep setup --remove claude-code removes one, and both print what they did, with --json.
- The user configuration lives in $XDG_CONFIG_HOME/prep/config.yaml (default ~/.config/prep/config.yaml) and records only the user's choices, starting with the integrated harnesses. Installed state is always asked from the harness.
- install.sh runs prep setup after installing the binary when a terminal is attached, and prints the prep setup command otherwise.
- prep update refreshes the integrations of the configured harnesses after replacing the binary, without prompting, so they carry the binary's version.

## Open questions
