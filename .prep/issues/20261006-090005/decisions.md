## D1: Refresh through the new binary
date: 2026-10-06

prep update runs prep setup --refresh with the freshly installed binary instead of refreshing from the old process, because the new version knows the current plugin layout and harness list. Alternative: refresh from the running (old) binary, which may install integrations the new binary no longer matches.

## D2: Checklist built on the design system
date: 2026-10-06

The harness selection is a tiny Bubble Tea checklist using the TUI design system, so it looks like the rest of prep and gets keyboard handling for free; it opens the terminal device itself, which curl | sh needs. Alternative: numbered line prompts (simpler, but inconsistent and clumsy for multi-select).
