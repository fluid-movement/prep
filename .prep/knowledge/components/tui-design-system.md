---
type: component
title: TUI design system
description: internal/tui/theme tokens and text styles, internal/tui/ui components, layout helpers, the gallery and golden snapshots; how to change looks or add a component.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/tui/theme
  - internal/tui/ui
---

# TUI design system

Applies when changing how the TUI looks or adding a screen or component. Screens compose components; they never set colors or styles themselves.

- **Tokens** (`internal/tui/theme/theme.go`): semantic colors (text, muted, subtle, accent, border, focus, selection, success, warning, error), one color per issue state and per kind. Each is a true-color, 256-color and 16-color value for dark and light backgrounds (`lipgloss.CompleteAdaptiveColor`). Change a value in the palette block, never at a use site. Spacing tokens `Gap`, `Pad`, `Section`.
- **Theme**: `theme.New(renderer)` binds tokens and text styles (title, heading, body, muted, subtle, code, key, key description) to one `lipgloss.Renderer`. Always build styles from `t.R.NewStyle()` or `t.S.*`, never the lipgloss globals, so the color profile and background follow the renderer.
- **Components** (`internal/tui/ui`): pure functions or small value types that take the theme and plain values and return strings: `StateBadge`, `KindTag`, `Progress`, `Note` with a `Tone`, `Tabs`, `ListRow`, `Pane` (title in the top border, focus color, body clipped to size), `KeyHelp`, `Empty`, `Loading`, `Error`, `Markdown` (Glamour with a style derived from the tokens), `Fit` (escape-safe truncation). Fixed widths `BadgeWidth` and `KindWidth` keep list columns aligned. Interactive widgets come from Bubbles and are styled from the theme.
- **Layout**: `Split` divides a width or height with minimums and collapses to one pane when both do not fit; `Stack` gives the body height between header and footer. Screens must work at 80×24.
- **Gallery**: `ui.Gallery(theme, width)` shows every token and component variant; `prep tui --gallery` displays it full-screen and reflows on resize. Check every visual change there.
- **Snapshots**: `TestGallerySnapshots` renders the gallery at 80 and 120 columns, dark and light, with a fixed true-color renderer, and compares with `internal/tui/ui/testdata/*.golden`; it also fails on any line wider than the width. After an intended change run `go test ./internal/tui/... -update`.
- **Adding a component**: write it in `internal/tui/ui` taking `*theme.Theme`, add every variant to `Gallery`, run the gallery, update the goldens.

Related: [Stack](/decisions/stack.md), [Architecture](/components/architecture.md), [CLI](/components/cli.md).
