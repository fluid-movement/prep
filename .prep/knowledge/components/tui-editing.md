---
type: component
title: TUI editing and keys
description: 'How the TUI writes and which keys do what: the write pipeline, action and edit menus, letter melodies, the keymap and ?, dialogs over the screen, inline text editing with $EDITOR hand-off, the create wizard and the editable settings.'
generated:
  by: claude-code/2.1.291
  at: 2026-10-06T12:36:56Z
scope:
  - internal/tui/actions.go
  - internal/tui/editing.go
  - internal/tui/ui/modal.go
confirmed_commit: 1a26cd9183a429cd0de6da33d6eba7f1dde968a0
---

# TUI editing and keys

Applies when changing how people write from the TUI or which keys do what. Screens and navigation are in [TUI](/components/tui.md); looks come from the [TUI design system](/components/tui-design-system.md).

- **Writes** (`actions.go`): `Options.Write` takes a plan function and runs the CLI's pipeline (fresh stores, plan, `CheckWrite`, Apply, stage via `record` in `internal/cli/cli.go`), serialized by a mutex; `Options.Actor` is `human:<OS user>`, so the domain's `G_ACTOR` gate keeps code completion with agents. The TUI plans with the same domain functions as the CLI: `PlanNew`, `PlanEdit`, `PlanRecord`, `Plan`, and `PlanConfig` for settings. After a write it reloads and selects the changed or created issue. Without `Write` the TUI is read-only.
- **Action menu** (`a`): new issue, title, requirement, context, tick off criteria, move, priority, define, ready, acknowledge, drop, complete. Only actions that apply now are listed (`openMenu` drops entries whose reason from `Tree.Gates`, evaluated for the human actor with placeholder reason and no-impact arguments, is not empty; code issues never offer complete). Menu letters run actions; arrows and j/k move in every menu (no menu uses j or k as a letter; link menus number their entries with `linkKeys`: digits, then letters other than h, j, k and l), q or esc close. 
- **Keys are melodies** (decided in 01M47Y6EA0SMCZAKG03785YF76): every action is a short sequence of unshifted letters, never a chord, except ctrl+s and ctrl+e inside text fields. From list and detail: `a` the action menu, `e` the edit menu (r requirement, c context, t title; `openEditMenu`), `i` the priority picker (c h m l), `n` the create wizard. Lifecycle transitions are the agents' work; the action menu offers the ones that apply to a human (drop, completing a manual issue), and criteria are ticked off with `a v` (verify). There is no next-transition key.
- **Keymap**: each screen has one table of `binding`s (group, key, description, essential): the footer shows the essential ones (`essentials`), `?` opens `modalHelp` with all of them grouped plus the action menu's sequences for the selected issue (those that do not apply with their reason), in two columns when it does not fit; any key closes it.
- **Dialogs** float over the screen: `View` renders the screen and `ui.Overlay` composites `modalView`'s `ui.Modal` box over it, the screen behind dimmed. They use `ui.MenuRow` and `ui.Field` with Bubbles text inputs and, for long text, a Bubbles `textarea` from `newArea` (static cursor, no blink timer; inputs share `inputStyles`)
; a rejected write keeps the dialog open with the error. New issue is a wizard (`openCreate`, `createKey`, `modal.step`): 1/3 the title (enter continues, an empty title is refused), 2/3 the kind (its letter c m r d picks it and continues; arrows and enter too), 3/3 the optional requirement (ctrl+s creates, ctrl+e hands off); esc steps back and cancels on the first step. Settings views are edited in `modalViewEdit` (name, query). Move offers top level and every unresolved issue outside the issue's own subtree, filtered as you type.
- **Text editing** (`editing.go`): `openText` edits requirement or context inline in `modalText`; ctrl+s saves (unchanged text writes nothing), esc cancels, ctrl+e (`handOff`) writes the text to a temp file and `runEditor` (`tea.ExecProcess` with `$VISUAL`, `$EDITOR` or `vi`) brings the result back into the field (`areaEditedMsg`), also in the create dialog. Text from a rejected write is kept per issue and field (`pending`) and offered by the next edit. `runEditor` and `after` (timers) are hooks tests replace; tests take editor text from the hook and save with ctrl+s (`viaEditor`).
- **Settings** (`settingsPane`, `settingsKey`): selectable rows, the theme first (under "Look"; `t`/`T`, or space/enter on its row, switch to the next/previous of `palette.Names` and save `theme:` through `PlanConfig`; the reload's `syncTheme` re-renders everything, see [Palette and themes](/components/palette.md)), the views numbered like their tabs, and last, under "You", the user's mouse capture ([TUI mouse](/components/tui-mouse.md); saved in the user configuration, not the project); arrows, j/k or digits select (digit n is view n on this screen), space or enter toggles the mouse or switches the theme,
 enter edits a view, n adds one, d asks and a second d deletes, m starts move mode (`Model.moving`: j/k or arrows move the view one place per write, enter or esc ends). Each change builds a full `Config` (`settingsConfig`) and writes it through `PlanConfig`; invalid queries come back as the dialog's error; tabs follow on the reload. The schema version and the project DoD are read-only.
