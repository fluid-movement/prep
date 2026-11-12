---
type: finding
title: 'Real-world run: hand-off by brief works when delegation is requested, not by default'
description: 'Real claude -p sessions (haiku, real plugin hooks and skill), scratch repos /tmp/liveB, fixture order-emails-ready (streams /tmp/liveB/*.jsonl). (1) Task "implement issue <id>", N=3 (r1-r3): 0/3 delegated; all implemented inline (small change), ran prep claim/complete themselves; no brief, no handback. (2) The case prompt asking to delegate, N=3 (d1-d3): 3/3 ran prep brief <id> --out first, 3/3 started one Agent whose prompt was exactly the one-line "Read .prep/local/handoff/<id>/brief.md and follow it exactly. It tells you where to write your report.", 3/3 subagents made no prep call and read no .prep file except brief.md, 3/3 implemented send(orderID), ran go test ./..., wrote report.md in the handoff dir, 3/3 orchestrator ran prep handback <id> successfully (report renamed report.done.md), issue left in-progress, no complete or criterion check. Brief before subagent before handback in all. Extras: orchestrators also ran prep show/knowledge find before the brief (fine); one subagent got model haiku. Conclusion: criterion 5 holds when the task asks for delegation; a bare implement request on a small issue is done inline by haiku.'
date: "2026-11-12"
issue: 01M74V9A3E6Y9N3KMS9QGHCY1J
---


