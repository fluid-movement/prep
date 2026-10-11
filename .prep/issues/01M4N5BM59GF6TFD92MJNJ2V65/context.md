Code paths:
- `internal/domain/links.go`: `XKnowledgeLinks` (skips code via `StripCode`) and `LinkRefs` already find `](name:/path.md)` links in records and entries; `guide.go` adds each cross-project link of the context as an input with the command that reads it.
- `internal/projects/links.go`: `Check` reports X007 when the linked project is checked out and has no such entry file (a warning, like X005: that checkout may be behind).
- `internal/cli/read.go`: `show` prints `knows: <link>:<path> (prep knowledge show <path> --project <link>)`, JSON `cross_knowledge`.
- `internal/cli/cli.go`: `--project` accepts the home project's link names first (`projectNamed`), so the commands work with the names references use.
- `internal/tui`: the detail lists `≡ knows <link> › <path>`; following it switches to the project on the knowledge screen (`Switch.Entry`, `Options.Entry`). Links inside knowledge entry bodies stay plain text there, as local links do.

Knowledge: [Cross-project work](/features/cross-project.md), [TUI detail links](/components/tui-detail.md).
