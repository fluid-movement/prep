---
title: Agent judge for fuzzy criteria
kind: code
parent: 01M4QKTQVR3J6PJZAJM0MVWQR0
depends_on:
  - 01M4QKVBMEVANE3EBQM2V2VKR8
  - 01M4QKVC87EPF7DT9AN3NTGF6B
tags:
  - testing
---

## Requirement

An agent judge scores what the deterministic assertions cannot: whether an issue's text is good, whether the agent followed the guidance in spirit, whether its final message is accurate about what it did. A case gives rubric items under `expect.rubric`; the judge reads the transcript, the state diff and the final message and returns, per item, pass or fail with a short reason that quotes the evidence.

The judge is kept small and repeatable: a pinned model and prompt, structured output only, no tools, and each item judged on its own. A judged result never overrides a deterministic failure, and reports show deterministic and judged results separately. The judge's own agreement is measured: the same item judged several times, and a calibration set of hand-labelled runs the judge must reproduce before its verdicts are trusted.

## Open questions
