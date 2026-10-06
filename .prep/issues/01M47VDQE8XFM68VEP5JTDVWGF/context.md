Extends the issue views in `internal/tui/app.go` ([TUI](/components/tui.md)); looks from the [TUI design system](/components/tui-design-system.md).

- Tree rows reuse the domain helpers `prep list --tree` uses (`internal/cli/read.go`): `Tree.WithAncestors` adds context ancestors, `Tree.TreeOrder` orders depth first, depth counts ancestors present in the set. Rows that only provide context (not in the query result) are dimmed and cannot be the reason a tab count grows.
- Each tab keeps its rows as a slice of display rows (issue ID, depth, context flag, tree prefix) instead of bare IDs; cursor, offset and selection-by-ID work on those rows.
- Parent focus: a per-tab scope stack of parent IDs. A scoped tab queries its flags plus `Under` the scope (`Filter.Under`, `Tree.Descendants`) and shows the scoped parent as the first row. The breadcrumb joins the tab name and the scope titles with ›.
- Design system: `ui.Row` gains a tree prefix and a context (dimmed) variant; `ui.ListRow` renders them; the gallery shows both and the goldens are updated.
- Relations in the detail: rendered with `ui.ListRow` (relation as the note: parent, child, depends on, blocks) above the markdown viewport, at most a few lines with a window around the selected one; `detailMarkdown` drops its relation list. A back stack of issue IDs serves backspace.
- Jumping to an issue (p, enter on a relation, backspace) selects it in the current tab when it is a row there, else switches to All, clearing that tab's scope.
- Tests in `internal/tui/app_test.go` extend the fixture (a nested parent) and add screen snapshots for tree mode and parent focus.
