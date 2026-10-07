# components

## Entries

* [Architecture](/components/architecture.md) - One Go binary; the domain layer sits between agent/human interfaces and storage ports with markdown and OKF adapters.
* [Claude Code integration](/components/claude-code.md) - The prep plugin for Claude Code: skill, SessionStart briefing, /prep:status, /prep:next, /prep:guide and the side panel (/prep:focus, /prep:pane), served from this repository as its own marketplace and installed per user by prep setup.
* [CLI](/components/cli.md) - internal/cli — command table, global flags, JSON output and errors, write pipeline and git staging.
* [Domain package](/components/domain.md) - internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
* [Markdown store](/components/markdown-store.md) - internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
* [OKF store](/components/okf-store.md) - internal/okf — reads the knowledge bundle (links, drift) and writes entries for prep knowledge new, update and confirm.
* [Release, install and update](/components/release.md) - How prep is released (GoReleaser on v* tags), installed (install.sh, go install, just install) and updated (prep update with checksum verification).
* [Harness setup and user configuration](/components/setup.md) - prep setup installs, refreshes and removes harness integrations through the setup.Harness interface; ~/.config/prep/config.yaml records which harnesses the user chose.
* [TUI design system](/components/tui-design-system.md) - internal/tui/theme tokens and text styles, internal/tui/ui components, layout helpers, the gallery and golden snapshots; how to change looks or add a component.
* [TUI editing and keys](/components/tui-editing.md) - How the TUI writes and which keys do what: the write pipeline, action and edit menus, letter melodies, the keymap and ?, dialogs over the screen, inline text editing with $EDITOR hand-off, the create wizard and the editable settings.
* [TUI mouse](/components/tui-mouse.md) - How the TUI takes clicks and the wheel: elements marked as Lip Gloss layers while rendering, hit-testing, click semantics, and the per-user mouse capture setting.
* [TUI](/components/tui.md) - prep tui screens in internal/tui — the issue views (tabs, list, detail and its links), loader injection, live reload with fsnotify, navigation keys, and how screens are tested; editing is in TUI editing and keys.
