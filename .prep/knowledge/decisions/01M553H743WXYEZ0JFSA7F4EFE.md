---
type: decision
title: Per-machine state lives in .prep/local
description: Machine-specific state (activity stream, write lock) goes in a git-ignored local directory.
date: "2026-10-17"
scope:
  - internal/mdstore
  - internal/activity
---

(implied) it must not be versioned or shared between clones
