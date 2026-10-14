---
title: Investigate TUI live refresh missing external issue completion until restart
kind: research
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - tui
  - watch
  - bug
priority: high
---

Observed in the prep project: issue 01M4WTN9CB82MSZYCDTWJY2RPN was completed and its resolution committed in 33d632f; prep show reported state done and all acceptance criteria checked, while an already-running TUI still appeared to show it as unresolved. Restarting the TUI displayed the correct state. Investigate whether live refresh misses external lifecycle changes. Establish a reproducible case and record the running binary version, active project/worktree, screen and filter; distinguish stale UI state or an older binary from a watcher/timer/load failure. The current TUI uses an fsnotify watch on .prep, with separate activity timers, so do not assume polling is the cause. Inspect event delivery, directory registration, atomic record writes, reload scheduling and loaded-state application. Report findings and a focused regression/fix proposal if a defect is confirmed. This request captures investigation work; it does not authorize implementing a fix or changing other issue states.

## Open questions
