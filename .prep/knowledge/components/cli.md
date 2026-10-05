---
type: component
title: CLI
description: internal/cli — command table, global flags, JSON output and errors, write pipeline, git staging and commit modes.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
scope:
  - internal/cli
  - cmd/prep
confirmed_commit: d902990b28eaba324184b9331bc0b3b243104443
---

# CLI

Applies when adding or changing commands.

- Commands are registered in `internal/cli/cli.go` with a read/write flag; read and write commands are strictly separate.
- Global flags anywhere: `--json`, `--root`, `--by` (or `PREP_ACTOR`, `PREP_ROOT`). The default actor is `cli/prep-<version>`; actors starting with `human:` cannot complete code issues.
- Under `--json` everything, including errors (`{ok:false,error:{code,message,unmet,diagnostics}}`), is JSON on stdout. Exit codes: 0 ok, 1 gate/check failure, 2 usage, 3 conflict.
- Write pipeline: load → `Plan` → `CheckWrite` → `store.Apply` → stage (commit mode `off`) or commit only `.prep` paths (`all`).
- `prep skill` prints the embedded `internal/cli/skill.md`; `.claude/skills/prep/SKILL.md` must stay identical (tested).
- Tests in `cli_test.go` drive whole lifecycles through `Main` with a fake clock.
