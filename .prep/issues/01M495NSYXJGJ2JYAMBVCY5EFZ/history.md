- 2026-10-06T18:26:31Z edited by claude-code/opus-5.5: title, requirement

- 2026-10-06T18:34:28Z claude-code/opus-5.5: Criterion 3 reworded: most TUI tests exercise tab mechanics (digit switching, per-tab filters, reordering in settings) and need several tabs, so sample() writes six views of its own (sampleViews) and the TUI goldens stay unchanged. TestDefaultViews covers what a freshly initialized project shows. Gallery goldens regenerated because the gallery's sample tabs now lead with Unresolved and All.

- 2026-10-06T18:34:28Z claude-code/opus-5.5: The default views are documented in /components/markdown-store.md, not /components/tui.md: adding them to the TUI entry pushed it over the 8 KiB limit (K004).
