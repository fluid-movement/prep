Commit mode touches every layer that handles the project configuration:

- `internal/domain/types.go`: `Config.CommitMode` and the constants `CommitOff`/`CommitAll` go; `Config` keeps `Views` and `ViewOrder`. `internal/domain/config.go`: `PlanConfig` drops the commit mode check (see [Domain](/components/domain.md)).
- `internal/mdstore/store.go`: `configFile` loses `commit_mode`, so `decodeStrict` rejects it as an unknown key (P003); `LoadConfig` drops the switch; `renderConfig` writes only `views`, and its default header is the comment lines before `views:` in `DefaultConfig`, which loses the commit mode line and comment ([Markdown store](/components/markdown-store.md), [Storage format](/conventions/storage-format.md)).
- `internal/cli/cli.go`: `record` always calls `gitx.Stage`; `gitx.Commit` in `internal/gitx/git.go` has no other caller and is removed ([CLI](/components/cli.md)).
- `internal/tui/app.go` and `internal/tui/editing.go`: the settings screen loses the commit mode row (row 0); `setIdx` then starts at the first view, space toggles only the mouse, and the key help changes ([TUI editing](/components/tui-editing.md)).
- Tests: `internal/domain/domain_test.go`, `internal/mdstore/store_test.go`, `internal/tui/app_test.go` reference commit mode; TUI goldens with the settings screen change.
- Fixtures: every `testdata/contract/*/.prep/config.yaml` drops its commit mode lines; `invalid/config-invalid` keeps failing with P003 through another invalid value (an unparsable view query).
- This repository's `.prep/config.yaml` drops the lines. README and the skill do not mention commit mode.
