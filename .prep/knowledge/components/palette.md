---
type: component
title: Palette and themes
description: internal/palette — TUI color token names, the built-in themes, custom theme resolution and validation; how a theme reaches the screen.
generated:
  by: claude-code/opus-5.5
  at: 2026-10-07T07:17:14Z
scope:
  - internal/palette
confirmed_commit: a9b23847bce01161522f5c997cb4fa8b5b0cc393
---

# Palette and themes

Applies when changing colors, adding a built-in theme or a token, or touching theme config.

- **Tokens** (`palette.Tokens`), deliberately few: the gray ramp `text`, `muted`, `subtle`, then `accent`, `selection` (a background), `success`, `warning`, `error`. There are no border, state or kind tokens (reduced in 01M4AM1CEQQPWSNCP05Y6SVD7T to keep the UI calm): the theme package draws them from these eight, so a new state or kind needs no palette change.
- **Dark first** (01M4AMC70GBW1R8YDA1TG44HKE): a theme is one palette made for dark terminals; nearly every developer runs dark. `Palette{Name, Light, Colors}`; `Light` marks themes made for light terminals: the `light` built-in and custom themes based on it. `Color` is a hex value, optionally with hand-picked 256/16-color values (only `default` and `light` have them, so they render exactly as the original design system). Without them the theme package downsamples the hex value (`colorprofile.ANSI256/ANSI.Convert`); Bubble Tea's output writer then converts everything for the terminal's profile, and the ASCII profile (NO_COLOR) drops colors.
- **Built-ins** (`builtin.go`, display order): default, high-contrast, monochrome, pastel, catppuccin (Mocha), nord (Polar Night), gruvbox, then light. Selections are a step lighter than common dark terminal backgrounds (#1E–#2A); the themes do not set a background, so the selection must show on whatever the terminal uses.
- **Choosing** (`Pick(configured, darkTerminal)`): a configured theme is used as is; without one, `default` on dark terminals and `light` on light ones. Only this case looks at the terminal's reported background.
- **Custom themes** (`domain.ThemeDef` in the config's `themes:`): `base` (default: `default`) plus token → `#RRGGBB` pairs directly under the theme name (`Colors` is an inline YAML map). `Resolve(name, custom)` walks the base chain onto a copy of the built-in (inheriting `Light`) and reports unknown themes, unknown bases, cycles, unknown tokens and invalid colors. `Validate(selected, custom)` checks every custom theme, refuses redefining a built-in name, and resolves the selected one. `Names` lists built-ins then custom themes by name; `Def(p)` turns a resolved palette into a full custom definition with base `light` for light palettes (what `prep theme new` writes).
- **Dependencies**: the package imports only domain, so [Markdown store](/components/markdown-store.md) validates config with it (load: P003; write: refused) and the [CLI](/components/cli.md) prints and writes themes without the TUI. Domain cannot import it (cycle), so `PlanConfig` does not validate themes; the store does.
- **Derived colors** (`derived.go`, `StateToken`, `KindToken`, `BorderToken`, so the mapping lives next to the tokens and needs no TUI import): borders use `BorderToken` (`subtle`); kinds use `KindToken` (`muted`, they are written out); states use `StateToken`: open `muted`, defined `text`, ready `success`, in progress `accent`, done and dropped `subtle` (glyphs tell states apart, color marks what needs attention and lets finished work fade). `internal/tui/theme` resolves them for the terminal's profile.
- **Reaching the screen**: `theme.From(palette, profile)` in the [TUI design system](/components/tui-design-system.md); the issue TUI re-resolves on every load and background report (`syncTheme`, `retheme`), so a theme switch or an edited custom theme re-renders everything (except during a settings theme preview, which decides until enter or esc); the gallery's `t`/`T` cycle `Names`.
