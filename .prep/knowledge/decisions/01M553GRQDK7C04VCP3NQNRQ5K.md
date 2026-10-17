---
type: decision
title: User config stores choices only
description: userconfig records choices (hooks.prompt, harnesses, tui.*); installed versions are always asked from the harness.
date: "2026-10-17"
scope:
  - internal/userconfig
  - internal/setup
---

Store only what cannot be derived (design principles; "a small database").
