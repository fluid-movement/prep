---
title: 'Variants: hook, skill and prompt overlays installed into the fixture'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVBFKQKCQP6N42BG2K8Z4
  - 01M4QKVBYFPQT50HP41EHZ6F4X
tags:
  - testing
---

## Requirement

A variant is the harness configuration under test: the plugin (skill text, hooks, hook modules) the agent runs with. Variants live in `eval/variants/<name>/` as a small `variant.yaml` plus any files it replaces. `current` is the plugin as shipped (`plugins/claude-code`), `none` is no prep integration at all (the baseline that shows what the integration adds), and others are derived from a base by replacing, removing or adding files, and optionally appending text to the system prompt.

The runner materializes a variant into a plugin directory inside the isolated environment and loads it with `--plugin-dir`. Variants are recorded in results by name and a content hash, so a result always says exactly what ran.

## Open questions
