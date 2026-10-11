Code paths:
- `internal/mdstore/records.go`: `projectFM` gains `id` and `name` (after `schema`); `parseProject` fills `domain.Project.ID` and `.Name`; `renderProject` takes the frontmatter. `DefaultProject(id, name)` for `prep init`. A store method writes the identity into an existing project.md, keeping its body (`SetIdentity`, beside `SetSchema` in `write.go`). mdstore stays free of git: the CLI computes the ID and the derived name and passes them in.
- `internal/domain`: `Project.ID`, `Project.Name`; `Validate` reports P008 (no `id:`, warning, fixable by `prep fix`) and P009 (invalid `id:` or `name:`, error, manual). `ValidProjectName`.
- `internal/gitx`: `Remote(dir)` reads origin's URL.
- `internal/projects`: `NormalizeRemote` (`host/owner/repo`), `DefaultName(root)` (origin's last path segment, else the directory name, sanitized to a valid name), `Describe(root)` (ID and name from project.md through `mdstore`, remote from git). The registry rows carry `id`, `name`, `remote`, refreshed from the checkout in `List`; `Collision` marks a name shared by several IDs. `Add(dir)` and `Remove(query)`.
- `internal/cli`: `cmdInit` writes the identity and takes `--new-id` (on an initialized project: replaces only the ID); `cmdFix` writes a missing ID and name; `cmdProjects` gains `add <dir>` and `remove <dir|name>` and prints ID suffix, name, remote.

Knowledge: [Project registry](/components/projects.md), [Markdown store](/components/markdown-store.md), [Storage format](/conventions/storage-format.md), [CLI](/components/cli.md).
