---
title: 'Mutation testing: generate hook and skill variants and compare'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVCW6CMS3Z21PVCK9RA01
  - 01M4QKVCQ4B5JC3ZF4RPME94AA
  - 01M4QKVD1SBR79P1QYECT9KYQ9
tags:
  - testing
---

## Requirement

Mutation testing of the harness configuration. `prep eval mutate` takes a base variant, generates mutants of its skill and hooks, runs the selected cases on the base and every mutant, and reports which mutations raised or lowered the pass rate and what they did to cost.

Mutants come from two sources: mechanical operators (remove a skill rule, shorten a rule, move a rule to the end, remove the SessionStart briefing, disable one hook, soften or sharpen wording by replacement tables) and agent-proposed rewrites (an agent is asked for N variations with a stated hypothesis each, for example "make the rule about reading knowledge shorter and put it first"). Every mutant is an ordinary variant, written to `eval/variants/mutants/<id>/`, with its description and hypothesis recorded, so a good mutant can be promoted by hand.

The result ranks mutants against the base using the compare logic, so a difference counts only when it exceeds the noise.

## Open questions
