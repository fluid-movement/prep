## D1: Custom themes give hex colors only
date: 2026-10-07

Custom themes and the new built-ins specify true-color hex values; 256 and 16-color fallbacks come from Lip Gloss downsampling for the detected profile. Only `default` keeps hand-picked fallbacks, so its goldens do not change.

Rationale: three values per token per background would triple what users must write, and few people can pick good 16-color fallbacks. Downsampling is what Lip Gloss does for any hex color anyway.

Alternatives: require all three depths (precise, but a 138-value theme); allow optional per-depth overrides (more format for a rare need, can be added later).

## D2: Palette lives in its own package
date: 2026-10-07

Token names, built-in themes and theme resolution go into `internal/palette`, which imports nothing from the TUI. The theme package builds Lip Gloss colors and styles from a palette; mdstore validates config with it; the CLI prints and writes themes with it.

Rationale: `prep check` and `prep theme new` must know the tokens without pulling the TUI into the store or the CLI's non-TUI paths, and domain must stay free of UI packages.

Alternatives: keep everything in `internal/tui/theme` (mdstore would import a TUI package); put tokens in domain (UI vocabulary in the domain model).

## D3: prep theme new selects the theme it writes
date: 2026-10-07

After writing `themes.<name>`, `prep theme new` also sets `theme: <name>`, so the next `prep tui` or gallery shows it and edits are visible right away.

Rationale: the command's purpose is starting a custom look; a user who wants to keep the current one switches back with one line or the settings screen.

Alternatives: write only (an extra step every time), a `--use` flag (more surface for the common case).
