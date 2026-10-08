---
title: TUI in a short terminal
kind: code
parent: 01M4EA4RB500QF2449Y3914VMQ
tags:
  - tui
---

The TUI stays useful at low heights, so a person can run it in a strip above or below the agent instead of beside it. With the two-tier navigation, each Agent view takes one pane.

- Below a height threshold the two navigation tiers share one line.
- Agent views and issue lists stay legible down to about 12 rows. Below that they degrade without breaking; the smoke test in `internal/tui/sizes_test.go` already guards against panics.
- Golden snapshots at a short size, 110×12, for each main screen.

## Open questions
