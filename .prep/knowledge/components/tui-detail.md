---
type: component
title: TUI detail links
description: 'The issue detail''s links above its document: breadcrumb, a children heading for any number of children, dependencies and knowledge, and link mode (o) for following them.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-09T09:59:18Z
scope:
  - internal/tui/app.go
  - internal/tui/editing.go
confirmed_commit: fab664b10a77353f6b423e14435c482bd7b444ba
---

# TUI detail links

Applies when changing what the issue detail shows above its document, or how its links are followed. The list, tabs and keys are in [TUI](/components/tui.md); clicks in [TUI mouse](/components/tui-mouse.md).

- The detail shows links by shape above the document (`relationBlock`, no label column): the ancestors as a breadcrumb (`↑ Release 0.1.0 › TUI …`), the children in one heading for any number of them (01M4G1BS7D78QHNNQE9963B259: `Children` with the progress, how many are open and `list all ›`, a link that opens the children view, through `listChildren`; the first link in link mode) over up to `maxChildren` (4) unresolved children as a tree (`├─`/`└─`), in progress first (`childRank`), then by priority, resolved ones only in the progress, then `← needs` for dependencies (in the warning tone while unresolved) and `→ unblocks` for issues that depend on this one; each link uses the list's row grammar (`ui.LinkLine`: glyph, ID, progress or note, priority, title).
- The block shows up to `maxRelations` lines, then `… n more · o selects`.
- `o` enters link mode (`openLinks`, `linkKey`; 01M48KB2N0W7FZNTG77283QZJ8): each link line shows its key (`linkKeys`) and the selected one sits on the selection background; j/k or arrows move (wrapping like menus), enter or a link's key follows it (`o 3` still jumps), esc or o leave, other keys leave and act. `linkTargets` orders the links as drawn: the parent (the breadcrumb line stands for it), children, needs, unblocks, knowledge; the block scrolls (`… n above`) to keep the selection visible. Outside link mode the block looks as before; backspace returns through the back stack. Tab never moves between links: it switches views in the list and the detail alike.
- Tests: `TestManyChildren` (the heading for 30+ children, following it, backspace, a click), `TestMenusMoveWithJK`, `TestLinkModeSelectsInPlace`.
- **Children view** (`Model.childView`, `childFrom`; a `tab` with `transient` set): a view after the configured ones named `↳ <parent title>`, counting the parent's subtree. It is unfiltered (no view flags, so a link followed from a filtered view such as Unresolved still shows every child), focused on the parent (`→` narrows further inside it, `scopeTop` reports the parent), and orders the parent's children by state, open first and resolved last, each with its subtree (`byState`, `stateRank`; stable, so priority order holds within a state). Switching to another view closes it (`closeChildren`); `←`, `h` or `esc` at its top closes it and returns to the view it was opened from, on the parent; backspace goes back as anywhere.
