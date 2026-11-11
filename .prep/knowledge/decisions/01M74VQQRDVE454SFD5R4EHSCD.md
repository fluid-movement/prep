---
type: decision
title: Hook failures, injected sizes and subagent ids are recorded on activity events
description: A hook that fails open after the project is opened records a hook event with verb error; hook events that add context carry its size in chars; tool and hook events carry the Claude Code agent_id.
date: "2026-11-11"
issue: 01M72AV3TQPT1H6GQ6RY46R17T
scope:
  - internal/cli/hook*.go
---

Extends "Hook outcomes and failed commands are activity events" (01M5722JD0K8915A9ZB214Z0B9) for the activity report. Failures: after `a.open()` succeeds, a failed step (user config, turn snapshot, turn state, commit diff, knowledge) records a `hook` event with verb `error` and `<step>: <error>` as target; the hook still prints nothing more and exits 0. Before the project is open there is nowhere to write, so bad input and a missing project stay silent, and `hooks.prompt: off` still skips input and project lookup. A project without a git repository records `turn snapshot` on every prompt and session start: the finish check cannot work there, which is what the event says. Sizes: `chars` on `primed` (session start, newly recorded), `sent`, `asked` and the before-tool denials is the length in characters of the text the hook handed to the model. Agent id: the event field is `agent_id` (`Event.AgentID`; `Agent()` is already the focus key). It comes from the hook payload, which Claude Code fills inside subagents; Codex tool hooks carry no agent id (only SubagentStart/Stop and model-request hooks do), so Codex events have none. Events of prep commands run by a subagent do not carry it: the binary sees only its environment, and Claude Code exports nothing there that names the subagent.
