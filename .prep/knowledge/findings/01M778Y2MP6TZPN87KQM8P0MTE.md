---
type: finding
title: 'Live check: plugin bridge loads headless, agent attribution and gap marker'
description: 'The TypeScript module loads in claude -p: --plugin-dir plugins/claude-code (init shows prep@inline), PATH with /tmp/bin/prep -> /tmp/prep-live. Run /tmp/live3 (stream /tmp/hookcap/stream3.jsonl): request events in .prep/local/activity.jsonl: main requests agent_id absent, subagent requests agent_id afea855681fc4534a with its own turn_id; agent_type absent (engine event gives no agentType); request_id absent; 3 summary events, session started/ended. activity report lists ''subagent afea855681fc4534a: 2 requests'' separate from main (3). Forced failure: run /tmp/live4 with PATH shim /tmp/badbin/prep failing ''activity'' calls: bridge disabled, wrote .prep/local/request-gap-d0d8b40d-....json {kind:gap, op:bridge-disabled}, activity report says ''coverage: partial (1 gap markers...)'', ''request events in 0 of 2 sessions'', gaps 1. Not exercised: per-request gap on a failed request (request-interrupted) and agent_type. Note: report showed ''partial'' with 0 gaps in live3 (likely counts sessions/shared with earlier data), unverified.'
date: "2026-11-12"
issue: 01M75HAMNFR4SDWSNHX6585DSY
---


