---
title: Claude Code panel stays inactive after prep init in the same session
kind: code
tags:
  - integration
---

The Claude Code panel decides once, in its `session.start` hook, whether the session is a prep project: `plugins/claude-code/hooks/register.tsx` runs `prep prime` and sets `active` only when it succeeds. When the session starts before `.prep/` exists and `prep init` runs later in that same session, `active` stays false for the rest of it, so `/prep:pane` keeps answering "Not a prep project." although `prep check` and `prep prime` now work in that directory. Nothing in the panel says why, and the only way out is to restart or resume the session.

Seen in fpa-events on 2026-10-07: the agent ran `prep init` mid-session at the user's request, and `/prep:pane` reported "Not a prep project." from the repo root until the session was resumed.

The panel should notice a prep project that appears during the session — at least when `/prep:pane` is invoked while inactive — and come up without a restart; if it still cannot activate, its message should say what to do instead of only "Not a prep project.".

## Open questions

- Retry only on `/prep:pane`, or also react to a `prep init` run by the agent (the hook already watches the agent's prep commands)?
