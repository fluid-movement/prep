Builds on [Repeated runs and statistics](01M4QKVCJ49YSM62B70ZENBYNB) (the model is already a matrix dimension; cells carry the resolved id) and [Results store and comparison](01M4QKVCQ4B5JC3ZF4RPME94AA); new files `internal/eval/models` and the `prep eval models <run-id>` subcommand in `internal/cli/eval.go`. `prep eval run --models haiku,sonnet,opus` is the extended run; with `--models` omitted a case uses its own `models` or one default.

- **Resolved id**: taken from the `system`/`init` event's `model` in the stream ([Transcript](01M4QKVBMEVANE3EBQM2V2VKR8)); the matrix groups by the requested name but the report prints the resolved id and warns when the id differs between runs of one cell.
- **Threshold**: `sufficient(cell) = rate >= 0.9 && ci.Lower >= 0.8`; model order for "cheapest" is by mean cost per passing run, not by name.
- **Failure breakdown**: from `CellResult.FailedBy`, grouped by assertion kind (tool, outcome, judged), with up to 3 example run directories each.
- **Output**: terminal table plus `--json` and a markdown export (`--md`) suitable for committing as `eval/models.md`.
- **Tests**: scripted cell results where a cheaper model passes in one area and fails in another; "not enough data" with small N; alias drift warning.
