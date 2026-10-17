- 2026-10-17T13:28:01Z edited by claude-code: requirement

- 2026-10-17T13:36:02Z claude-code: Implemented, ready for check: ui.ListRow drops the ID, then the state label, then the kind until the title has 30 cells (the glyph stays); the gallery shows narrow rows, goldens regenerated and the drop order and clickable narrow rows are tested.

- 2026-10-17T13:39:56Z claude-code: Checker: columns were chosen per row from its tree lines and note, so rows of one list had different columns and misaligned glyphs; the choice now depends on the width alone (test added), so at the default split only the ID goes.
