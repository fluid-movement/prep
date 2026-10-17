Where things are:
- Cause: `classify` in `internal/eval/run/run.go` returns Completed for result subtype `success` (and for an empty status with steps) without looking at `tr.Final.IsError`. A session that ends with "You've hit your session limit" arrives as subtype success with is_error true, so it is scored as a completed run whose assertions failed. Check `tr.Final.IsError` before the success cases (after the timeout, cancel, turns and budget cases) and return Errored with the result text as the message.
- Parsing: `internal/eval/transcript/claudecode/parse.go` already fills `Final.IsError`; Codex transcripts, if parsed elsewhere, need the same field when their harness reports an error.
- Scoring: `matrix.RunResult.Passed` and the rates in `internal/eval/matrix/matrix.go` already exclude Errored runs; verify the report (`internal/eval/results/report.go`, `RateText`) shows the errored count next to the rate, and `internal/eval/models/models.go` keeps skipping them.
- Tests: a classify test with a success/is_error result (use the session-limit text), and a matrix or report test showing such a run excluded from the pass rate and counted as an error. `TestRatesFlakyErrorsAndCost` in `internal/eval/matrix/matrix_test.go` is the nearby pattern.
- Docs: knowledge [Agent behavior tests](/components/eval.md). Update through `prep knowledge update`.

Checks before stopping: `go test ./...`, `go build -o prep ./cmd/prep && ./prep check` (0 errors), `./prep fmt --check`.
