---
type: decision
title: Cost comes from native request events, not transcript parsing
description: Keep the plugin module and Codex PostModelRequest as request sources, add agent attribution and gap markers (issue 01M75HAMNFR4SDWSNHX6585DSY); no transcript-summing finish hook
date: "2026-11-11"
issue: 01M74VB2DVCENFEXXT7E1CX54A
---

Decision: the activity report takes token use from native request events. Measured: stream matched transcripts 110/110 requests incl. subagents (finding 01M75GQ1HEVNX6MZ6M9SYZE4HQ). Rejected: Stop/SubagentStop transcript parsing; the hooks give transcript_path and agent_transcript_path (finding 01M75HAAS1DYCVPA0YNN8CF43K), but writes are async, Stop does not fire on interrupts, whole-file sums double count, and the transcript format is undocumented (Codex calls it unstable). Follow-up: code issue 01M75HAMNFR4SDWSNHX6585DSY adds agent IDs and gap markers. Missing coverage is reported as partial, never zero.
