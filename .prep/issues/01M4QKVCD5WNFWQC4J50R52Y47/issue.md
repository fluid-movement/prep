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

An agent scores what cannot be checked mechanically (quality of an issue's text, whether guidance was followed in spirit), given the transcript, the state diff and a rubric from the case. Kept small so its flakiness stays contained; pinned model and prompt.

## Open questions
