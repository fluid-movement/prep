---
type: decision
title: Every run is isolated
description: Each run gets its own HOME, Claude config, registry and PATH with a pinned binary copy; Leaks compares init against the variant.
date: "2026-10-17"
scope:
  - internal/eval/env/**
---

Each run gets its own HOME, Claude config, registry and PATH with a pinned binary copy; Leaks compares init against the variant.

(implied) no contamination of runs or the user's setup
