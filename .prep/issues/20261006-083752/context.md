Knowledge: [Domain package](/components/domain.md), [Storage format](/conventions/storage-format.md), [Markdown store](/components/markdown-store.md), [TUI](/components/tui.md), [Stack](/decisions/stack.md) (standard library first: no ULID dependency needed).

- Generation: `Stamp` and `nextStamp` in `internal/domain/fsm.go`; `PlanNew` calls `nextStamp(now, taken)` for the issue ID. Baseline names also use `nextStamp` (per issue, compared with `latest`); they are not issue IDs and stay timestamps.
- Validation: `idPattern` and `ValidID` in `internal/domain/tree.go` (`^\d{8}-\d{6}$`); `mdstore` uses `ValidID` for directory names (I001 message in `store.go` names the YYYYMMDD-HHMMSS shape) and for `issueIDs`.
- Ordering: `Tree.IDs` sorts strings ("chronological order"); ULIDs start with `0`/`1` and would sort before every timestamp ID, so ordering must compare creation times (a helper that decodes both formats). Check every `sort.Strings` over IDs and `TreeOrder`/`Query` tie-breaks in `internal/domain/query.go`.
- Resolution: `Tree.Resolve` already accepts a full ID or any unique suffix.
- Display: `shortID` in `internal/tui/app.go` takes the part after the last `-` (the time); for a ULID it takes the last 6 characters. CLI text output prints full IDs; commit subjects from the TUI (`internal/cli/tui.go`) use the ID.
- The fixture clock in tests makes IDs stable; ULIDs need an injectable entropy source (a reader on the tree or in `PlanNew` input) so tests and goldens stay deterministic.
- Docs: storage format (`issues/<YYYYMMDD-HHMMSS>/`), the skill (`prep skill`, plugin `skills/prep`), and the contract corpus (`testdata/contract`), which needs a valid tree with ULID and mixed IDs.
