Knowledge: [TUI mouse](/components/tui-mouse.md), [TUI](/components/tui.md).

- `wheel` in `internal/tui/mouse.go` moves the issue list's cursor (`m.cursor[tab]`) and the knowledge list's cursor (`m.know.cursor`) one row per notch. While a dialog is open, it does nothing.
- `listPane` (`app.go`) and `knowledgeList` (`knowledge.go`) keep a per-list offset (`m.offset[tab]`, `m.know.offset`) and snap it on every render so the cursor stays visible. That snap has to give way after a wheel scroll, or the view jumps back to the selection.
- The move picker (`modalReparent` in `modalView`, `actions.go`) windows its picks around the cursor (`start := clamp(d.cursor-rows/2, …)`). It needs its own scroll offset.
