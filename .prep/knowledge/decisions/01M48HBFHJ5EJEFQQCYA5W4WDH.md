---
type: decision
title: "Mouse capture is a per-user setting in the user configuration"
description: "Mouse capture is a per-user setting in the user configuration"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
---

Whether the TUI captures the mouse depends on the person and their terminal, not on the project. So it lives in `~/.config/prep/config.yaml` (`tui.mouse`, default on), which the setup knowledge entry already names as the place for per-user TUI choices. It does not go in `.prep/project.md`, which is committed and shared. The settings screen shows it in its own group, labeled as a setting for this user, so it isn't mistaken for a project setting. The CLI loads and saves it through `Options` callbacks, so `internal/tui` does not import `internal/userconfig`. Toggling it applies at once (`tea.EnableMouseCellMotion` / `tea.DisableMouse`).

The selection modifier differs by terminal (shift in most of them, option in iTerm2 and Terminal.app). The keymap names both instead of detecting the terminal.
