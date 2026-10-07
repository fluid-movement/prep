---
title: Remove dead code left by recent changes
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - cli
priority: low
---

`(*Model).editText` in `internal/tui/actions.go` is unused (staticcheck U1000), and `applyAll` in `internal/cli/write.go` still takes a commit message argument that nothing reads since commit mode was removed. Both go, with their callers simplified.

## Open questions
