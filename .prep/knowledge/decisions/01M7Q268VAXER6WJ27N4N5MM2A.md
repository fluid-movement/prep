---
type: decision
title: Cloud sessions get prep through the environment setup script
description: Custom cloud environment setup script installs the pinned release and runs prep setup --harness claude-code.
date: "2026-11-18"
---

# Cloud sessions get prep through the environment setup script

The user's call: to run prep in Claude Code cloud sessions, a custom cloud environment carries a setup script that installs the pinned release with install.sh and runs `prep setup --harness claude-code`. The result is the same integration as a local `prep setup`: binary, plugin, hooks, skill and activity bridge.

## Why

- One code path: local and cloud both run `prep setup`, so the integration cannot drift between them, and hook logic lives in the binary.
- The environment caches the result, so session start does not pay for the install.
- It needs no per-project files, which keeps repositories clean.
- Verified working in a real cloud session; see the finding on the environment setup script.

## Trade-offs accepted

- The script lives in the user's claude.ai environment, not in the repository, so other contributors need their own environment or a committed SessionStart hook. That is not needed while the user is the only one using it.
- The pinned version must be moved when prep is updated.

This supersedes the recommendation to install through a project SessionStart hook as the primary route; that route stays the fallback for people without such an environment.
