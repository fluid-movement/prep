---
type: finding
title: Recorded Claude requests match the stream, including subagents; raw finish-hook inputs remain unverified
description: 'Daypack on 2026-11-11: 73/73 main and 37/37 subagent requests match stream usage; the module loses agent attribution, not these requests. Historical loss and real Stop/SubagentStop stdin remain unverified.'
date: "2026-11-11"
issue: 01M74VB2DVCENFEXXT7E1CX54A
---

## Measurement

Read-only measurement on 2026-11-11; no live eval or harness run. Source checkout: 0fbe80f. Claude Code version in the recorded session: 2.1.322.

Session: daypack `3a49dced-99b1-4b8b-aa4d-f38b6c6fd44e`, 2026-11-11 13:50:58–14:33:56 UTC. Sources: `/Users/andre/.claude/projects/-Users-andre-coding-tooling-daypack/<session>.jsonl`, its `<session>/subagents/agent-*.jsonl`, and `/Users/andre/coding/tooling/daypack/.prep/local/activity.jsonl`. The available daypack local directory contains only that current stream, no archives. These are not the historical two-day inputs of the earlier finding; its reported 125-event gap cannot be reproduced from this retained stream.

For each transcript file, keep the last assistant message with usage for each message.id (multiple assistant rows describe one request). Match its four token counts against a multiset of token-bearing stream request events with the same session. Decrement matches so a stream event cannot match twice. Counts are not based on raw assistant rows or on all request-kind events.

| Source | Unique transcript requests | Assistant rows | Matched stream token events | Coverage |
| --- | ---: | ---: | ---: | ---: |
| Main session | 73 | 134 | 73 | 100% |
| Three subagents | 37 | 59 | 37 | 100% |
| Total | 110 | 193 | 110 | 100% |

Subagent request counts: `a436f9cce554902cc` 5, `ab300af48446fa6fe` 5, `ae834edb5251f2a7a` 27. No token-bearing stream event remains unmatched. The 128 request-kind events comprise 110 token records plus 18 context/cost summaries, NOT 128 model requests. Stream totals exactly equal transcript totals: input 256, output 58,361, cache-read 7,854,009, cache-write 243,301. All main requests use claude-sonnet-5-5; subagents use 10 sonnet and 27 opus requests. A second recorded session, `0eee90c4-852c-4e35-851a-d9d2b7016128`, also matches all 43 main requests to 43 token events (plus 8 summary events); it has no subagent transcript files.

This measures recorded requests carrying usage, not invisible failed attempts, billing adjustments, or every request a provider might charge for. Token-vector matching is a measurement aid, not a production deduplication key.

## What the module does and loses

`plugins/claude-code/hooks/register.ts` records result.usage after each successful `turn.step`, and a separate context/cost delta at `turn.complete`. The current event reference explicitly says turn.step and turn.complete fire for subagents too, with e.agentId. The measurement confirms that behavior. The earlier premise that the module runs only for main requests is false for this recorded version.

The module omits e.agentId, so subagent requests are present but indistinguishable by agent in the stream. It does not retain request/message identity either. Requests without result.usage, failed or interrupted generators that never return usage, sessions without the module, and sends after the bridge disables itself are not covered by this mechanism. send clears the queue before the process succeeds; any send failure disables the bridge, without retry, until session.start re-enables it. Batches wait for turn.complete/session.end; a crash can lose pending or unawaited batches. These are source-derived loss mechanisms, NOT demonstrated causes of the old two-day discrepancy.

turn.start calls joinSession even without session.start, so a session ID change is handled; however it resets lastCost and does not re-enable a previously disabled bridge. Resetting a cumulative cost baseline may also misstate cost on resume; the recorded token equality does not validate dollar deltas. Background shell tasks consume no model tokens themselves. Background subagents use the same documented agent request events, but these recorded subagents were not separately established to be background tasks; no background-specific coverage percentage is claimed.

## Finish hooks: reference versus recorded evidence

Current references consulted 2026-11-11:
- https://code.claude.com/docs/en/hooks.md (Common input fields, Stop, SubagentStop)
- https://code.claude.com/docs/en/plugins/mods/events.md (turn events)
- https://developers.openai.com/codex/hooks (Common input fields, model-request hooks, SubagentStop, Stop)

Claude Code documents transcript_path on Stop. SubagentStop includes the main transcript_path plus agent_transcript_path and agent_id for the child. A matcher must include all agent types to avoid silently excluding internal agents. Both transcript paths correspond to real main/child files in the measured session. Its recorded Stop summary at 13:52:47.893 UTC shows prep hook finish executed successfully, alongside two other Stop commands. However, the transcript records hook command/outcome, NOT the hook's stdin; no raw Stop/SubagentStop payload was available in these inspected records, and per-session debug logs were absent. There is no SubagentStop registration in prep's hooks.json today. Therefore the actual path fields in real hook inputs have NOT been verified. File existence and a successful Stop command are not proof of its stdin schema.

The reference warns that transcript writes are asynchronous and the final message may be missing at Stop. Stop also does not fire for user interrupts (API errors use StopFailure). A finish parser cannot guarantee one complete cost event per turn just by reading once: it would need persistent per-agent message-ID deduplication, turn attribution, delayed reconciliation, and interruption/error coverage. Summing the whole file on every Stop would double-count earlier turns. SubagentStop would have to read agent_transcript_path, never mistakenly sum the parent again. Neither input directly promises complete request usage.

## Codex

The current Codex reference offers PostModelRequest for each attempt, including retries: model, turn_id, optional agent_id/agent_type, status (completed/failed/cancelled), error, and nullable token_usage. This is a better prospective source than finish-time parsing. Stop has nullable transcript_path; SubagentStop adds nullable agent_transcript_path and agent_id. The docs explicitly call transcript format unstable. The repository installs neither PostModelRequest nor SubagentStop today, so these shared command hooks collect no Codex request usage. This was a reference/source comparison, not a measured Codex session; no empirical Codex coverage is claimed.

## Recommendation and remaining blocker

Prefer native request usage for report denominators: retain the Claude module, add agent/request attribution and observable delivery coverage; use Codex PostModelRequest rather than add a second transcript parser. Treat turn.complete as context/cost summary, not another request, and never add transcript totals to native totals. Dollar cost needs model-aware rates (the prior Sonnet-relative weights are only an estimate), and cumulative session cost deltas must not be allocated to individual agents without validation. Missing usage must be reported as unknown/partial, never zero or a claim of complete prep cost. Existing historical partial streams cannot establish a complete denominator.

This is a recommendation, not a completed route decision or a filed implementation issue. Verification of actual recorded Stop/SubagentStop stdin, as required by the research acceptance, remains blocked by unavailable raw payloads. No live session was launched to manufacture that evidence. Resolve that evidence gap before declaring this research ready for check or claiming a finish-hook replacement is proven.
