---
type: decision
title: Appends are single small writes
description: Each event is one JSON line written with `O_APPEND` in one write under 4 KiB (strings clipped, oversize refused), and the file is rewritten to its newest lines past 1 MiB.
date: "2026-10-17"
scope:
  - internal/activity
---

Each event is one JSON line written with `O_APPEND` in one write under 4 KiB (strings clipped, oversize refused), and the file is rewritten to its newest lines past 1 MiB.

So concurrent agents never interleave lines.
