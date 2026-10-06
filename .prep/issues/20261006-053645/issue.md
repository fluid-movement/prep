---
title: 'TUI hierarchy: tree mode, parent focus, navigable relations'
kind: code
parent: 20261005-152616
---

Parents and children are visible and navigable in the TUI, so the user can see where an issue sits and move along the hierarchy without leaving the screen.

- Tree mode, on by default and toggled with t: the list shows children indented under their parent with tree lines, siblings grouped together. When a view matches a child but not its parent, the parent appears dimmed as context, like prep list --tree; tab counts still count only matching issues.
- Parent focus: pressing right on a parent shows only that parent and its descendants within the current view, with a breadcrumb in the list pane title (All › Export); left or esc goes back up one level. Nested parents can be focused in turn.
- Jump to parent: p selects the selected issue's parent, switching to the All view when the current view does not show it.
- Navigable relations: the detail pane lists the issue's parent, children, dependencies and the issues it blocks as rows with state and title; in the detail, tab and shift+tab move between them and enter opens the chosen issue. Backspace returns to the previously shown issue.
- The footer shows the keys that apply to the focused pane.

## Open questions
