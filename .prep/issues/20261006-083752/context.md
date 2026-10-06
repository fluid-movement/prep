Knowledge: [Domain package](/components/domain.md), [Storage format](/conventions/storage-format.md), [Markdown store](/components/markdown-store.md), [CLI](/components/cli.md), [TUI](/components/tui.md), [Stack](/decisions/stack.md) (standard library first: no ULID dependency needed).

- Generation: `Stamp` and `nextStamp` in `internal/domain/fsm.go`; `PlanNew` calls `nextStamp(now, taken)` for the issue ID. Baseline names also use `nextStamp` (per issue, compared with `latest`); they are not issue IDs and stay timestamps.
- Validation: `idPattern` and `ValidID` in `internal/domain/tree.go` (`^\d{8}-\d{6}$`); `mdstore` uses `ValidID` for directory names (I001 message in `store.go` names the YYYYMMDD-HHMMSS shape) and for `issueIDs`.
- Schema: `domain.SchemaVersion` is 1 (`internal/domain/types.go`); `prep migrate` (`cmdMigrate` in `internal/cli/write.go`) runs `migrations[v-1]` for each step and sets the schema in project.md; P002 reports a project on another schema. Schema 2 adds the ID migration. The store needs a rename operation for issue directories (writes today are per file, with compare-and-swap).
- Ordering: `Tree.IDs` sorts strings; ULIDs sort by creation time as strings, so string order stays correct once every ID is a ULID.
- Resolution: `Tree.Resolve` already accepts a full ID or any unique suffix.
- Display: `shortID` in `internal/tui/app.go` takes the part after the last `-` (the time); for a ULID it takes the last 6 characters. Commit subjects from the TUI (`internal/cli/tui.go`) use the ID.
- Tests: the fixture clock makes IDs stable; ULIDs need an injectable random source so tests and goldens stay deterministic. Every fixture, golden and contract tree uses timestamp IDs today and has to move to ULIDs (`testdata/contract`, `internal/tui/testdata`).
- After migrating this repository, IDs in past commit messages no longer resolve; that is accepted. Notes outside .prep that name IDs (handoff.md, the user's memory) are not rewritten.
