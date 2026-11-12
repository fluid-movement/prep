---
type: finding
title: Healthy request streams were always labeled partial by the report
description: 'BuildReport had no full-coverage branch: any request forced partial even with zero gaps. Coverage now uses recorded gaps and sessions lacking request usage, not unrelated stream event kinds.'
date: "2026-11-12"
issue: 01M75HAMNFR4SDWSNHX6585DSY
---

The latest live finding recorded three main requests, two subagent requests and no gaps, yet partial coverage. The cause was the unconditional Requests > 0 || Gaps > 0 branch in internal/activity/report.go, not a hidden durable marker or turn summaries. This worktree contained no .prep/local/request-gap-*.json files. The bridge already sends request-interrupted gap events for failed requests in the activity stream, while bridge-disabled markers are separate durable JSON files; current and archived readers include both. Regression tests read a healthy mixed-event stream with the same 3+2 request shape and require full coverage, read and count the bridge-shaped request-interrupted event, and cover missing sessions/usage. A plugin test verifies failed requests actually emit the attributed interruption payload. Legacy request-kind context/cost summaries remain readable without being treated as missing-usage attempts.
