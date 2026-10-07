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
confirmed_commit: 7731bc5a24bcf259be0467375aa28efc21f182be
---

# TUI filter bar

Applies when changing how the issue list is filtered in `prep tui` or what the filter bar suggests.

- `f` opens a text input (`/` too, unlisted, for US-keyboard habits: it is a chord on German layouts); `parseFilterText` reads `prep list` flags (which take a value comes from `domain.FlagKind`) and joins bare words into one `--text` phrase; a word starting with `-` is a flag, perhaps half typed, never a title search (so live filtering does not jump on the first `-`). While typing, a picker under the input offers candidates (`complete.go`, a pure function of the text and cursor): flag names after `--`, a list flag's values after the flag (from `Tree.QueryFlags`, read when the bar opens; prefix matches first, then substring, the title counts for `--under`; comma lists complete the last part and skip listed values; a typed `not-` or `!` stays in front of the completed value, whose count is then hidden), up to six as `ui.Suggestion` rows with counts. ↑/↓ move the highlight, enter or tab inserts it, a click on one inserts it; enter with no candidates showing applies (or shows the parse error), esc restores the filter the tab had when the bar opened (`m.before`; esc in the list still clears). Live: every edit (typing, paste, inserting a candidate) runs `filterChanged`, which sets the tab's filter whenever the text parses (empty text: no filter) and otherwise keeps the last rows without an error; the pane title shows the live query. A value typed in full has no candidates (so the next enter applies); a flag typed in full still offers itself (its insert adds the space). Typing clears the error line. After a finished flag (a value, or a switch) and a space, the flags not used yet are offered (`finished`, `nextFlags`; `--under` and `--text` can repeat and come last); after a bare word, a flag still waiting for its value, or on empty input nothing is offered, so enter applies. Bare words, booleans and `--text` get no candidates; the knowledge filter has none. The filter's query and the tab's query run separately and are intersected, so a filter only narrows. Filters are per tab, shown in the pane title; esc clears.
- Values come from [Domain](/components/domain.md) `Tree.QueryFlags`, the same source as `prep flags` ([CLI](/components/cli.md)); clickable candidates are described in [TUI mouse](/components/tui-mouse.md).
