Code paths:
- `internal/tui/projects.go`: `Switch`, the header name (`projectName`), the switcher (`modalProjects`: `openProjects`, `projectCandidates`, `projectKey`, `projectRows`), `switchProject` (ends the program with where to go).
- `internal/tui/app.go`: `Options.Project`, `Root`, `Projects`, `Previous`, `Select`; `Run` returns `*Switch`; `P` opens the switcher (keymap, not the footer).
- `internal/tui/nav.go`: the header draws the project's name and marks it clickable (`project`); `goBack` switches to `Options.Previous` when the back stack is empty.
- `internal/tui/mouse.go`, `actions.go`: clicks and the wheel in the switcher, like the move picker.
- `internal/cli/tui.go`: `runTUI(root, agent, previous, sel)` and the loop in `cmdTUI` that starts over on the chosen project (touching it in the registry). `prep tui --project <name>` works through the global flag.

Knowledge: [TUI navigation](/components/tui-navigation.md), [Cross-project work](/features/cross-project.md).
