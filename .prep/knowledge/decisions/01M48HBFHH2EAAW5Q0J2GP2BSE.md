---
type: decision
title: "Click semantics: select, click again to activate"
description: "Click semantics: select, click again to activate"
date: "2026-10-06"
issue: 01M4863WN8D1HHTP302Q8G2PQX
---

- A left press acts. Releases and motion are ignored, so `WithMouseCellMotion` stays enough.
- A click on an unselected row (issue, knowledge entry, settings row, menu entry, reparent pick) selects it. A click on the selected row does what enter does: opens the detail, runs the entry, or scrolls to the knowledge entry. There is no timed double-click: Bubble Tea v1 reports no click count, a timer is fragile over SSH, and a second click already covers it.
- A click on an unavailable menu entry shows its reason, as its key does.
- A click on a tab switches to that view. From the check, settings or knowledge screen it also returns to the issue views.
- A click on a relation line follows it as `o <n>` does and pushes the back stack, so backspace returns.
- A click inside a pane focuses that pane. The wheel scrolls the pane under the pointer, not the focused one: the list moves its selection by one row per tick, and viewports scroll by their own wheel delta.
- In dialogs, clicks hit menu entries, reparent picks, criteria (a click toggles the checkbox), the wizard's kind options and form fields (a click focuses the field). A click outside an open dialog does nothing, so it cannot throw away typed text. The help dialog closes on any click, as on any key.
- Footer key hints are not clickable. They document keys rather than being controls.
- In the narrow layout, clicks go to the pane that is showing.
