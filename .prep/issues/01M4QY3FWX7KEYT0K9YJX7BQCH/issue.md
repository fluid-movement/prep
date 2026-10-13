---
title: 'Mutation run: which skill rules can be dropped or shortened?'
kind: research
parent: 01M498VQD4Q85CZYK965WW3H9Q
depends_on:
  - 01M4QY3ERYD6RCACH6EEBGXFNE
  - 01M4QY2YM2T9X68RHQW6HTCH53
tags:
  - testing
  - tokens
---

Run `prep eval mutate` on the shipped plugin and find which rules of the skill, and which hook, can go without changing agent behavior, and which changes make it better. Done when the findings give, per mutant, the change in pass rate (beyond the noise or not) and in cost, and a list of proposed edits to `internal/cli/skill.md` and the hooks.

## Open questions
