---
type: decision
title: "Mouse capture is a per-user setting; the mouse is not documented in the TUI"
description: "Mouse capture is a per-user setting; the mouse is not documented in the TUI"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
supersedes:
  - 01M48HBFHMPKZY4MN7C646QAVX
---

Whether the TUI captures the mouse depends on the person and their terminal, not on the project. So it lives in `~/.config/prep/config.yaml` (`tui.mouse`, default on), not in the committed `.prep/project.md`. The settings screen shows it in its own group, labeled as a setting for this user. The CLI loads and saves it through `Options` callbacks (`Mouse bool`, `SaveMouse func(bool) error`), so `internal/tui` does not import `internal/userconfig`. `View` sets `MouseMode` to `tea.MouseModeCellMotion` or `tea.MouseModeNone` from the model, so a toggle applies on the next render.

The `?` keymap and the footer stay about keys (the user's call): clicking does what pointing suggests, and listing it, or the terminal-specific modifier for text selection, would crowd the keymap. Turning capture off in the settings is the documented way to get plain drag selection back.
