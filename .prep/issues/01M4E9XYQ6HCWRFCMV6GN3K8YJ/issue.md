---
title: 'Knowledge and test gaps: watcher and git entries, gate tests'
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - knowledge
---

Found in the 0.2.0 codebase audit.

- `internal/watch` and `internal/gitx` had no knowledge entry: the watcher's split between project and local changes and its debounce, and how git output is read and why git is optional, were documented nowhere.
- No test exercised the G_REQUIREMENT and G_DOC_ENTRY gates, nor the documentation choice of complete, directly in the domain.

## Open questions
