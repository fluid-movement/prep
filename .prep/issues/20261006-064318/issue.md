---
title: Tolerate blank lines in decision metadata
kind: code
depends_on:
  - 20261005-152628
---

decisions.md entries keep the format ## <id>: <title> with date:, supersedes: and outcome: lines, but blank lines between the heading and those lines are accepted when reading, and prep fmt removes them. Today a blank line before date: is reported as "decision D1 has no date", which reads as if the date were missing.

## Open questions
