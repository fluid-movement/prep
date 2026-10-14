---
type: component
title: TUI navigation
description: 'prep tui navigation: Inbox, Issues, Agent and Knowledge screens, view keys, clicks, back stack and short layout.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T18:04:10Z
scope:
  - internal/tui/nav.go
  - internal/tui/ui/nav.go
confirmed_commit: 3a9c5801f8e077e2e96c5f82abe24bce2025705a
---

# TUI navigation

Applies when adding a screen or a view, or changing how people move around [prep tui](/components/tui.md). Decided in 01M4EA4RB500QF2449Y3914VMQ.

- **Two tiers** (`nav.go`): the first line is the screens bar, the project's name (`Options.Project`, else `prep`; plain body text, never the accent, since it is not a place; a click opens the project switcher), then `ui.Nav` over `navScreens` (Inbox, Issues, Agent, Knowledge). Both tiers share one grammar (01M4EDFAWDFFX96Y4ZQCKS9BQ3): plain labels, the current one accent and bold, the others muted; no underline, chips, separators or status dots (the Agent screen's second tier shows the agent's liveliness itself). Its right end (`utilities`) holds the check's `✕ n ▲ n` counts and a `⚙`. Check and Settings are screens but not tier-one entries. The second line is the current screen's views (`views`, `activeView`, `setView`) drawn with `ui.Tabs`: the saved issue lists on Issues, Issue · Activity · Usage on Agent (with the agent's label and status at the right). Screens without views (Inbox, Knowledge, Check, Settings) show a muted caption there, such as the entry count. A new screen's views go through these three functions, so keys and clicks need no new code.
- **Projects** (`projects.go`, 01M4N5BMA0V0WEB4RM7X9XRGC6): `P` or a click on the name opens the switcher (`modalProjects`): the registry's projects (`Options.Projects`) newest first, filtered as you type, the current one marked; the wheel scrolls it, a click selects. Choosing another ends the program with a `Switch`; `prep tui` (`cli.runTUI`) starts over on that root with the old one as `Options.Previous`, so backspace with an empty back stack returns there. `Options.Select` opens on an issue.
- **Keys**: `tab`/`shift+tab` cycle and 1–9 pick the second tier on every screen (Settings keeps its digits for its rows). Screens: `u` Inbox, `i` Issues from any other screen (priority is `p`), `w` Agent, `b` Knowledge, `c` check, `s` settings; a screen's key on that screen returns to Issues (`toggleScreen`). esc leaves Agent, Check and Settings for Issues; on Knowledge it first clears the filter; on Inbox it first returns from the review/detail to its list. Inbox has no second-tier views: enter focuses the detail, `o` opens the ordinary issue while preserving Inbox selection/detail focus in the back stack, and `y` copies its issue ID.
- **Back stack**: `back []place` (screen, selected issue, focus, knowledge entry; at most 100). `goScreen` pushes before every screen change, as do jumps that remember (`jump(id, true)`: links, the parent), the Agent screen's enter and a knowledge entry's issue link. Backspace anywhere, or the mouse's back button (`tea.MouseBackward`, where the terminal reports it; ignored while a dialog or the filter bar is open), runs `goBack` (there is no crumb in the bar): the screen and its selection come back, so Agent → issue detail → backspace is the Agent screen again.
- **Clicks** ([TUI mouse](/components/tui-mouse.md)): `screen:n` goes to that screen, `check` (the counts) opens the check, `settings` (the gear) toggles settings, `tab:n` picks view n of the current screen.
- **Spacing** ([TUI design system](/components/tui-design-system.md)): both tiers and the footer sit inside the `SpaceM` gutter (`gutter`), so their text starts in the column where pane text starts; tab labels line up with it since `ui.Tabs` pads them by `SpaceS`. `prep` and the screens are `SpaceL` apart, views `SpaceS` apart inside their `SpaceS` padding, the check's two counts `SpaceM` apart. From `tallHeight` (30 rows) a blank `Row` sets the navigation off from the body. The Agent status at the second tier's right drops the agent's position, then its label, before it would crowd the views (`agentStatus(detail)`).
- **Short terminals**: below `shortHeight` (16 rows) both tiers share the first line (`… Knowledge › Issue │ Activity │ Usage`), so the TUI fits in a strip above or below the agent. Wide and short keeps the side-by-side split, since only the height is short. `headerH` gives the navigation's height to `bodyHeight` and the click offsets.
- **Tests**: `TestClickTabsRowsAndRelations` (bar, gear, tabs, backspace), `TestAgentScreenKeys` (views, backspace across screens, `i`), `TestScreenKeys`, `TestShortTerminal` (goldens `short-*-110x12`).
Inbox key, mouse/back, selection and rendering tests are in `inbox_test.go`.
