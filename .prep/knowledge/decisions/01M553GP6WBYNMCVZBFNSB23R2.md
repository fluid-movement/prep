---
type: decision
title: Store validates themes, domain does not
description: Theme names are checked in the store (load P003, write refused), not in PlanConfig.
date: "2026-10-17"
scope:
  - internal/mdstore
  - internal/palette
---

Theme names are checked in the store (load P003, write refused), not in PlanConfig.

Domain cannot import palette (import cycle). (Same source as palette.md; merged.)
