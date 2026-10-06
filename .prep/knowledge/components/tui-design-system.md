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
confirmed_commit: 2cde71a6de8c4d95e07837a979a8277e02e8e532
---

# TUI design system

Applies when changing how the TUI looks or adding a screen or component. Screens compose components; they never set colors or styles themselves.

- **Tokens** (`internal/tui/theme/theme.go`): semantic colors (text, muted, subtle, a tangerine accent that is also the focus color, a neutral slate selection, border, success, warning, error), one color per issue state and per kind. Each is a true-color, 256-color and 16-color value for dark and light backgrounds. Change a value in the palette block, never at a use site. Spacing tokens `Gap`, `Pad`, `Section`.
- **Theme**: `theme.New(dark, profile)` resolves the tokens for a background and a `colorprofile.Profile`: the dark or light value, then `lipgloss.Complete(profile)` picks the hand-picked true-color, 256 or 16-color value (Lip Gloss v2's automatic downsampling would compute its own). It also builds the text styles (title, heading, body, muted, subtle, code, key, key description). Styles are plain values: build them from `lipgloss.NewStyle()` with theme colors, or start from `t.S.*`; never write color literals at a use site. Programs start dark with the profile detected from their output, request the background in `Init`, and on `tea.BackgroundColorMsg` or `tea.ColorProfileMsg` overwrite `*th` in place (components hold the pointer) and drop cached renders; startup never waits for the terminal.
- **Components** (`internal/tui/ui`): pure functions or small value types that take the theme and plain values and return strings: `StateBadge`, `KindTag`, `Progress`, `Note` with a `Tone`, `Overlay` (a block centered over a background whose styling is stripped and redrawn subtle; both are Lip Gloss layers composited on a canvas of the body's size, so a block taller than the body is clipped; `OverlayAt` says where the block lands)
, `PriorityMark` (`!crit` in the error tone, `!high` in warning, `low` subtle, nothing for medium; `Row.Priority` puts it before the title), `Tabs` (with `TabSpans` for where each tab is), `ListRow` (with tree lines from `TreePrefixes` and a dimmed context variant), `Link`/`LinkLine` (a detail link drawn by shape: a pre-rendered lead such as tree lines or an arrow, then the list's row grammar), `LinkRow` (compact relation link, used by the move picker: label, state glyph, ID, title), `Pane` (title in the top border, focus color, body clipped to size), `Diff` with `DiffLines` (LCS line diff, removed in error, added in success color), `DiagnosticRow` (severity, code, location, message, class and fix), `Modal` and `Center` (dialogs over the body), `MenuRow` (key, label, reason for unavailable actions), `Field` (labeled form input), `KeyHelp`, `Empty`, `Loading`, `Error`, `Markdown` (Glamour with a style derived from the tokens; Glamour v2 wraps links in OSC 8 hyperlinks, invisible in the snapshots' text), `Fit` (escape-safe truncation). Fixed widths `BadgeWidth` and `KindWidth` keep list columns aligned. Interactive widgets come from Bubbles and are styled from the theme.
- **Layout**: `Split` divides a width or height with minimums and collapses to one pane when both do not fit; `Stack` gives the body height between header and footer. Screens must work at 80×24.
- **Gallery**: `ui.Gallery(theme, width)` shows every token and component variant; `prep tui --gallery` displays it full-screen and reflows on resize. Check every visual change there.
- **Snapshots**: `TestGallerySnapshots` renders the gallery at 80 and 120 columns, dark and light, with `theme.New(dark, colorprofile.TrueColor)`, and compares
 with `internal/tui/ui/testdata/*.golden`; it also fails on any line wider than the width. After an intended change run `go test ./internal/tui/... -update`.
- **Adding a component**: write it in `internal/tui/ui` taking `*theme.Theme`, add every variant to `Gallery`, run the gallery, update the goldens.

Related: [Stack](/decisions/stack.md), [Architecture](/components/architecture.md), [CLI](/components/cli.md).
