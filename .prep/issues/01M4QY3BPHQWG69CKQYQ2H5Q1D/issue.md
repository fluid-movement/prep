---
title: prep knowledge confirm refuses entries without a scope with a message agents cannot act on
kind: code
parent: 01M498VQD4Q85CZYK965WW3H9Q
tags:
  - knowledge
  - cli
priority: low
---

Asked to mark a knowledge entry as reviewed, an agent ran `prep knowledge confirm /components/payments.md` and got `has no scope; confirmed_commit only means something for scoped entries`. The sentence explains the model, not what to do. Either allow the confirmation to succeed (recording that the entry was reviewed) or make the error say how to proceed (give the entry a scope with `prep knowledge update <entry> --scope <path>`, or that unscoped entries cannot drift and need no confirmation).

## Open questions
