---
type: decision
title: Track implementation work in issues
description: Search before implementing, reuse a clear match, file unmatched work of any size, honor opt-outs, keep the request's scope
date: "2026-10-17"
scope:
  - internal/cli/hook*.go
  - internal/cli/skill.md
  - internal/cli/similar.go
  - eval/cases/natural/**
---

User decision, 2026-10-13 (the untracked-work decision issue); carried over from the descriptive knowledge entry in the 2026-10-17 conversion.

In a prep project, search for existing issues before implementing. Reuse a clear match and follow its guide; if no issue covers authorized implementation, create one first, whatever the size: a one-line behavior fix is still implementation, and guessing whether work is substantial is not a reliable rule. Honor explicit tracking opt-outs and no-record-write restrictions; ordinary questions, exploration and planning need no issue. When an issue is named, use it rather than file another; when several plausibly match, ask which one before changing any.

The original request defines scope for the whole turn: updating a requirement does not authorize implementing or completing it, and an automatic hook continuation grants no new permission.

Similar-issue hints from prep new are advisory: inspect the candidates, never pick or merge one automatically, and do not read their absence as proof that no issue exists.
