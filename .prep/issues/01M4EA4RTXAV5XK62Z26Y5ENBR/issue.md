---
title: TUI in a short terminal
kind: code
parent: 01M4EA4RB500QF2449Y3914VMQ
tags:
  - tui
---

The TUI stays useful at low heights, so a person can run it in a strip above or below the agent instead of beside it. With the two-tier navigation, each view needs only one pane at a time.

- Below a height threshold, every screen shows one pane (as with z), the navigation tiers take one line each or merge into one, and the footer shrinks to the essential keys.
- Agent views and issue lists stay legible down to about 12 rows. Below that they degrade without breaking; the smoke test in `internal/tui/sizes_test.go` already guards against panics.
- Golden snapshots at a short size, such as 110×12, for each screen.

## Open questions
