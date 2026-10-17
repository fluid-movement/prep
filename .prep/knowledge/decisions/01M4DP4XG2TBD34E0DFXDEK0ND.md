---
type: decision
title: "Briefing as a system prompt section, refreshed on session events"
description: "Briefing as a system prompt section, refreshed on session events"
date: "2026-10-08"
issue: 01M46ARM9G3492V298TM82KD45
---

A `before_agent_start` handler sets `systemPromptOptions.sections.prep` to the `prep prime --hook` output. Pi records sections in the session's first system message and appends a delta only when the text changes, so an unchanged briefing costs nothing per turn and is never compacted away. It is recomputed at `session_start` (startup, resume, new, fork, reload) and after `session_compact`, the moments Claude Code's SessionStart hook fires, and stays fixed between them; the agent learns the rest from its own prep calls.
