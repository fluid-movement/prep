---
title: 'Dark-first themes: one palette per theme, a single light theme'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - tui
  - theming
---

Almost everyone runs their terminal dark, so themes are tuned for dark backgrounds only. Each built-in theme is one palette of the eight tokens: default, high-contrast, monochrome, pastel, catppuccin, nord, gruvbox. One built-in theme, `light`, serves light terminals (today's light values); light-mode users pick it or build their own on it.

- A theme no longer has dark and light variants: a custom theme is `base:` plus token colors directly under its name, and inherits whether it is light from its base.
- Without a configured theme, prep uses `default` on a dark terminal and `light` on a light one; a configured theme is always used as is.
- The selection background is clearly visible on typical dark terminal backgrounds in every theme (the default's was nearly invisible).
- The dropped state badge renders its label struck through in one color (the strikethrough was nested inside the badge style and broke the color).
- No migration: prep has no outside users.

## Open questions
