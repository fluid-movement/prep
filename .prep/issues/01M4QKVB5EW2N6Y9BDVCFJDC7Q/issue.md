---
title: 'Test fixtures: build a deterministic fake prep project from a spec'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
tags:
  - testing
---

A builder turns a spec (issues with states, parents, dependencies, knowledge entries, config) into a temp prep project with fixed IDs and timestamps, so a run is reproducible and a fixture reads as a few lines, not a checked-in snapshot.

## Open questions
