New package `internal/eval/results`, and the first `prep eval` subcommands in `internal/cli` (new file `eval.go`, command table entry in `internal/cli/cli.go`, documented in [CLI](/components/cli.md)). Input is `matrix.CellResult` ([Repeated runs](01M4QKVCJ49YSM62B70ZENBYNB)).

- **Layout**: `eval/results/<run-id>/matrix.json`, `cells.json`, `runs/<case>/<variant>/<model>/<n>/{record.json,stream.jsonl,transcript.json,results.json}`. `run-id` is the UTC time plus a short hash. `eval/results/` goes into `.gitignore`; a small `summary.json` may be committed by hand when a result is worth keeping.
- **Report**: plain table to the terminal using the same width-aware helpers the CLI uses; `--json` emits the cell results. `--failures` lists failed assertions with their reasons and the run directories to inspect.
- **Compare**: for each shared cell, `delta = rateB - rateA`; `significant` when the Wilson intervals do not overlap; cost deltas as percentages. Output groups improvements, regressions, unchanged and noisy.
- **Garbage**: `prep eval prune --keep N`.
- **Tests**: round trip store/load, report golden output, compare with overlapping and separated intervals.
