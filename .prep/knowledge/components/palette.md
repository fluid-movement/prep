---
type: component
title: Palette and themes
description: internal/palette — TUI color token names, the built-in themes, custom theme resolution and validation; how a theme reaches the screen.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-07T07:17:14Z
scope:
  - internal/palette
---

# Palette and themes

Applies when changing colors, adding a built-in theme or a token, or touching theme config.

- **Tokens** (`palette.Tokens`): `text`, `muted`, `subtle`, `accent`, `border`, `selection`, `success`, `warning`, `error`, then `state.<state>` for `domain.States` and `kind.<kind>` for `domain.Kinds`, in that order. A new state or kind needs a color in every built-in (the built-in test fails otherwise).
- **Palette**: a name and a `Variant` (token → `Color`) for dark and light. `Color` is a hex value, optionally with hand-picked 256/16-color values; only `default` has them, so its rendering is exactly the original design system. Without them the theme package downsamples the hex value (`colorprofile.ANSI256/ANSI.Convert`); Bubble Tea's output writer then converts everything for the terminal's profile, and the ASCII profile (NO_COLOR) drops colors.
- **Built-ins** (`builtin.go`, display order): default, high-contrast, monochrome, pastel, catppuccin (Mocha/Latte), nord (Polar Night/Snow Storm), gruvbox. Each defines every token for both backgrounds; values are listed in token order.
- **Custom themes** (`domain.ThemeDef` in the config's `themes:`): `base` (default: `default`) plus `dark`/`light` maps of token → `#RRGGBB`. `Resolve(name, custom)` walks the base chain onto a copy of the built-in and reports unknown themes, unknown bases, cycles, unknown tokens and invalid colors. `Validate(selected, custom)` checks every custom theme, refuses redefining a built-in name, and resolves the selected one. `Names` lists built-ins then custom themes by name; `Def(p)` turns a resolved palette into a full custom definition (what `prep theme new` writes).
- **Dependencies**: the package imports only domain, so [Markdown store](/components/markdown-store.md) validates config with it (load: P003; write: refused) and the [CLI](/components/cli.md) prints and writes themes without the TUI. Domain cannot import it (cycle), so `PlanConfig` does not validate themes; the store does.
- **Reaching the screen**: `theme.From(palette, dark, profile)` in the [TUI design system](/components/tui-design-system.md); the issue TUI re-resolves on every load (`syncTheme`), so a theme switch or an edited custom theme re-renders everything; the gallery's `t`/`T` cycle `Names`.
