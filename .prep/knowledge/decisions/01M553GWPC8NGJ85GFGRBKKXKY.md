---
type: decision
title: Live filter bar parses list flags
description: The TUI filter bar takes `prep list` flags plus bare words (joined into one text phrase), filters live on every parseable edit, and a word starting with `-` is always a flag, never a title search; flag values come from `Tree.QueryFlags`, the same source as `prep flags`.
date: "2026-10-17"
scope:
  - internal/tui
  - internal/domain
---

So live filtering does not jump on the first `-`; one value source keeps TUI and CLI in step.
