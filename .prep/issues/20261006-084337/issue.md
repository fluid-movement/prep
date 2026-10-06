---
title: Mouse support in the TUI
kind: code
parent: 20261005-152616
tags:
  - tui
---

The TUI is keyboard-only apart from scrolling the detail pane. Mouse users can click a tab to switch views, click an issue row to select it, double-click (or click again) to open its detail, click relation rows in the detail to follow them, click in a pane to focus it, and scroll the list and the detail with the wheel. Keyboard behavior stays unchanged; the mouse is an additional way in.

## Open questions

- Should action menu entries and dialog buttons be clickable too, or only navigation?
- Should text selection in the terminal keep working (mouse capture blocks it in many terminals unless a modifier key is held)?
