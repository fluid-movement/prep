---
title: Remove eval mutate, the CI eval workflow and the knowledge pilot
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - eval
priority: low
---

Remove prep eval mutate, .github/workflows/eval.yml and eval/experiments (its findings stay on 01M4TSWY40Z4473D9SPRTT711C). Evals run locally only. Each removal issue deletes the eval cases of the feature it removes, so this issue does not. Decided in 01M4ZCMHFA382AG6Q8T9MXFVQT. Delete the code, its tests, fixtures, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed parts and GitHub CI on main is green.

## Open questions
