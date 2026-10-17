---
type: decision
title: Debounced, merged, never blocking
description: Events wait 150 ms because a transition writes several files; one goroutine owns pending state and a new report merges into the waiting one.
date: "2026-10-17"
scope:
  - internal/watch
---

So a slow reader never blocks the watcher.
