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
confirmed_commit: ad855be3907fbc5ab7df8b7e06529318670b1a9b
---

# TUI Agent screen

Applies when changing the Agent screen of `prep tui`. Decided in 01M4DTQZPMFCKN0DHDWA95VKTN: harnesses draw no panels of their own; the person runs `prep tui --agent` next to the agent.

- `w` opens or leaves it from any screen, `prep tui --agent` starts on it (`Options.StartAgent`); its place in the screens bar carries the followed agent's live dot ([TUI navigation](/components/tui-navigation.md)).
- **Data**: what an agent does comes from the [activity stream](/features/agent-activity.md) (`Options.Activity`, the newest 2,000 events).
- **Label**: the actor of the agent's latest prep or focus event (the name the agent gives itself; its bridge reports as `claude-code`), else of any event, with the session's first 8 characters.
- **Agent shown**: the pinned one (`a` cycles `activity.Agents`; cycling back to the first follows the most recent again), else the most recently active.
- **Views** (01M4EA4RPQVBHWWZRPATC7XPQC): one full-size pane at a time, chosen in the second tier (`agentS.view`; tab, shift+tab and 1–3 like the issue lists): Issue (the default; the current issue matters more than the tool calls), Activity and Usage. Each view keeps its own scroll (`agentS.offsets`; j/k, g/G, pgup/pgdn and the wheel). The second tier's right end names the agent with `● active <ago>` or `idle for <ago>` after two minutes and `agent n/m` (`agentStatus`).
- **Issue** (`issuePane`): the current issue in full under its ID and title: state, kind and step from `BuildGuide` (cached per tree and issue in `agentS.guide`, since building it validates the whole tree), acceptance progress, the next transition or its first unmet gate (`nowCard`: ack when stale, else the first besides ack, release and drop), up to three knowledge entries the guide points to, then the issue's document without its title (`detailSections`: requirement, context, decisions, records). It scrolls to the top when the agent moves to another issue.
- **Usage**: context meter, tokens, cache, requests and cost from request events (only when a harness sends them; requests count only events with tokens, since a turn's end reports context and cost without them; costs are per-turn differences and add up), This issue (prep commands, criteria checked, knowledge read in the current chapter), read back by area (chars/4) and the three largest knowledge reads.
- **Feed**: newest first, requests left out, grouped under a `ui.Chapter` per issue (the issue current at each event), a `ui.Celebrate` line over a `complete`, issue events drop the target the chapter already names; j/k, g/G, pgup/pgdn and the wheel scroll it.
- **Liveliness**: events newer than the last read are marked fresh for two seconds (one timer); relative times refresh on a 30-second tick that runs only while the screen shows. `enter` opens the current issue's detail. Empty stream: one full-width empty pane.
