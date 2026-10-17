---
type: component
title: Agent behavior test cases
description: 'What the eval cases ask and why: built from real session shapes, which cases were dropped, the variants that split the plugin.'
generated:
  by: claude-code
  at: 2026-10-12T11:04:44Z
scope:
  - eval/**
confirmed_commit: 2c0c3af9f0e968721abae5a633ff084828af63c8
---

# Agent behavior test cases

Applies when adding, changing or dropping a case in `eval/cases`, or reading a result. Designed from the shapes of real sessions in other prep projects, not from the CLI's command list): `real/*` are prompts as users write them (set up the knowledge base, audit with subagents, plan what is next, a feature or bug told in the user's words); `natural/*` are tasks that never mention prep, with and without a matching issue (does the agent track it, and keep knowledge in step with the code it changed); `adversarial/*` try to break it (a claim that tests pass while the project's test fails, instructions hidden in an issue, a stale requirement, an ambiguous reference); `scale/*` repeat simple questions in a 60-issue project; `batch` has several chores in one message. Cases every model passes trivially were dropped: they cost money and show nothing. Variants `skill-only` and `hook-only` split the plugin; with `none` and `current` they answer what each part buys (`--tag plugin-ab`). `prep eval errors` lists the prep commands that failed in stored runs, which is how agents' guesses (`list --max`, `--tags`) were found.

Parent: [Agent behavior tests](/components/eval.md).
## Tracking policy and deferred model checks

Natural unmatched implementation now expects one new issue, even for a small fix. natural-fix-opt-out covers explicit no-record-write scope; natural-read-only covers questions that must create no issue. adversarial-ambiguous-reference requires clarification and unchanged candidates rather than permitting a guessed selection. This follows [Tracking implementation work](/decisions/tracking.md). No live Claude runs are currently authorized; original haiku/sonnet requirements are preserved in issue 01M4J8789E81QAJHG8R8ZPEMRD under 0.4.0. Shared mechanics and targeted Codex checks are validated first; their results do not claim other models passed.
