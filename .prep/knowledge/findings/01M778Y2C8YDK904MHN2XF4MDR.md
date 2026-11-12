---
type: finding
title: 'Live check: PreToolUse decision context reaches model once per session'
description: 'Headless claude -p (haiku, 2.1.322), scratch repo /tmp/livescratch, project hooks via wrapper to /tmp/prep-live. Transcript /Users/andre/.claude/projects/-private-tmp-livescratch/a9023bc0-539b-4657-8635-6b1008fa408e.jsonl holds an attachment hook_additional_context (PreToolUse, toolu_..nykLEe, first Edit of src/cart.go) containing ''Decisions scoped to src/cart.go'' naming decision 01M778TK3Z6AC0DGFDV8BYTPCY. Prompt never mentioned the rule; model said ''The hook points to a recorded decision'' and renamed the param totalCents (applied the Cents rule). Second Edit of src/cart.go in the same session got no further attachment (activity.jsonl: exactly 1 hook context before-tool event, 377 chars). Also: the preceding Read of cart.go got no context, and the subagent''s Read (agent_id a3fc8adb7d0e1b706) got none either (only Reads, no edit). Stream: /tmp/hookcap/stream1.jsonl; stream-json has no hook attachment event without --include-hook-events, the transcript shows it. Passed.'
date: "2026-11-12"
issue: 01M72ACGWREA9E1TFNGTJB0A98
---


