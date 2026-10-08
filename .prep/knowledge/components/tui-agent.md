---
type: component
title: TUI Agent screen
description: 'prep tui''s Agent screen: the followed agent''s current issue, its activity feed by issue, its usage; keys, layout, liveliness and data from the activity stream.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-08T16:24:39Z
scope:
  - internal/tui/agent.go
  - internal/tui/agent_test.go
  - internal/tui/ui/activity.go
---

# TUI Agent screen

Applies when changing the Agent screen of `prep tui`. Decided in 01M4DTQZPMFCKN0DHDWA95VKTN: harnesses draw no panels of their own; the person runs `prep tui --agent` next to the agent.

- `w` opens or leaves it from any screen, `prep tui --agent` starts on it (`Options.StartAgent`).
- **Data**: what an agent does comes from the [activity stream](/features/agent-activity.md) (`Options.Activity`, the newest 2,000 events).
- **Label**: the actor of the agent's latest prep or focus event (the name the agent gives itself; its bridge reports as `claude-code`), else of any event, with the session's first 8 characters.
- **Agent shown**: the pinned one (`tab`/`shift+tab` cycle `activity.Agents`; cycling back to the first follows the most recent again), else the most recently active.
- **Layout**: with room (`ui.Split(w, 0.42, 44, 48)`) a Now card and a Usage pane on the left, the Activity feed on the right; narrower, or in one-pane mode (`z`), the Now card (with a one-line usage summary) over the feed.
- **Now**: the agent's current issue (ID, title, state, kind, step from `BuildGuide`, cached per tree and issue in `agentS.guide` since building it validates the whole tree, acceptance progress, the next transition or its first unmet gate: ack when stale, else the first besides ack, release and drop), then `● active <ago>` or `idle for <ago>` after two minutes and `agent n/m`.
- **Usage**: context meter, tokens, cache, requests and cost from request events (only when a harness sends them; requests count only events with tokens, since a turn's end reports context and cost without them; costs are per-turn differences and add up), This issue (prep commands, criteria checked, knowledge read in the current chapter), read back by area (chars/4) and the three largest knowledge reads; short panes keep the top.
- **Feed**: newest first, requests left out, grouped under a `ui.Chapter` per issue (the issue current at each event), a `ui.Celebrate` line over a `complete`, issue events drop the target the chapter already names; j/k, g/G, pgup/pgdn and the wheel scroll it.
- **Liveliness**: events newer than the last read are marked fresh for two seconds (one timer); relative times refresh on a 30-second tick that runs only while the screen shows. `enter` opens the current issue's detail. Empty stream: one full-width empty pane.
