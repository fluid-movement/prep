New package `internal/eval/matrix`, driving the [session runner](01M4QKVBYFPQT50HP41EHZ6F4X) and the assertions ([tool use](01M4QKVC39NZBH95845QPC1D5Q), [outcome](01M4QKVC87EPF7DT9AN3NTGF6B), [judge](01M4QKVCD5WNFWQC4J50R52Y47)).

- **Types**: `Cell{Case, Variant, Model}`; `RunResult{Record, Assertions []Result, Judged []Result, Cost}`; `CellResult{Cell, Runs, Passed, Errored, Rate, CI (Wilson interval), Flaky, FailedBy map[string]int, Cost{Calls, Errors, TokensIn, TokensOut, USD, Duration: mean and stddev}}`.
- **Scheduling**: a worker pool of `Parallel` runs (default 2; each run has its own isolated environment so they are independent); order is shuffled by a seed so a time-of-day effect does not fall on one cell; context cancellation stops cleanly and keeps finished runs.
- **Pass**: a run passes when every deterministic assertion passes and no judged item failed; errored judge items make it `unjudged`, shown separately.
- **Stop rule**: `Stop: fixed|sequential`; sequential stops when the Wilson interval is entirely above `0.9` or below `0.1` after at least 3 runs.
- **Tests**: with a fake runner returning scripted results, the rate and interval math, flaky marking, error exclusion, shuffle determinism by seed, and early stopping.
