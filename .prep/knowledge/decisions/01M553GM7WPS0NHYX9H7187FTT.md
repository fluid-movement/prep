---
type: decision
title: Hooks fail open
description: Missing project, malformed events, missing binary, git or snapshot errors, and a failed bridge send let the call through; only deliberate direct-record denials block. (merges claude-code.md, codex.md)
date: "2026-10-17"
scope:
  - internal/cli/hook*.go
  - plugins/**
---

Missing project, malformed events, missing binary, git or snapshot errors, and a failed bridge send let the call through; only deliberate direct-record denials block. (merges claude-code.md, codex.md)

(implied) a prep fault must not stop the agent
