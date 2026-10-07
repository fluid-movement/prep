---
type: component
title: TUI filter bar
description: 'internal/tui filter bar: query flags and words per tab, and the picker that suggests flags and their values while typing.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-07T07:43:51Z
scope:
  - internal/tui/complete.go
  - internal/tui/app.go
confirmed_commit: 9d81b630261b8fdbd508474e4cf57fa2c73230f9
---

# TUI filter bar

Applies when changing how the issue list is filtered in `prep tui` or what the filter bar suggests.

- `f` opens a text input (`/` too, unlisted, for US-keyboard habits: it is a chord on German layouts); `parseFilterText` reads `prep list` flags (which take a value comes from `domain.FlagKind`) and joins bare words into one `--text` phrase. While typing, a picker under the input offers candidates (`complete.go`, a pure function of the text and cursor): flag names after `--`, a list flag's values after the flag (from `Tree.QueryFlags`, read when the bar opens; prefix matches first, then substring, the title counts for `--under`; comma lists complete the last part and skip listed values), up to six as `ui.Suggestion` rows with counts. ↑/↓ move the highlight, tab inserts it, a click on one inserts it; enter applies, esc clears. Bare words, booleans and `--text` get no candidates; the knowledge filter has none. The filter's query and the tab's query run separately and are intersected, so a filter only narrows. Filters are per tab, shown in the pane title; esc clears.
- Values come from [Domain](/components/domain.md) `Tree.QueryFlags`, the same source as `prep flags` ([CLI](/components/cli.md)); clickable candidates are described in [TUI mouse](/components/tui-mouse.md).
