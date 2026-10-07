---
title: 'TUI theming: choose and customize themes'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
depends_on:
  - 01M4AHPY6HZ0ZFMMH349V1V2TC
tags:
  - tui
  - theming
---

Users choose how the TUI looks. The design system's palette becomes one theme among several: prep ships a few built-in themes and users define their own in their `.prep/config.yaml` by setting tokens (text, muted, subtle, accent, border, selection, success, warning, error, issue states and kinds) for dark and light backgrounds. Components keep reading tokens from the theme, so a theme changes every screen and the gallery at once. NO_COLOR and terminals without color support keep working.

- The theme choice and custom themes live in the per-user project config (see the per-user configuration issue), so each person picks their own.
- `prep tui --gallery` stays the place to inspect every component and token, and gains a theme switch: cycling through the built-in and custom themes re-renders all components in the gallery at once, so a theme can be judged before choosing it.
- A command prefills the config file with a custom theme that lists every token with its current value (from the active or a named built-in theme), so users can see and edit all tokens without reading the source.

## Open questions

- Which built-in themes ship first (for example default, high contrast, a light-only and a dark-only one)?
- Can a custom theme override a subset of tokens and inherit the rest from a built-in one?
- Is switching themes inside the running issue TUI needed too, or is the gallery switch plus choosing in config enough?
- Name of the prefill command: `prep theme new <name> [--from <theme>]`, or a flag on `prep tui`?
