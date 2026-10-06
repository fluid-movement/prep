## D1: Filters intersect with the tab
date: 2026-10-06

A filter narrows what the tab shows. Concatenating flags into one ParseFilter call would OR repeated flags (--state in the tab and in the filter) and widen the result, so the two queries run separately and intersect. Alternative: replace the tab query with the filter (loses the view the user chose).

## D2: Check runs on demand with drift
date: 2026-10-06

The check screen needs knowledge drift, which calls git and costs about 170 ms per load. Running it only while the screen is open keeps live reload of the issue views at about 10 ms. Alternative: always load with drift (slower reloads for a screen rarely open), or never show drift in the TUI (hides warnings prep check reports).

## D3: Own diff rendering, not a markdown code block
date: 2026-10-06

A diff fenced as markdown would be colored by Glamour's syntax highlighter, outside the design tokens. A small LCS line diff rendered by a ui component keeps the colors in the theme. Requirements are short, so an O(n·m) diff is fine.
