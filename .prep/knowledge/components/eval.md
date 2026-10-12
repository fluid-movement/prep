---
type: component
title: Agent behavior tests (prep eval)
description: internal/eval — fixtures, isolated environment, runner, assertions, judge, matrix, results, variants and mutation for testing agents' use of prep; layout of eval/ and how to run it.
generated:
  by: claude-code
  at: 2026-10-12T08:52:42Z
scope:
  - internal/eval/**
---

Applies when changing the agent behavior test suite (01M4QKTQVR3J6PJZAJM0MVWQR0).

- **Fixtures** (`internal/eval/fixture`): a YAML spec (`Spec`, or `WorkspaceSpec` for several linked projects) is built by running the real prep binary as a subprocess with `--by fixture`, so a fixture is valid by construction and reaches issue states through the real transitions. Issues have aliases; `{{alias}}` and `{{project/alias}}` expand to IDs in bodies, contexts, criteria and knowledge. Determinism comes from two test-only environment variables read in `cli.Main` (`testSeams`): `PREP_TEST_NOW` (RFC3339 clock) and `PREP_TEST_SEED` (integer seed for the random bits of IDs); the builder advances the clock one second per call. A workspace has its own HOME and registry (`XDG_STATE_HOME`) under its directory and builds projects in link order. The optional `repo` section makes a git repository with stub files and fixed commit dates.
