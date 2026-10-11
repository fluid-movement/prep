## D1: Switching restarts the program on the other project
date: 2026-10-11

Every TUI dependency (loader, writer, watcher, activity stream, check) is bound to one root when `prep tui` starts. Rebinding them inside the running model would touch each of them; ending the Bubble Tea program with a `Switch` and starting a new one on the chosen root reuses the startup path unchanged. The new program gets the old root as `Previous`, so backspace with an empty back stack returns there, and `Select` to open on an issue (used by cross-project links). The key is `P` (project; `p` is priority), in the keymap but not the footer, which has no room left.
