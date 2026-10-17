---
type: decision
title: TUI writes reload under the lock
description: The TUI builds a fresh store per write, holds the lock, loads again, then plans and applies.
date: "2026-10-17"
scope:
  - internal/tui
  - internal/cli
  - internal/mdstore
---

A long-running TUI must never plan on a stale screen.
