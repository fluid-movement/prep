---
title: Mouse support in the TUI
kind: code
parent: 01M46AREE07YNNVPBA8F4MT56D
depends_on:
  - 01M48N7E9GRPWH1SKX5MTB931Y
tags:
  - tui
---

The TUI is keyboard-only apart from scrolling the detail pane. Mouse users can click a tab to switch views, click an issue row to select it, double-click (or click again) to open its detail, click relation rows in the detail to follow them, click in a pane to focus it, and scroll the list and the detail with the wheel. Action menu entries and dialog buttons are clickable too, so every element a key can activate can also be clicked. Keyboard behavior stays unchanged; the mouse is an additional way in.

Mouse capture is on by default, and the settings screen has an option to turn it off entirely, for example to select text with a plain drag. There is no runtime toggle key. The keymap and footer document keys only; mouse interactions are not listed.

## Open questions
