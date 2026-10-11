Code paths:
- `desktop/Sources/PrepKit/Models.swift`: `CrossDep` (`cross_deps` of `prep show --json`; `blocks` only when found and not done), `ProjectRef.projectID`, `remote`, `collision`, `displayName`.
- `desktop/Sources/PrepDesktop/Relations.swift`: `DependencyMap` counts and draws cross-project dependencies through `blocking`; `CrossIssueCard` (link name, ID, state and title from that project; dashed and disabled with the clone hint when not checked out).
- `desktop/Sources/PrepDesktop/ProjectModel.swift`: `go(to:inProject:)` opens that project in this window on the issue (one project per window, as prep tui switches).
- `desktop/Sources/PrepDesktop/ContentView.swift`: the switcher shows `displayName`.

Knowledge: [Desktop app](/components/desktop.md).
