## D1: o enters link mode in place
date: 2026-10-07

`o` starts a link mode inside the detail pane: keys are shown on the link lines, j/k move a highlight, enter follows, esc leaves, and a link's key follows it at once. It replaces the Go to menu.

Rationale: selection happens where the links are drawn; no key changes meaning with focus (j/k mean "move" in the list and in link mode alike, and link mode is entered explicitly); `o <key>` stays the fast path. The mouse already follows links.

Alternatives: arrows in the focused detail (conflicts with scrolling the document); keeping the menu next to it (two ways to show the same list).
