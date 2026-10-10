---
title: 'Project registry: prep remembers the projects it runs in'
kind: code
parent: 01M4JVP89F3PGDHYETHEPS3YBB
tags:
  - desktop
---

Every prep command run in a project records its root in a user-level registry (under `internal/userconfig`), with the time it was last used. `prep projects [--json]` lists them, newest first, and drops entries whose directory or `.prep` no longer exists. The desktop app's project dropdown reads it; the TUI can use it too. Recording never fails a command.

## Open questions
