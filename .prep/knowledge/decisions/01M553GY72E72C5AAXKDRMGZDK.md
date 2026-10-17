---
type: decision
title: Mouse capture is the user's choice
description: Capture is `tui.mouse` in user config (unset = on), not a project setting; with it off the terminal selects text with a plain drag and the keymap/footer never list mouse actions.
date: "2026-10-17"
scope:
  - internal/tui
  - internal/userconfig
---

(implied) the terminal setup is per person, not per project.
