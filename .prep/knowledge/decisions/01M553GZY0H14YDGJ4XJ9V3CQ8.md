---
type: decision
title: Live reload splits tree from activity
description: A project change reloads the issue tree; a `.prep/local` change only re-reads the activity stream; a load error keeps the last good tree and shows the error; selections are kept by issue ID.
date: "2026-10-17"
scope:
  - internal/tui
  - internal/watch
---

So a busy agent never makes the TUI reparse every issue.
