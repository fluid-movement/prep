---
title: Cloud sessions for prep
kind: manual
tags:
  - cloud
priority: high
---

Run prep work in Claude Code cloud sessions: define issues locally, hand them to a cloud session, bring the results back, with the same prep integration as a local install.

Decision: cloud sessions get prep through the environment setup script (install.sh, then prep setup --harness claude-code). Finding: that route was verified end to end in a cloud session with prep v0.3.0: binary, plugin, hooks, skill and activity bridge all work.

This parent closes when a user can hand a ready issue to a cloud session with one command, the result returns through the records, and the setup is documented.

## Open questions
