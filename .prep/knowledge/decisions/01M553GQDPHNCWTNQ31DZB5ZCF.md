---
type: decision
title: Registry is machine-local state outside user config
description: projects.json lives in XDG state, not in the user configuration.
date: "2026-10-17"
scope:
  - internal/projects
  - internal/userconfig
---

It is machine-local state, not a choice. Damaged file is treated as empty and rebuilt.
