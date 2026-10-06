## D1: A wheel scroll detaches the view from the selection until a key
date: 2026-10-06

Each list keeps its offset. A wheel notch moves the offset by three rows, the same as the viewports' default delta, and marks the list as scrolled (`m.wheeled`, or `scrolled` on the dialog). While a list is scrolled, rendering only clamps its offset to the rows and no longer follows the cursor, so the selection may sit outside the view. Any key press clears the flag, so keyboard navigation brings the selection back into view. A click selects a visible row, so it needs no snapping.

Alternative considered: a separate viewport per list. That would duplicate the row rendering the panes already do and change every list's goldens.
