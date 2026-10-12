---
title: 'Test fixtures: build a deterministic fake prep project from a spec'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
tags:
  - testing
---

## Requirement

A fixture builder turns a declarative spec into a fake prep project in a temp directory, so agent behavior tests start from a known, reproducible state.

The builder drives the real prep CLI for everything it can: init, new, edit, define, ready, claim, knowledge new, link add and the rest. The project is valid by construction, reaches any state through real transitions, and follows format changes without maintenance. Only what the CLI cannot do (the git repo, stub source files) is written directly.

Its own calls carry a distinct actor (`--by fixture`), so the activity stream and any tool-call tracking tell setup apart from the agent under test.

The spec is YAML. Issues, knowledge entries and links have symbolic aliases (`auth-epic`, `login-bug`). A case refers to them as `{{login-bug}}` in prompts and assertions, and the builder substitutes the generated IDs. The spec covers: issues with kind, state, parent, dependencies, tags, priority, criteria and body; knowledge entries; config and views; and an optional repo section.

The repo section is optional and off by default. When present, the fixture is a git repo with a small stub codebase (a few source files), for cases where the agent must edit code or the commit mode matters.

Runs are fully deterministic: IDs and timestamps are fixed, so a prompt can name an ID and a failing run reproduces exactly. Because the builder runs the real binary as a subprocess, prep honors two test-only environment variables, a fixed clock and an ID seed. Production behavior is unchanged when they are unset.

The package lives in `internal/eval/fixture`, usable from `prep eval` and from Go tests, and may replace ad-hoc project setup in existing tests over time.

## Open questions
