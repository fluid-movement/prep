---
type: decision
title: "Mouse capture is a per-user setting, applied through View.MouseMode"
description: "Mouse capture is a per-user setting, applied through View.MouseMode"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
supersedes:
  - 01M48HBFHJ5EJEFQQCYA5W4WDH
---

Whether the TUI captures the mouse depends on the person and their terminal, not on the project. So it lives in `~/.config/prep/config.yaml` (`tui.mouse`, default on), not in the committed `.prep/project.md`. The settings screen shows it in its own group, labeled as a setting for this user. The CLI loads and saves it through `Options` callbacks (`Mouse bool`, `SaveMouse func(bool) error`), so `internal/tui` does not import `internal/userconfig`. `View` sets `MouseMode` to `tea.MouseModeCellMotion` or `tea.MouseModeNone` from the model, so a toggle applies on the next render without enable or disable commands.

The selection modifier differs by terminal (shift in most of them, option in iTerm2 and Terminal.app). The keymap names both instead of detecting the terminal.
