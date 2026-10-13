---
schema: 1
id: 01M4N72MS7N69A2QJ356M0P8H2
name: prep
completion_policy:
  verification: Run just ci plus relevant feature tests and record results. Agents execute checks; prep validates and records evidence.
  documentation: Review documentation impact during implementation. Update affected knowledge before completion, or record why no update is needed.
  by: human:azaharias
  at: 2026-10-13T17:04:28Z
---

# prep

Workflow engine and memory for coding agents. Built with prep itself since the first implementation slice.

## Definition of Done

- go test ./... passes
- prep check reports no errors
- prep fmt --check passes
- knowledge entries describing changed behavior are updated
