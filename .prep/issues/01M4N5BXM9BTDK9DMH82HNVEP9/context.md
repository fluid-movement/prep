Code paths:
- `internal/tui/app.go`: `relations` adds cross-project dependencies (and knowledge links); `relationBlock` renders them through `xDepLine`.
- `internal/tui/projects.go`: `xDepLine` (`← needs <link> › <ID>` with the state and title from that project, warning tone while unresolved; missing and unknown ones in the subtle tone with the clone hint), `followElsewhere` (a found dependency switches to its project on the issue; others flash why not), `linkRoot`.
- `internal/cli/tui.go`: the TUI loader fills `Tree.External`, so `Blocked` (the list's blocked marks) counts cross-project blockers; `runTUI` passes `Switch.Issue`/`Entry` on.

Knowledge: [TUI detail links](/components/tui-detail.md), [Cross-project work](/features/cross-project.md).
