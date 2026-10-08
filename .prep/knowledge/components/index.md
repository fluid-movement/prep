# components

## Entries

* [Architecture](/components/architecture.md) - One Go binary; the domain layer sits between agent/human interfaces and storage ports with markdown and OKF adapters.
* [Claude Code side panel](/components/claude-code-panel.md) - The prep pane in Claude Code: activation, following the agent, data, drawing, the Live, Project and Usage views, /prep:focus and /prep:pane, opening and developing.
* [Claude Code integration](/components/claude-code.md) - The prep plugin for Claude Code: skill, SessionStart briefing, /prep:status, /prep:next, /prep:guide and the side panel (/prep:focus, /prep:pane), served from this repository as its own marketplace and installed per user by prep setup.
* [CLI](/components/cli.md) - internal/cli — command table, global flags, JSON output and errors, write pipeline and git staging.
* [Domain package](/components/domain.md) - internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
* [Markdown store](/components/markdown-store.md) - internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
* [OKF store](/components/okf-store.md) - internal/okf — reads the knowledge bundle (links, drift) and writes entries for prep knowledge new, update and confirm.
* [Palette and themes](/components/palette.md) - internal/palette — TUI color token names, the built-in themes, custom theme resolution and validation; how a theme reaches the screen.
* [Release, install and update](/components/release.md) - How prep is released (GoReleaser on v* tags), installed (install.sh, go install, just install) and updated (prep update with checksum verification).
* [Harness setup and user configuration](/components/setup.md) - prep setup installs, refreshes and removes harness integrations through the setup.Harness interface; ~/.config/prep/config.yaml records which harnesses the user chose.
* [TUI Agent screen](/components/tui-agent.md) - prep tui's Agent screen: the followed agent's current issue, its activity feed by issue, its usage; keys, layout, liveliness and data from the activity stream.
* [TUI design system](/components/tui-design-system.md) - internal/tui/theme tokens and text styles, internal/tui/ui components, layout helpers, the gallery and golden snapshots; how to change looks or add a component.
* [TUI editing and keys](/components/tui-editing.md) - How the TUI writes and which keys do what: the write pipeline, action and edit menus, letter melodies, the keymap and ?, dialogs over the screen, inline text editing with $EDITOR hand-off, the create wizard and the editable settings.
* [TUI filter bar](/components/tui-filter.md) - internal/tui filter bar: query flags and words per tab, and the picker that suggests flags and their values while typing.
* [TUI knowledge screen](/components/tui-knowledge.md) - internal/tui knowledge screen: entries with attention marks, the entry view, filters and links between issues and entries.
* [TUI mouse](/components/tui-mouse.md) - How the TUI takes clicks and the wheel: elements marked as Lip Gloss layers while rendering, hit-testing, click semantics, and the per-user mouse capture setting.
* [TUI](/components/tui.md) - prep tui screens in internal/tui — the issue views (tabs, list, detail and its links), loader injection, live reload with fsnotify, navigation keys, and how screens are tested; editing is in TUI editing and keys.
