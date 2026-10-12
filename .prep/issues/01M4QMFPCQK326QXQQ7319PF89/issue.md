---
title: 'Model comparison: run cases across models and say which is enough for which task'
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVCJ49YSM62B70ZENBYNB
  - 01M4QKVCQ4B5JC3ZF4RPME94AA
tags:
  - testing
  - tokens
---

## Requirement

The suite tells which model is enough for which task. An extended run executes the same cases on several models (for example haiku, sonnet, opus) with the same variants and N, and a model report answers, per area and per case: which models pass, at what rate, at what cost, and exactly what fails when one does not.

The report must be exact rather than impressionistic. It records the resolved model id from each session (aliases like `haiku` move over time), the harness version, and N. It lists for each model the failing assertions with counts and example runs, separating wrong tool use (for example read `.prep` files directly), wrong outcome (state), and judged quality. It computes cost per passing run (dollars divided by pass rate) so a cheap model that fails half the time is not mistaken for a saving. For each area it names the cheapest model whose pass rate meets a threshold (default 90%, interval lower bound above 80%) and says "not enough data" when N is too small to tell.

The result is a decision table the user can read: area, cheapest sufficient model, its pass rate and cost, the next model up, and what the cheaper ones get wrong.

## Open questions
