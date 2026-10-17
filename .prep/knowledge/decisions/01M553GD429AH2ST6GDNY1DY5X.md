---
type: decision
title: Harnesses are adapters, workflow logic is shared
description: Lifecycle hook behavior lives once in internal/cli/hook*.go; Claude Code and Codex provide only event adapters. (also hooks.md)
date: "2026-10-17"
scope:
  - internal/cli/hook*.go
  - plugins/**
---

(implied) one behavior for all harnesses
