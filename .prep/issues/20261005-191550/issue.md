---
title: 'TUI theming: choose and customize themes'
kind: code
parent: 20261005-152616
---

Users choose how the TUI looks. The design system's palette becomes one theme among several: prep ships a few built-in themes and users can define their own by overriding tokens (text, muted, subtle, accent, border, selection, success, warning, error, issue states and kinds) for dark and light backgrounds. Components keep reading tokens from the theme, so a theme changes every screen and the gallery at once. NO_COLOR and terminals without color support keep working.

## Open questions

- Where does the theme choice live: user-level config (themes are personal; ~/.config/prep), the project's config.yaml, an environment variable, or several with a precedence?
- Which built-in themes ship first (for example default, high contrast, a light-only and a dark-only one)?
- Can a custom theme override a subset of tokens and inherit the rest from a built-in one?
- Is switching themes inside a running TUI needed, or is choosing at start enough?
