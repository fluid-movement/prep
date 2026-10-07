---
title: Command that lists every query flag with all its values
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - cli
  - views
---

Writing a view means knowing the query flags of `prep list` and the values each accepts. A command prints every query flag together with all values currently available for it, so a user can build a view without guessing: fixed sets such as `--state` (open, defined, ready, in_progress, done, dropped), `--kind` and `--priority`, and project-dependent ones such as `--tag` (every tag in use), `--view` (the saved views) and the parents usable with `--under`/`--parent` (ID suffix and title). Boolean flags (`--stale`, `--actionable`, `--blocked`, `--leaf`, `--top`, `--tree`) are listed with what they select. `--json` gives the same as structured output.

## Open questions

- Name: a new `prep flags`, `prep list --options`, or part of `prep views`?
- Should the TUI's filter bar offer the same values (completion or a picker), or is that a separate issue?
