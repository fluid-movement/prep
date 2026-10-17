---
type: decision
title: Watcher errors mean full reload
description: An overflow or failed watch is reported as a change of everything.
date: "2026-10-17"
scope:
  - internal/watch
---

Reload everything rather than miss changes.
