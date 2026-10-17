---
type: decision
title: "Prune the tree: delete everything prep does not use"
description: "Prune the tree: delete everything prep does not use"
date: "2026-10-16"
issue: 01M4ZCK8C17RDGBVCTMQPGMEK7
---

User direction, 2026-10-16. Prune the tree: delete all code and features prep is not using, not only those named in the area decisions. Anything without real use and without a decision to keep it goes, including issues that plan such features. Step 5 starts with removals and ends with a sweep for whatever is left unused: dead code found by tooling, commands and options with no use in the transcripts, records and tests that only serve removed features.
