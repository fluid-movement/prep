---
type: decision
title: Git calls never take the index lock
description: Every git call sets GIT_OPTIONAL_LOCKS=0.
date: "2026-10-17"
scope:
  - internal/gitx/**
---

Every git call sets GIT_OPTIONAL_LOCKS=0.

So reads like status from the TUI or prep check never block a commit the person runs at the same time (stated).
