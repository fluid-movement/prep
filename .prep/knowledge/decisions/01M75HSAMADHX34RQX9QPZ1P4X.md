---
type: decision
title: Request usage is native, agent-scoped and explicitly partial
description: Claude and Codex request events carry agent identity; missing usage and bridge delivery failures are gaps, while turn summaries stay separate.
date: "2026-11-11"
issue: 01M75HAMNFR4SDWSNHX6585DSY
---

Use Claude turn.step and Codex PostModelRequest only, never transcript totals or a SubagentStop cost hook. Store agent_id, optional agent_type/request_id, turn_id and status on the shared event. Codex input_tokens includes cached_input_tokens, which is split into the shared input and cache_read counters. Claude sends each request promptly; a failed send disables the bridge without retrying and writes one durable request-gap-<session>.json marker under .prep/local through the engine filesystem, independently of the prep binary. Current and archived activity reads include these markers. Reports group native request tokens by session and main/subagent identity, keep cumulative turn cost/context summaries out of per-agent usage, and label coverage partial (or unknown without observations); absence of a marker is not proof that every provider attempt was observed. Old request-kind summaries and events without the new optional fields remain readable.
