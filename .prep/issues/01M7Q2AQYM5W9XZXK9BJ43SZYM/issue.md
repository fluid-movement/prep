---
title: Fail clearly when the repository schema is newer than the binary
kind: code
parent: 01M7Q2A5H8TN2QGQD97YFVX5FA
tags:
  - cloud
---

Loading a project whose schema is newer than the binary already fails with `project uses schema N; this prep supports M — update prep` (internal/domain/ports.go). What is missing is what a person or agent sees in the cases that matter for cloud sessions, where an old pinned binary meets a repository that moved on.

1. The session-start hook (`prep hook session-start`) must not stay silent when the project's schema is newer than the binary: it prints one line naming both schema numbers and the fix, so the agent and the user learn at once that the briefing is missing and why. It still exits 0 and never blocks the session.
2. The error and the hook line say how to update in both settings: `prep update` locally, and in a cloud session move the `PREP_VERSION` pin in the environment setup script (print the current script with `prep cloud script`).
3. Other commands keep failing with the same error and make no writes.

The schema number is the only compatibility contract; this issue adds no separate minimum-version field to the project and `prep check` keeps its existing schema error. Why: the schema changes exactly when stored records change shape; a second version field would duplicate it and could disagree with it. A repository that needs a newer binary for a feature without a schema change is a release-notes matter, not a check.

## Open questions
