---
title: prep prints the cloud environment setup script for its version
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

The cloud environment setup script lives in the user's claude.ai environment, outside any repository, so it can silently fall behind the prep version the user runs. Let prep print the script for its own version: `prep cloud script` (name open) prints the script with `PREP_VERSION` set to the running binary's release, ready to paste into the environment dialog. Development builds that have no release say so and print the script with the latest release.

The script printed is the one verified in the finding on the environment setup script: install.sh pinned, `prep setup --harness claude-code`, diagnostics, every step ending in `|| true`. There must be exactly one definition of it in the code, which the README and docs quote, so they cannot drift from what prep prints.

## Open questions
