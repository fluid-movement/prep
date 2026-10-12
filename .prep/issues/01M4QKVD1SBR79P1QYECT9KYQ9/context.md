Content in `eval/cases/<area>/` and `eval/fixtures/` ([case format](01M4QKVBSB5FWC25MM7CB8RBNM)); command in `internal/cli/eval.go` (started in [Results store](01M4QKVCQ4B5JC3ZF4RPME94AA)). The behaviors to cover are the skill's rules in `internal/cli/skill.md` and the step contracts from `prep guide` ([Issue lifecycle](/features/lifecycle.md), [Agent context](/features/agent-context.md), [Cross-project work](/features/cross-project.md)). The existing ledger analysis 01M4J8780KZRKWHJR4NN8AD85M names where agents went around prep; start from its findings.

- **Areas** (directories): `session`, `define`, `enrich`, `implement`, `issues`, `knowledge-read`, `knowledge-write`, `links`, `queries`, `lifecycle-other`. Prefer a fixture per area shared by its cases.
- **Each case**: a natural prompt in the user's voice (including vague ones), not a command script; assertions that pin the behavior (for example `not_ran` of `Read .prep/**` and `ran` of `knowledge find`).
- **Commands**: `run`, `list`, `doctor` as `eval` subcommands; `--models`, `--variants`, `--n`, `--parallel`, `--smoke`; `just eval-smoke` runs the smoke subset.
- **Baseline**: a first run on `current` with sonnet, N=5, stored; failures that reveal real guidance gaps become issues under the hooks issue 01M4J8789E81QAJHG8R8ZPEMRD.
- **Tests**: case files load and validate in `go test` (no agent); live runs only by command.
