---
title: Claude Code panel uses the prep theme
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - integration
  - theming
---

The Claude Code panel draws with Claude Code's own theme colors, while `prep tui` uses the prep theme; the same states and notes look different in the two places. The panel takes its colors from prep instead: a read command, `prep theme colors --json`, prints the active theme resolved for the user's config (the eight tokens and the colors prep derives from them: one per state, kinds, borders), and the plugin paints with those hex colors (Claude Code's Text accepts raw colors).

- The derived colors come from prep, so the panel never repeats the state mapping; one source for both clients.
- The plugin reads the colors at session start and again when `prep watch` reports a change, so a theme switch in the TUI recolors the panel.
- When the command fails (an older binary, no project), the panel keeps drawing with Claude Code's theme keys, as today.

## Open questions

- Without a configured theme prep picks default or light by the terminal background, which the plugin cannot query: should the plugin follow Claude Code's light/dark setting (if the API exposes it), or always use the configured theme or default?
- Should the plugin offer an option to keep Claude Code's colors (pane option, for example `colors: prep | claude`)?
