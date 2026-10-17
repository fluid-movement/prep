Survey on 2026-10-17 (172 items: 50 commands, 85 flags, 16 record types, 15 TUI action groups, 6 hook paths, checked against the activity streams of all six prep projects and the Claude Code transcripts, plus keep decisions). After the rethink removals almost nothing is unused. Delete exactly this, with the tests, fixtures and docs that exist only for it:
- Flags: the `--tags` aliases of `--tag` on `prep new` and `prep edit` (`internal/cli/write.go`, help text in `internal/cli/cli.go`).
- Record layout: the `attachments/` entry in the allowed issue-directory names in `internal/mdstore/store.go` (nothing writes it; prep check should then report an unexpected `attachments/` like any other unknown name).
- Dead code (golang.org/x/tools/cmd/deadcode, unreachable even from tests): `withNext` (`internal/cli/cli.go`), `HasErrors` (`internal/domain/diag.go`), `Center` (`internal/tui/ui/modal.go`), and in `internal/gitx/git.go` `ChangedBetween`, `LastCommit`, `IsAncestor`, `Dirty`, `Head`; test helpers `newHarnessNoBootstrap` (`internal/cli/cli_test.go`) and `harness.rename` (`internal/cli/projects_test.go`).
- Used only by their own tests: `ChangedSince` and `pathspec` (`internal/gitx/git.go`), `ByID` (`internal/projects/projects.go`), `Unregister` (`internal/setup/setup.go`), `ULIDTime` (`internal/domain/ulid.go`): delete them and the tests that only test them. Keep `fixture.Build` (test fixture builder).
- The dead contract corpus `testdata/contract/` (no Go test reads it; `just test` is `go test ./...`). Remove references to it (justfile comment 'including the contract corpus', README, AGENTS.md).
Keep everything else, including the `--note` flags, `activity --actor`, eval subcommands and flags (keep decision 01M529ANGZY6HXDCN13EGPQKYJ), recovery commands (reopen, release, init --new-id) and setup/update flags.
Do not touch TOKEN-EFFICIENCY.md (the user decides).

Finish with `go run golang.org/x/tools/cmd/deadcode@latest ./...` reporting nothing unreachable (test-only uses of kept helpers are fine), and put its output in the log.

Checks before stopping: `go test ./...`, `go vet ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`. Do not run `prep eval run`.
