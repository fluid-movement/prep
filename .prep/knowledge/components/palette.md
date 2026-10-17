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

Applies when changing colors or touching the two built-in themes.

- **Tokens** (`palette.Tokens`), deliberately few: the gray ramp `text`, `muted`, `subtle`, then `accent`, `selection` (a background), `success`, `warning`, `error`. There are no border, state or kind tokens (reduced in 01M4AM1CEQQPWSNCP05Y6SVD7T to keep the UI calm): the theme package draws them from these eight, so a new state or kind needs no palette change.
- **Dark first** (01M4AMC70GBW1R8YDA1TG44HKE): a theme is one palette made for dark terminals; nearly every developer runs dark. `Palette{Name, Light, Colors}`; `Light` marks the `light` built-in. `Color` is a hex value, optionally with hand-picked 256/16-color values (the `default` and `light` built-ins have them, so they render exactly as the original design system). Without them the theme package downsamples the hex value (`colorprofile.ANSI256/ANSI.Convert`); Bubble Tea’s output writer then converts everything for the terminal’s profile, and the ASCII profile (NO_COLOR) drops colors.
- **Built-ins** (`builtin.go`, display order): `default` and `light`. Selections are a step lighter than common dark terminal backgrounds (#1E–#2A); the themes do not set a background, so the selection must show on whatever the terminal uses.
- **Choosing** (`Resolve(name)`): resolves one of the built-in names. `Validate(selected)` checks the selected name. Without a configured theme, `default` is used on dark terminals and `light` on light ones; only this fallback uses the terminal’s reported background. `Builtins` lists the two names.
- **Dependencies**: the package imports only domain, so [Markdown store](/components/markdown-store.md) validates the selected built-in theme with it (load: P003; write: refused). Domain cannot import it (cycle), so `PlanConfig` does not validate themes; the store does.
- **Derived colors** (`derived.go`, `StateToken`, `KindToken`, `BorderToken`, so the mapping lives next to the tokens and needs no TUI import): borders use `BorderToken` (`subtle`); kinds use `KindToken` (`muted`, they are written out); states use `StateToken`: open `muted`, defined `text`, ready `success`, in progress `accent`, done and dropped `subtle` (glyphs tell states apart, color marks what needs attention and lets finished work fade). `internal/tui/theme` resolves them for the terminal’s profile.
- **Reaching the screen**: `theme.From(palette, profile)` in the [TUI design system](/components/tui-design-system.md); the issue TUI re-resolves on every load and background report (`syncTheme`, `retheme`), so a theme switch re-renders everything. Settings toggle directly between the two built-ins; the gallery’s `t`/`T` switches between them.
