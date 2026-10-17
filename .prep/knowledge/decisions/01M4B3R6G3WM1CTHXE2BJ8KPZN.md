---
type: decision
title: "prep theme new selects the theme it writes"
description: "prep theme new selects the theme it writes"
date: "2026-10-07"
issue: 01M46QWSKGQZ5S0N3RHTEE86Y1
---

After writing `themes.<name>`, `prep theme new` also sets `theme: <name>`, so the next `prep tui` or gallery shows it and edits are visible right away.

Rationale: the command's purpose is starting a custom look; a user who wants to keep the current one switches back with one line or the settings screen.

Alternatives: write only (an extra step every time), a `--use` flag (more surface for the common case).
