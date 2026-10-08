---
title: 'Agent view drops ended sessions: session lifecycle events in the activity stream'
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - tui
  - activity
---

The TUI Agent view lists every agent key in the newest 2,000 activity events and never removes one (`activity.Agents`, `internal/activity/activity.go`). Seen on 2026-10-08: the footer said "agent 1/4" while one agent was running; the other three were a finished session (`832549f2…`), prep calls from before `PREP_SESSION` existed (`actor:claude-code/opus-5.5`) and leftover `live-demo-1` events. Subagents are not involved: their tool calls carry the parent's session.

The stream has no notion of a session ending. The Claude Code bridge hooks `session.end` (`plugins/claude-code/hooks/register.ts`) only to flush its queue, and `Event.Validate` accepts only the `tool` and `request` kinds from `prep activity add`.

A second symptom: one agent can show up as two. Session `842b9f9d…` stops at 17:37:02 and `880e3ed8…` starts at 17:37:23 on the same issue, most likely because the session ID changed when the conversation moved to the background (the bridge re-exports `PREP_SESSION` at `turn.start`, but nothing links the old ID to the new one).

Wanted:
- A `session` event kind that harnesses send through `prep activity add`, with `started` and `ended` (and an optional `previous` session when an ID changes), validated like the other harness kinds.
- The Claude Code bridge sends `started` at `session.start`, `ended` at `session.end`, and on an ID change at `turn.start` sends `ended` for the old ID and `started` with `previous` for the new one.
- `Agents` (and whatever the Agent view counts and cycles with tab) leaves out ended sessions and treats a chain of linked IDs as one agent.
- Agents without an end event (crashed sessions, harnesses without lifecycle hooks such as Pi) drop out after a quiet period with no events.
- This repo's activity file is trimmed of the stale `live-demo-1` and actor-keyed lines once; no migration (no outside users).
- The agent-activity, TUI Agent screen and Claude Code integration knowledge entries describe the lifecycle events.

## Open questions

- How long is the quiet period, and is it fixed or configurable?
- Are ended or quiet agents hidden from the Agent view entirely, or kept in a "finished" group that tab skips?
- Does `prep prime`'s "Recently worked on" (Foci) also follow linked session IDs, so a resumed session sees its predecessor's issue as its own?
