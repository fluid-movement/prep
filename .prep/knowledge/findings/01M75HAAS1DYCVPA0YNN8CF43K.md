---
type: finding
title: Stop/SubagentStop inputs verified against the reference; cost comes from native request events
description: 'Hook inputs carry transcript paths per the current reference; route chosen: native request events with agent and request IDs, Codex PostModelRequest, no transcript parsing'
date: "2026-11-11"
issue: 01M74VB2DVCENFEXXT7E1CX54A
---

## Answer

Completes finding 01M75GQ1HEVNX6MZ6M9SYZE4HQ (measurement, module losses, Codex), which stays valid; this adds the verified hook reference and the route.

## Stop and SubagentStop inputs (current reference, https://code.claude.com/docs/en/hooks, read 2026-11-11)

- Common fields on every hook: session_id, prompt_id, transcript_path (the main session transcript), cwd, scratchpad_dir, permission_mode, effort, hook_event_name; inside a subagent also agent_id and agent_type.
- Stop adds stop_hook_active, last_assistant_message, background_tasks, session_crons. It carries transcript_path only; the reference tells hooks to use last_assistant_message because transcript writes are asynchronous and lag.
- SubagentStop adds stop_hook_active, agent_id, agent_type, agent_transcript_path (the subagent's own file under <session>/subagents/agent-<id>.jsonl), last_assistant_message. transcript_path stays the main one.
- Codex (reference, per the earlier finding): Stop has nullable transcript_path, SubagentStop adds nullable agent_transcript_path and agent_id; the docs call the transcript format unstable. PostModelRequest carries agent_id, turn_id, status and nullable token_usage per attempt, retries included.

Real-session check: the activity stream and the session transcripts record only that prep hook finish ran, not the hook stdin, and prep does not log raw hook input (hookEvent in internal/cli/hook.go reads no transcript path). The path fields are therefore verified against the reference and the transcript layout (both paths exist as real main and subagent files in the measured session), not against a captured stdin. No live session was run to capture one.

## Route

Cost denominators should come from the native request events, not from finish-hook transcript parsing:
- The module's per-request events already matched the transcripts 110 of 110 for the recorded session, subagents included.
- A finish hook parsing transcripts would need per-agent message-ID dedup state, handling of lagging async writes, and does not fire on interrupts; it ties prep to an undocumented transcript format that Codex itself calls unstable.
- So: keep the plugin module as the Claude source, drop the idea of a transcript-summing finish hook, and use Codex PostModelRequest as the Codex source. SubagentStop is not needed for cost.
- Gap to close: events lack agent ID and a message/request ID, delivery loss (send failure disables the bridge, batches lost on crash) is not observable, and the report must say partial/unknown rather than zero when coverage is not shown. Dollar cost needs model-aware rates; cumulative session cost deltas must not be attributed to agents.
