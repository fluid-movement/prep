---
type: decision
title: Fresh stores per load
description: The CLI builds fresh stores for each TUI load and check because they can run concurrently.
date: "2026-10-17"
scope:
  - internal/cli
  - internal/tui
---

The CLI builds fresh stores for each TUI load and check because they can run concurrently.

Stated: reloads and checks can run concurrently.
