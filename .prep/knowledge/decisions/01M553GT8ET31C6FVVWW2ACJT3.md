---
type: decision
title: Spacing tokens, no screen margin
description: Use SpaceS/M/L and Row, not literals; no margin around the whole screen.
date: "2026-10-17"
issue: 01M4EDFAWDFFX96Y4ZQCKS9BQ3
scope:
  - internal/tui/ui
---

Cells are not pixels; a margin costs four columns and two rows at 80x24 (01M4EDFAWDFFX96Y4ZQCKS9BQ3).
