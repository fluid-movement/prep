## D1: Loader injected, no storage imports in the TUI
date: 2026-10-06

The TUI gets a load function and a directory to watch from the CLI, so it uses exactly the CLI's loading path and stays independent of the markdown and OKF adapters. Alternative: the TUI opens the stores itself (duplicates opening logic, couples it to the adapters).

## D2: Narrow terminals show one pane
date: 2026-10-06

Below the width where list and detail both meet their minimums, the focused pane takes the whole body and enter/esc switch between them, instead of squeezing two unreadable panes. Alternative: stack the panes vertically (the detail gets too few lines at 24 rows).

## D3: Copy with OSC 52
date: 2026-10-06

OSC 52 works in most modern terminals and over SSH without a clipboard tool. Alternative: shelling out to pbcopy or xclip (platform-specific, fails over SSH).
