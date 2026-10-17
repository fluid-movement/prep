Measured on 2026-10-17 in this repository after the removals (248 issues, 303 knowledge records): `prep knowledge list` 56,472 characters, `prep list` 32,869 (it prints done and dropped issues too), `prep knowledge find <words>` 7,400 to 7,800 per query, `prep show <parent>` 7,683, `prep help` 4,224, `prime` 2,765, `guide` 2,405, `check` 56. The rethink removals already fixed check and most of help; the remaining cost is lists that grow with the project.

Targets (each default says in one closing line how to see more, with the exact flag):
- `prep list` without a state filter shows unresolved issues only and ends with '<n> resolved issues hidden; --state all shows them' (`internal/cli/read.go`, the list command, and `domain.ParseFilter` in `internal/domain` if the default filter lives there). Saved views and explicit `--state` filters are unchanged.
- `prep knowledge list` shows the 30 newest records (non-superseded first) and ends with '<n> more; --max <n> or --all'; `--all` shows everything (`internal/cli/knowread.go`).
- `prep knowledge find` keeps its ranking but prints at most 5 entries and at most about 3,000 characters in total, cutting a long block with '…', and ends with '<n> more matches; --max <n>' when it cut (`internal/cli/knowread.go`, `domain.FindKnowledge` in `internal/domain/knowread.go`).
- `prep show <parent>` lists children compactly as today; leave it unless a child list exceeds 30 lines, then show the unresolved ones and '<n> resolved children hidden; prep list --under <id> --state all'.
- `help` and `guide`: no change in this issue.
- Tests: each bound and its closing line, and that the flag in the line shows the rest. Update golden outputs if CLI goldens exist.
- Eval cases that read lists (`eval/cases/scale/*`, `eval/cases/session/*`) may assert on output; check they still load and note in the log which ones a later run should cover. Do not run them.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`, and the measurements above repeated with the new binary (put them in the log). Do not run `prep eval run`.
