---
title: 'Test case format: prompt, fixture, other projects, expectations'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVB5EW2N6Y9BDVCFJDC7Q
tags:
  - testing
---

## Requirement

A test case is a declarative file saying what to run and what to expect. It names the prompt (with `{{alias}}` placeholders for fixture IDs), the fixture or workspace to run in (inline or by reference), the primary project, the other projects it needs, optional model and variant restrictions, and the expectations: tool-use assertions, outcome assertions and, optionally, judge rubric items. Cases live in the repository under `eval/cases/`, one YAML file each, grouped by CLI area in directories.

Loading validates a case completely before anything runs: unknown keys, unknown aliases in prompts or expectations, references to missing fixtures, and unknown assertion kinds are errors with the file and line.

## Open questions
