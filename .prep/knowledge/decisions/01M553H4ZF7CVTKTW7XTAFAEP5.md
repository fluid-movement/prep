---
type: decision
title: Record writes and edits never change state
description: '`context`, `decide`, `criterion`, `findings`, `log` and `edit` work on unresolved issues without moving state; only a requirement or kind edit can make a defined issue stale.'
date: "2026-10-17"
scope:
  - internal/domain
  - internal/cli
---

`context`, `decide`, `criterion`, `findings`, `log` and `edit` work on unresolved issues without moving state; only a requirement or kind edit can make a defined issue stale.

reason not stated
