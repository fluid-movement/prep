Spans `internal/eval/matrix` ([Repeated runs](01M4QKVCJ49YSM62B70ZENBYNB)), the command in `internal/cli/eval.go` ([Initial suite](01M4QKVD1SBR79P1QYECT9KYQ9)), the `justfile`, and `.github/workflows/` (release workflow exists, see [Release, install and update](/components/release.md)).

- **Estimate**: `matrix.Estimate(cells, n, history)` reads `cells.json` of recent results for mean cost per cell; falls back to a per-model default table in code.
- **Limit**: `Options.MaxUSD`; the pool stops scheduling when spent + estimated in-flight would exceed it; the run per-run `--max-budget-usd` is the hard guard per session.
- **Baseline check**: `prep eval check --against <run-id|baseline.json>` exits nonzero when a cell's interval lies fully below the baseline's, used by CI; `eval/baseline.json` (committed, small) holds the smoke subset's pass rates.
- **Recipes**: `just eval-smoke`, `just eval-full`; workflow `eval.yml` (path filter for smoke, weekly cron for full, `ANTHROPIC_API_KEY` secret, results artifact).
- **Tests**: estimate and limit logic with a fake runner; check exit codes against fixture baselines. The workflow itself is validated by reading, not run.
