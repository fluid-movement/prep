---
type: decision
title: Tool activity is reported only by after-tool
description: prep hook after-tool is the single reporter of tool events; the bridge reports only token usage, context fill and cost; prep-only calls are omitted because prep records them itself. (merges codex.md, hooks.md)
date: "2026-10-17"
scope:
  - plugins/**/hooks/**
  - internal/cli/hook*.go
---

prep records its own commands (stated); single source avoids double counting (implied).
