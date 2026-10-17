Where things are:
- The prompt hook: the `case "prompt":` branch of `cmdHook` in `internal/cli/hook.go` (around line 103). It builds `message` (the `hookReminder` text plus, at level large, up to two guides for issue IDs named in the prompt) and sends it with `hookContext` on every prompt.
- Send only what changed: keep the text it last sent per session in a small file under `.prep/local/hooks/` (next to the turn snapshots of `hookTurnPath` in `internal/cli/hook_finish.go`; key it by session the same way), and skip `hookContext` when the new message equals it. `hookMarkTurn` must still run on every prompt.
- Compaction: `prep hook session-start` re-fires after compaction and restores the briefing; it also removes this session's last-sent prompt file, so the next prompt sends its text again. A missing or unreadable file means send.
- Tests in `internal/cli/hook_test.go`: the first prompt sends the reminder; an identical second prompt sends nothing; a prompt naming a new issue at level large sends again; session start resets it; the turn snapshot is written on every prompt either way.
- Eval: the requirement asks for evidence that tracking behavior holds; the existing natural cases (`eval/cases/natural/*`, for example natural-fix-no-issue) cover it. Name the cases to rerun in the log; do not run them (live evals need the user's OK).
- Docs: none; knowledge is decision records now.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`. Do not run `prep eval run`.
