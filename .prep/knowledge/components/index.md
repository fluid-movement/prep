# components

## Entries

* [Architecture](/components/architecture.md) - One Go binary; the domain layer sits between agent/human interfaces and storage ports with markdown and OKF adapters.
* [Claude Code integration](/components/claude-code.md) - Project-local SessionStart hook running prep prime and the /prep-status, /prep-next and /prep-guide slash commands; read-only, no FSM logic.
* [CLI](/components/cli.md) - internal/cli — command table, global flags, JSON output and errors, write pipeline, git staging and commit modes.
* [Domain package](/components/domain.md) - internal/domain — issue model, state derivation, FSM gates, validation codes, query engine, guide and knowledge retrieval.
* [Markdown store](/components/markdown-store.md) - internal/mdstore — parses and renders .prep files canonically, atomic writes with compare-and-swap, fmt and fix.
* [OKF store](/components/okf-store.md) - internal/okf — reads the knowledge bundle (links, drift) and writes entries for prep knowledge new, update and confirm.
* [TUI design system](/components/tui-design-system.md) - internal/tui/theme tokens and text styles, internal/tui/ui components, layout helpers, the gallery and golden snapshots; how to change looks or add a component.
* [TUI](/components/tui.md) - prep tui screens in internal/tui — the issue views (tabs, list, detail), loader injection, live reload with fsnotify, keys, and how screens are tested.
