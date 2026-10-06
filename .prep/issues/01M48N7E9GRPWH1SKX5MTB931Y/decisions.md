## D1: Theme from background and color profile, kept as one pointer
date: 2026-10-06

`theme.New(dark bool, profile colorprofile.Profile)` resolves each token with `lipgloss.LightDark(dark)` and `lipgloss.Complete(profile)`. This keeps the hand-picked 256- and 16-color values, where v2's automatic downsampling would compute its own. Programs start with dark and the program's detected profile. On `tea.BackgroundColorMsg` or `tea.ColorProfileMsg` they overwrite `*m.th` in place, so the components that hold the pointer pick up the new theme, and they reset render caches.

Alternatives considered: `compat.CompleteAdaptiveColor` is the shortest port, but it does blocking terminal I/O on global state outside Bubble Tea's loop, which is the startup delay we have today. Dropping the 256- and 16-color values in favor of automatic downsampling is a possible simplification for later, but it changes looks, so it is not part of the port.
