---
title: Every skill rule needs a case that fails without it
kind: code
parent: 01M4QY2WM6XN0JSSTHR5TGQFJT
tags:
  - testing
  - tokens
---

The skill (`internal/cli/skill.md`) has 11 numbered rules, each paid for in tokens by every session that loads it. The mutation tool (`prep eval mutate`, operators `drop-rule`, `first-sentence`, `move-rule`) can tell which rules can be dropped or shortened only if some case actually depends on the rule; for a rule no case exercises, "no effect" means nothing. Map each rule to the cases that exercise it, add cases for the uncovered ones, and check that removing the rule makes its cases fail.

## Open questions
