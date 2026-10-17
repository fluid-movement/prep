---
type: decision
title: Record writes always, reads only for a named actor
description: Prep commands log writes always but reads only when run with `--by` or `PREP_ACTOR`.
date: "2026-10-17"
scope:
  - internal/activity
---

Panels, hooks and people reading without a name buried the agent's own events. Focus follows prep events naming an issue except `show`, and never events with `project` or harness tool events, so a write in another project leaves the home focus alone.
