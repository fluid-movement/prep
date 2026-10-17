---
type: decision
title: Hooks never touch ordinary permissions
description: Normal calls get no permission decision, are never rewritten or auto-approved; Codex env assignments only on prep CLI invocations.
date: "2026-10-17"
scope:
  - internal/cli/hook*.go
  - plugins/codex/**
---

Normal calls get no permission decision, are never rewritten or auto-approved; Codex env assignments only on prep CLI invocations.

Blanket inline env assignments stop approved command prefixes from matching (stated).
