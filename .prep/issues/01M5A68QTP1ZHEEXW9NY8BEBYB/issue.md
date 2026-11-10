---
title: Hooks deny every Bash call once the working directory is inside .prep
kind: code
tags:
  - bug
priority: high
---

When an agent's shell working directory is inside .prep (for example after cd .prep/issues), the before-tool hook denies every Bash call, including cd .., pwd and git status, so the agent cannot leave and the person must move it out by hand.

Cause: hookRecordArea (internal/cli/hook.go) joins a relative or empty path to the event cwd. An empty target (a command with no path arguments) resolves to the cwd itself, which counts as the .prep issues area, so the redirect fires for any command. Reproduced by feeding prep hook before-tool a Bash event with cwd .prep/issues.

## Open questions
