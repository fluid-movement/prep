## D1: Theme from a renderer, components as pure render functions
date: 2026-10-05

Building the theme from a lipgloss renderer instead of the global default makes snapshot tests deterministic and keeps the color profile right when output is not stdout. Components take values and return strings, like presentational components in a frontend: screens own state, components own looks. Alternative: Bubble Tea models per component (needed only for interactive widgets; those come from Bubbles and are styled from the theme).

## D2: Golden snapshots include color
date: 2026-10-05

Goldens are rendered with a fixed true-color profile, so a changed token shows up as a snapshot diff, not only layout changes. Reviewing a diff of escape codes is noisy, so the gallery is the place to look; the test only guards against unintended change. Alternative: ASCII-only goldens (readable, but blind to color).
