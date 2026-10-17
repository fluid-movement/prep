---
type: decision
title: "Evals gate prep's own changes, run locally, harness as a flag"
description: "Evals gate prep's own changes, run locally, harness as a flag"
date: "2026-10-16"
issue: 01M4ZCMHFA382AG6Q8T9MXFVQT
---

User decisions, 2026-10-16. Evals are the gate for prep itself: any change to gates, hooks or guidance comes with an eval case that fails without it. Evals run only locally, when behavior of the system changed and needs testing; this private project keeps runs to the minimum. Remove the CI eval workflow. Fix first: harness errors (such as an account session limit) must be recorded as errors, never as failed assertions. Rewrite the cases for the new workflow: remove cases for removed features; add the handoff case (a cheap model implements a ready issue from prep guide alone), the decision-respect case and the implementer-logs-checker-completes case. Keep fixtures, runner, assertions, variants, reports, the judge and the model axis (which model is enough for subagent implementation). Drop mutate. Delete the one-off Python pilot in eval/experiments; its findings stay on its issue. Harnesses: the suite takes the harness as a flag and runs fully with it; only Claude Code is implemented now, Codex eval support is dropped for now.
