---
type: component
title: Agent behavior tests (prep eval)
description: Fake eval setup records explicitly reported fixture observations and uses a synthetic stable work scope; those reports are not independent implementation checks.
generated:
  by: claude-code
  at: 2026-10-12T08:52:42Z
scope:
  - internal/eval/**
confirmed_commit: 28f150511800c7e93ff7bcd5e1ed4f499c1f8651
---

# Agent behavior tests (prep eval)

Applies when changing the agent behavior test suite or anything that runs agents against prep to measure them. The suite builds fake prep projects, runs an agent session on a prompt in one of them, and checks expected tool use and outcomes. It is deterministic wherever possible and uses an agent judge only for what cannot be checked mechanically. Results compare variants of hooks and skill, models, and later other harnesses.

Layout: code in `internal/eval/*`, content in `eval/` (`cases/<area>/*.yaml`, `fixtures/*.yaml`, `variants/<name>/variant.yaml`, `judge/calibration/*.yaml`, `results/` git-ignored, `baseline.json` once created). Commands: `prep eval run|list|doctor|report|compare|models|errors|check|judge-check|prune|clean` (`internal/cli/eval.go`, `eval_run.go`; `--dir` names the eval directory, default `./eval`). Needs a credential, because a private Claude config directory has no login: `just eval-login` once, or `CLAUDE_CODE_OAUTH_TOKEN` / `ANTHROPIC_API_KEY` in the environment. Evals run locally only, using `just eval-smoke` or `just eval-full`.

- **Fixtures** (`fixture`): YAML specs are built by running the real prep binary as a subprocess with `--by fixture`, so fixtures reach issue states through real transitions. Issues have aliases; aliases expand in bodies, contexts, criteria and knowledge. Determinism uses test-only `PREP_TEST_NOW` and `PREP_TEST_SEED`. Workspaces have their own HOME and registry; optional repo stubs, bootstrap issues and generated synthetic issues can be configured. Knowledge entries take a scope.
- **Environment** (`env`): each run gets its own HOME, Claude config, registry and PATH with a pinned copy of the binary. Auth, TERM, LANG and TMPDIR come from the caller. `Leaks` compares the session init event with the variant; `Probe` runs one trivial session for `prep eval doctor`.
- **Transcript** (`transcript`, `transcript/claudecode`): Claude stream-json is parsed into a harness-neutral transcript with steps, prep commands, metadata and counts. Another harness needs only an adapter.
- **Cases** (`evalcase`): prompt, fixture (inline or by reference), expectations (`tools`, `outcome`, `rubric`) and limits; loaded and validated completely. Assertion kinds cover tool use and outcomes such as issue state, unchanged files, knowledge, checks, claims, repositories and bounded issue/entry creation.
- **Runner** (`run`): builds the workspace, runs claude headless with case limits and returns a record with status (completed, timeout, budget, turns, error); a harness failure is an error status, never a failed assertion. A result with `is_error` true is an error even when its subtype is `success`: a session that ends with a harness message such as "You've hit your session limit" is recorded as an error with that message, not as a completed run whose assertions failed. `Release(pass)` keeps a failing run's workspace.
- **Judge** (`judge`): one call per rubric item, pinned model, no tools, JSON verdict with evidence, one retry, errors never count as passes; `Calibrate` measures agreement with hand-labelled samples and marks results untrusted below a threshold (`prep eval judge-check`).
- **Matrix** (`matrix`): cells are case × variant × model, each run N times with bounded parallelism and a seeded order; a cell reduces to a pass rate with a Wilson interval, failed assertions, flakiness and cost; harness errors are excluded from the rate and the report shows the errored count next to it; optional sequential stop; `--max-usd` stops cleanly. `Estimator` predicts cost from earlier runs.
- **Results** (`results`): `eval/results/<run-id>/` with matrix, cells and per-run files; `Report`, `Compare` (a difference counts only when intervals do not overlap), baselines (`prep eval check`).
- **Variants** (`variant`): `current`, `none`, or a directory with a `variant.yaml`; materialized with a content hash recorded in results.
- **Models** (`models`): `prep eval models <run>...` pools cases per area, names the cheapest sufficient model by cost per passing run, says "not enough data" with the number of runs needed, lists what cheaper models get wrong, and warns when a model alias resolved to different IDs.
- **What the cases ask**: see [Agent behavior test cases](/components/eval-cases.md).
- **Not covered yet**: other harnesses and a live baseline run, which needs a token.

Fixtures complete issues using checked criteria, the kind's result and a documentation decision. They model the current transition requirements and do not record retired completion contracts or scoped review artifacts. Live agent evaluation must supply its own project checks and documentation decisions.

Kept run environments (workspace, private Claude config and agent edits) live outside the checkout under the system temporary directory in `prep-eval-work-*`. `prep eval prune --keep N` removes environments recorded in results it deletes, including passing runs retained with `--keep`, and removes empty workspace parents. `prep eval clean` removes all `prep-eval-work-*` directories older than 24 hours; `clean --all` includes recent directories too. `clean` searches the current system temporary directory (`TMPDIR` where set); stop active eval runs before using `clean --all` or cleaning a run lasting over a day. Cleaning workspaces leaves stored transcripts and results in place, but their workspace paths will no longer exist.
