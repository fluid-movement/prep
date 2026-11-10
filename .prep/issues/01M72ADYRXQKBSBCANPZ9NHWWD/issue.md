---
title: Before-tool hook says once when code is edited in a session that has no issue
kind: code
parent: 01M5FS7KX63E8KZARSXXAS44C4
depends_on:
  - 01M72ACGWREA9E1TFNGTJB0A98
tags:
  - hooks
priority: low
---

When an agent makes its first edit to a code file in a session in which no issue has been read, created or changed, the before-tool hook adds one line: no prep issue is in use; if this is implementation, find or create one first, unless the user opted out of tracking. It is said once per session and does not block the edit.

Why: "track implementation in an issue" ([decision](/decisions/01M553MC6A94WGF0S1M2KM5MSA.md)) is enforced only by text sent at session start and with the first prompt, and the agent has to judge on every prompt whether it applies. The first code edit is the one moment where "implementation has started" is a fact, not a judgment. In the sessions analysed on 2026-11-10 (finding 01M72A9TBJ5JQXT2VGT15N2HW1), 22-41% of the turns that changed code or delegated work ran no prep command, and in the two clear untracked fixes in fpa-events the agent had searched, found no issue and went on without creating one.

If this works, the prompt hook's reminder can shrink to the scope rules it carries, since the search-or-create instruction is then given at the moment it applies. The prompt hook cannot do this job itself: matching the prompt text against issue titles found the issue the agent then opened in under 10% of cases.

This builds on the edit-time decision notice (01M72ACGWREA9E1TFNGTJB0A98), which adds the hook output path and the per-session memory.

## Open questions

- Is the evidence enough? Only two clear untracked in-repo fixes were found; most turns without prep commands were small requests ("commit and push this") or edits outside the repository. Measuring first with the activity report (see the sibling issue) may be the better order.
- How does the hook know the session has an issue? Claude Code prep commands are not tied to a session in the activity stream today (the same gap as in the activity report issue). Focus is per project, not per session.
- Should a session whose user opted out of tracking be told at all? The hook cannot see the opt-out; the line would have to say "unless the user opted out".
