- 2026-10-17T07:34:19Z edited by claude-code: depends_on

- 2026-10-17T07:34:29Z edited by claude-code: requirement

- 2026-10-17T09:06:07Z claude-code/opus-5.5: Flaky tests seen 2026-10-17: TestLiveRunner (internal/eval/matrix) failed on GitHub CI run 39018950350 with fork/exec of the prep binary it had just built (passed on the next two runs; likely text file busy); TestStatuses/timeout (internal/eval/run) and TestEvalMutate time out under heavy local CPU load. Make them deterministic with this issue.
