---
title: Remove eval mutate, the CI eval workflow and the knowledge pilot
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQ5YCA96FHQT9WQB6M8C2
tags:
  - eval
priority: low
---

Remove prep eval mutate, .github/workflows/eval.yml and eval/experiments (its findings stay on 01M4TSWY40Z4473D9SPRTT711C). Remove eval cases that test removed features. Evals run locally only. Decided in 01M4ZCMHFA382AG6Q8T9MXFVQT. Delete the code, its tests, fixtures, eval cases, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed feature and GitHub CI on main is green.

## Open questions
